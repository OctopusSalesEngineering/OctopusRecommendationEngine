terraform {
  required_providers {
    octopusdeploy = { source = "OctopusDeployLabs/octopusdeploy", version = "0.11.2" }
  }
}


provider "octopusdeploy" {
  address  = "${var.octopus_server}"
  api_key  = "${var.octopus_apikey}"
  space_id = "${var.octopus_space_id}"
}

variable "octopus_server" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The URL of the Octopus server e.g. https://myinstance.octopus.app."
}
variable "octopus_apikey" {
  type        = string
  nullable    = false
  sensitive   = true
  description = "The API key used to access the Octopus server. See https://octopus.com/docs/octopus-rest-api/how-to-create-an-api-key for details on creating an API key."
}
variable "octopus_space_id" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The space ID to populate"
}

resource "octopusdeploy_user_role" "octolintrole" {
  can_be_deleted                = true
  description                   = "A read only role used by the octolint service account."
  granted_space_permissions     = [
    "AccountView", "ActionTemplateView", "ArtifactView", "CertificateView", "DeploymentView", "EnvironmentView",
    "EventView", "FeedView", "GitCredentialView", "InsightsReportView", "InterruptionView", "LibraryVariableSetView",
    "LifecycleView", "MachinePolicyView", "MachineView", "ProcessView", "projectGroupView", "ProjectView", "ProxyView",
    "ReleaseView", "RunbookRunView", "RunbookView", "SubscriptionView", "TaskView", "TeamView", "TenantView",
    "TriggerView", "VariableView", "VariableViewUnscoped", "WorkerView"
  ]
  granted_system_permissions    = []
  name                          = "Octolint"
  space_permission_descriptions = []
}

resource "octopusdeploy_user" "octolint" {
  display_name  = "Octolint"
  email_address = "a@example.org"
  is_active     = true
  is_service    = true
  username      = "octolint"
}

resource "octopusdeploy_team" "octolint" {
  name = "Octolint"
  users = [octopusdeploy_user.octolint.id]
  user_role {
    space_id     = var.octopus_space_id
    user_role_id = octopusdeploy_user_role.octolintrole.id
  }
}