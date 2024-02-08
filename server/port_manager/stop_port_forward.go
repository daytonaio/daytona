package port_manager

import log "github.com/sirupsen/logrus"

func StopPortForward(workspaceName string, projectContainerName string, port ContainerPort) error {
	workspacePortForward, ok := workspacePortForwards[workspaceName]
	if !ok {
		log.Debugf("Workspace %s does not have any port forwards", workspaceName)
		return nil
	}

	projectPortForwards, ok := workspacePortForward.ProjectPortForwards[projectContainerName]
	if !ok {
		log.Debugf("Project %s does not have any port forwards", projectContainerName)
		return nil
	}

	portForward, ok := projectPortForwards[port]
	if !ok {
		log.Debugf("Port %d is not forwarded", port)
		return nil
	}

	portForward.cancelFunc()
	delete(projectPortForwards, port)

	if len(projectPortForwards) == 0 {
		delete(workspacePortForward.ProjectPortForwards, projectContainerName)
	}

	if len(workspacePortForward.ProjectPortForwards) == 0 {
		delete(workspacePortForwards, workspaceName)
	}

	return nil
}

func StopAllWorkspaceForwards(workspaceName string) error {
	workspacePortForward, ok := workspacePortForwards[workspaceName]
	if !ok {
		log.Debugf("Workspace %s does not have any port forwards", workspaceName)
		return nil
	}

	for projectContainerName, projectPortForwards := range workspacePortForward.ProjectPortForwards {
		for _, portForward := range projectPortForwards {
			StopPortForward(workspaceName, projectContainerName, portForward.ContainerPort)
		}
	}

	return nil
}

func StopAllWorkspaceProjectForwards(workspaceName string, projectContainerName string) error {
	workspacePortForward, ok := workspacePortForwards[workspaceName]
	if !ok {
		log.Debugf("Workspace %s does not have any port forwards", workspaceName)
		return nil
	}

	projectPortForwards, ok := workspacePortForward.ProjectPortForwards[projectContainerName]
	if !ok {
		log.Debugf("Project %s does not have any port forwards", projectContainerName)
		return nil
	}

	for _, portForward := range projectPortForwards {
		StopPortForward(workspaceName, projectContainerName, portForward.ContainerPort)
	}

	return nil
}
