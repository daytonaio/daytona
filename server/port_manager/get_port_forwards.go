package port_manager

import log "github.com/sirupsen/logrus"

func GetPortForwards(workspaceName string) (WorkspacePortForward, error) {
	workspacePortForward, ok := workspacePortForwards[workspaceName]
	if !ok {
		return WorkspacePortForward{}, nil
	}

	return workspacePortForward, nil
}

func GetProjectPortForwards(workspaceName string, projectContainerName string) (PortForwards, error) {
	workspacePortForward, ok := workspacePortForwards[workspaceName]
	if !ok {
		log.Debugf("Workspace %s does not have any port forwards", workspaceName)
		return nil, nil
	}

	projectPortForwards, ok := workspacePortForward.ProjectPortForwards[projectContainerName]
	if !ok {
		log.Debugf("Project %s does not have any port forwards", projectContainerName)
		return nil, nil
	}

	return projectPortForwards, nil
}
