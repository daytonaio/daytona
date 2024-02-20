package workspace_dto

type CreateWorkspaceDTO struct {
	Name         string
	Repositories []string
	Provisioner  string
}

type WorkspaceCreationDTO struct {
	Event   string
	Payload string
}

type WorkspaceDTO struct {
	Id string
}

type WorkspacePortForwardDTO struct {
	ContainerPort uint32
	HostPort      uint32
}

type StartWorkspaceDTO struct {
	Id          string
	ProjectName string
}

type StopWorkspaceDTO struct {
	Id          string
	ProjectName string
}

type RemoveWorkspaceDTO struct {
	Id string
}
