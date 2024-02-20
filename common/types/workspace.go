package types

type Repository struct {
	Url      string `json:"url"`
	Branch   string `json:"branch,omitempty"`
	Sha      string `json:"sha"`
	Owner    string `json:"owner"`
	PrNumber uint32 `json:"prNumber,omitempty"`
	Source   string `json:"source"`
	Path     string `json:"path,omitempty"`
}

type Project struct {
	Name        string      `json:"name"`
	Repository  *Repository `json:"repository"`
	WorkspaceId string      `json:"workspaceId"`
	AuthKey     string      `json:"authKey"`
}

type WorkspaceProvisioner struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

type Workspace struct {
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	Projects    []*Project            `json:"projects"`
	Provisioner *WorkspaceProvisioner `json:"provisioner"`
}

type ProjectInfo struct {
	Name                string       `json:"name"`
	Created             string       `json:"created"`
	Started             string       `json:"started"`
	Finished            string       `json:"finished"`
	IsRunning           bool         `json:"isRunning"`
	ProvisionerMetadata *interface{} `json:"provisionerMetadata,omitempty"`
	WorkspaceId         string       `json:"workspaceId"`
}

type WorkspaceInfo struct {
	Name                string         `json:"name"`
	Projects            []*ProjectInfo `json:"projects"`
	ProvisionerMetadata *interface{}   `json:"provisionerMetadata,omitempty"`
}
