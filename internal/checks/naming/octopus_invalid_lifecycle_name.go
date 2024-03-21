package naming

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

const OctoLintInvalidLifecycleNames = "OctoLintInvalidLifecycleNames"

// OctopusInvalidLifecycleName find targets that have not been healthy in the last 30 days.
type OctopusInvalidLifecycleName struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusInvalidLifecycleName(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusInvalidLifecycleName {
	return OctopusInvalidLifecycleName{
		client:       client,
		errorHandler: errorHandler,
		config:       config,
	}
}

func (o OctopusInvalidLifecycleName) Id() string {
	return OctoLintInvalidLifecycleNames
}

func (o OctopusInvalidLifecycleName) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	if o.config.Verbose {
		zap.L().Info("Starting check " + o.Id())
	}

	if strings.TrimSpace(o.config.LifecycleNameRegex) == "" {
		return nil, nil
	}

	regex, err := regexp.Compile(o.config.LifecycleNameRegex)

	if err != nil {
		return checks.NewOctopusCheckResultImpl(
			"The supplied regex "+o.config.LifecycleNameRegex+" does not compile",
			o.Id(),
			"",
			checks.Error,
			checks.Naming), nil
	}

	lifecycles, err := o.client.Lifecycles.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Naming, err)
	}

	responses := []string{}
	for _, l := range lifecycles {
		if !regex.Match([]byte(l.Name)) {
			responses = append(responses, l.Name)
		}
	}

	if len(responses) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following lifecycle names do not match the regex "+o.config.LifecycleNameRegex+":\n"+strings.Join(responses, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Naming), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"All lifecycles match the regex "+o.config.LifecycleNameRegex,
		o.Id(),
		"",
		checks.Ok,
		checks.Naming), nil
}
