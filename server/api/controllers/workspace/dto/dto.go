package dto

type CreateWorkspace struct {
	Name         string
	Repositories []string
	Provisioner  string
}

type WorkspaceCreation struct {
	Event   string
	Payload string
}

type Workspace struct {
	Id string
}

type WorkspacePortForward struct {
	ContainerPort uint32
	HostPort      uint32
}

type StartWorkspace struct {
	Id          string
	ProjectName string
}

type StopWorkspace struct {
	Id          string
	ProjectName string
}

type RemoveWorkspace struct {
	Id string
}
