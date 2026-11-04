package models

// DeploymentResponse represents the fields needed from the site deployment response.
type DeploymentResponse struct {
	Message string         `json:"message"`
	Data    DeploymentData `json:"data"`
}

type DeploymentData struct {
	UploadedFilePaths []string `json:"uploaded_file_paths"`
	PublishedDomain   string   `json:"published_domain"`
}
