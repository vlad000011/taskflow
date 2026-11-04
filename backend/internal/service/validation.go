package service

func ValidTaskStatus(s string) bool {
	switch s {
	case "backlog", "todo", "in_progress", "review", "done":
		return true
	}
	return false
}
func ValidPriority(s string) bool {
	switch s {
	case "low", "medium", "high", "critical":
		return true
	}
	return false
}
func ValidRole(s string) bool {
	switch s {
	case "admin", "participant", "lead", "teacher":
		return true
	}
	return false
}
