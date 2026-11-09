package service

// RBAC policy helpers are deliberately independent of HTTP and persistence,
// so the permission matrix can be tested without a running PostgreSQL instance.
func CanCreateProject(globalRole string) bool {
	return globalRole == "admin" || globalRole == "lead"
}

func CanReadProject(globalRole, projectRole string) bool {
	return globalRole == "admin" || projectRole == "lead" || projectRole == "participant" || projectRole == "teacher"
}

func CanManageProject(globalRole, projectRole string) bool {
	return globalRole == "admin" || projectRole == "admin" || projectRole == "lead"
}

func CanDeleteProject(globalRole string, ownerID, userID uint) bool {
	return globalRole == "admin" || (globalRole == "lead" && ownerID == userID)
}

// Participant may update only the status field of a task assigned to them.
func CanUpdateTask(globalRole, projectRole string, assigneeID *uint, userID uint, statusOnly bool) bool {
	if globalRole == "admin" || projectRole == "admin" || projectRole == "lead" {
		return true
	}
	return projectRole == "participant" && assigneeID != nil && *assigneeID == userID && statusOnly
}

func CanComment(globalRole, projectRole string) bool {
	return globalRole == "admin" || projectRole == "admin" || projectRole == "lead" || projectRole == "participant"
}

func CanUpload(globalRole, projectRole string, assigneeID *uint, userID uint) bool {
	if globalRole == "admin" || projectRole == "admin" || projectRole == "lead" {
		return true
	}
	return projectRole == "participant" && assigneeID != nil && *assigneeID == userID
}
