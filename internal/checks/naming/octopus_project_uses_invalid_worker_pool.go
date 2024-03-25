package naming

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/core"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/deployments"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/workerpools"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

const OctoLintProjectWorkerPool = "OctoLintProjectWorkerPool"

// OctopusProjectWorkerPoolRegex checks to see if any project has too many steps.
type OctopusProjectWorkerPoolRegex struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusProjectWorkerPoolRegex(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusProjectWorkerPoolRegex {
	return OctopusProjectWorkerPoolRegex{
		client:       client,
		errorHandler: errorHandler,
		config:       config,
	}
}

func (o OctopusProjectWorkerPoolRegex) Id() string {
	return OctoLintProjectWorkerPool
}

func (o OctopusProjectWorkerPoolRegex) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	if o.config.Verbose {
		zap.L().Info("Starting check " + o.Id())
	}

	defer func() {
		if o.config.Verbose {
			zap.L().Info("Ended check " + o.Id())
		}
	}()

	if strings.TrimSpace(o.config.ProjectStepWorkerPoolRegex) == "" {
		return nil, nil
	}

	regex, err := regexp.Compile(o.config.ProjectStepWorkerPoolRegex)

	if err != nil {

		return checks.NewOctopusCheckResultImpl(
			"The supplied regex "+o.config.ProjectStepWorkerPoolRegex+" does not compile",
			o.Id(),
			"",
			checks.Error,
			checks.Naming), nil
	}

	projects, err := o.client.Projects.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Naming, err)
	}

	workerPools, err := o.client.WorkerPools.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Naming, err)
	}

	defaultWorkerPools := lo.Filter(workerPools, func(item *workerpools.WorkerPoolListResult, index int) bool {
		return item.IsDefault
	})

	defaultWorkerPool := ""
	if len(defaultWorkerPools) == 1 {
		defaultWorkerPool = defaultWorkerPools[0].Name
	}

	actionsWithInvalidWorkerPools := []string{}
	for _, p := range projects {
		deploymentProcess, err := o.stepsInDeploymentProcess(p.DeploymentProcessID)

		if err != nil {
			if !o.errorHandler.ShouldContinue(err) {
				return nil, err
			}
			continue
		}

		if deploymentProcess == nil {
			continue
		}

		for _, s := range deploymentProcess.Steps {
			for _, a := range s.Actions {

				if a.WorkerPoolVariable != "" {
					continue
				}

				if a.WorkerPool == "" {
					if defaultWorkerPool != "" && !regex.Match([]byte(defaultWorkerPool)) {

						actionsWithInvalidWorkerPools = append(actionsWithInvalidWorkerPools, p.Name+"/"+a.Name+": "+defaultWorkerPool+" (default)")
					}
				} else if !regex.Match([]byte(a.WorkerPool)) {
					workerPool := lo.Filter(workerPools, func(item *workerpools.WorkerPoolListResult, index int) bool {
						return item.ID == a.WorkerPool
					})

					if len(workerPool) == 1 && !regex.Match([]byte(workerPool[0].Name)) {
						actionsWithInvalidWorkerPools = append(actionsWithInvalidWorkerPools, p.Name+"/"+a.Name+": "+workerPool[0].Name)
					}
				}
			}
		}

	}

	if len(actionsWithInvalidWorkerPools) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following project actions use worker pools that do not match the regex "+o.config.ContainerImageRegex+":\n"+strings.Join(actionsWithInvalidWorkerPools, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Organization), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"There are no actions that use worker pools that do not match the regex "+o.config.ContainerImageRegex,
		o.Id(),
		"",
		checks.Ok,
		checks.Organization), nil
}

func (o OctopusProjectWorkerPoolRegex) stepsInDeploymentProcess(deploymentProcessID string) (*deployments.DeploymentProcess, error) {
	if deploymentProcessID == "" {
		return nil, nil
	}

	resource, err := o.client.DeploymentProcesses.GetByID(deploymentProcessID)

	if err != nil {
		// If we can't find the deployment process, assume zero steps
		if err.(*core.APIError).StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return resource, nil
}
