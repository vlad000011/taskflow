package service

import "testing"

func TestValidTaskStatusAllValues(t *testing.T) {
	valid := []string{"backlog", "todo", "in_progress", "review", "done"}
	for _, value := range valid {
		if !ValidTaskStatus(value) {
			t.Errorf("ValidTaskStatus(%q) = false, want true", value)
		}
	}
	invalid := []string{"", "finished", "TODO", "in progress", "blocked"}
	for _, value := range invalid {
		if ValidTaskStatus(value) {
			t.Errorf("ValidTaskStatus(%q) = true, want false", value)
		}
	}
}

func TestValidPriorityAllValues(t *testing.T) {
	valid := []string{"low", "medium", "high", "critical"}
	for _, value := range valid {
		if !ValidPriority(value) {
			t.Errorf("ValidPriority(%q) = false, want true", value)
		}
	}
	invalid := []string{"", "urgent", "LOW", "normal"}
	for _, value := range invalid {
		if ValidPriority(value) {
			t.Errorf("ValidPriority(%q) = true, want false", value)
		}
	}
}

func TestValidRoleAllValues(t *testing.T) {
	valid := []string{"admin", "participant", "lead", "teacher"}
	for _, value := range valid {
		if !ValidRole(value) {
			t.Errorf("ValidRole(%q) = false, want true", value)
		}
	}
	invalid := []string{"", "student", "ADMIN", "owner"}
	for _, value := range invalid {
		if ValidRole(value) {
			t.Errorf("ValidRole(%q) = true, want false", value)
		}
	}
}
