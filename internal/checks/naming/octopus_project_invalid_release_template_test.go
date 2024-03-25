package naming

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"github.com/OctopusSolutionsEngineering/OctopusTerraformTestFramework/octoclient"
	"github.com/OctopusSolutionsEngineering/OctopusTerraformTestFramework/test"
	"path/filepath"
	"testing"
)

func TestInvalidReleaseTemplate(t *testing.T) {
	testFramework := test.OctopusContainerTest{}
	testFramework.ArrangeTest(t, func(t *testing.T, container *test.OctopusContainer, client *client.Client) error {
		// Act
		newSpaceId, err := testFramework.Act(
			t,
			container,
			filepath.Join("..", "..", "..", "test", "terraform"), "30-defaultstepnames", []string{})

		if err != nil {
			return err
		}

		newSpaceClient, err := octoclient.CreateClient(container.URI, newSpaceId, test.ApiKey)

		if err != nil {
			return err
		}

		check := NewOctopusProjectReleaseTemplateRegex(
			newSpaceClient,
			&config.OctolintConfig{
				ProjectReleaseTemplateRegex: "#\\{Octopus\\.Version\\.LastMajor\\}\\.#\\{Octopus\\.Version\\.LastMinor\\}\\.#\\{Octopus\\.Version\\.LastPatch\\}",
			},
			checks.OctopusClientPermissiveErrorHandler{})

		result, err := check.Execute()

		if err != nil {
			return err
		}

		// Assert
		if result == nil || result.Severity() != checks.Warning {
			return errors.New("check should have produced a warning")
		}

		return nil
	})
}
