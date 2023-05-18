package hooklistener

// GitlabWebhookPayload is part of the payload GitLab sends us after a pipeline event happened
type GitlabWebhookPayload struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes struct {
		Ref    string `json:"ref"`
		Status string `json:"status"`
	} `json:"object_attributes"`
}

// IsActionable checks if the webhook is actionable for us, we filter out other hooks we receive and don't need
func (p GitlabWebhookPayload) IsActionable() bool {
	if p.ObjectKind == "pipeline" && p.ObjectAttributes.Ref == "master" && p.ObjectAttributes.Status == "success" {
		return true
	}
	return false
}
