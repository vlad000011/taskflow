package service

import "testing"

func uintPtr(v uint) *uint { return &v }

func TestCanCreateProject(t *testing.T) {
	for _, tc := range []struct {
		role string
		want bool
	}{
		{"admin", true}, {"lead", true}, {"participant", false}, {"teacher", false}, {"", false},
	} {
		t.Run(tc.role, func(t *testing.T) {
			if got := CanCreateProject(tc.role); got != tc.want {
				t.Fatalf("CanCreateProject(%q)=%v, want %v", tc.role, got, tc.want)
			}
		})
	}
}

func TestCanReadProject(t *testing.T) {
	for _, tc := range []struct {
		global, member string
		want           bool
	}{
		{"admin", "", true}, {"participant", "lead", true}, {"participant", "participant", true},
		{"participant", "teacher", true}, {"participant", "", false}, {"teacher", "", false},
	} {
		if got := CanReadProject(tc.global, tc.member); got != tc.want {
			t.Errorf("CanReadProject(%q,%q)=%v, want %v", tc.global, tc.member, got, tc.want)
		}
	}
}

func TestCanManageProject(t *testing.T) {
	for _, tc := range []struct {
		global, member string
		want           bool
	}{
		{"admin", "", true}, {"participant", "lead", true}, {"participant", "admin", true},
		{"participant", "participant", false}, {"participant", "teacher", false}, {"teacher", "", false},
	} {
		if got := CanManageProject(tc.global, tc.member); got != tc.want {
			t.Errorf("CanManageProject(%q,%q)=%v, want %v", tc.global, tc.member, got, tc.want)
		}
	}
}

func TestCanDeleteProject(t *testing.T) {
	for _, tc := range []struct {
		role        string
		owner, user uint
		want        bool
	}{
		{"admin", 10, 20, true}, {"lead", 10, 10, true}, {"lead", 10, 20, false},
		{"participant", 10, 10, false}, {"teacher", 10, 10, false},
	} {
		if got := CanDeleteProject(tc.role, tc.owner, tc.user); got != tc.want {
			t.Errorf("CanDeleteProject(%q,%d,%d)=%v, want %v", tc.role, tc.owner, tc.user, got, tc.want)
		}
	}
}

func TestCanUpdateTask(t *testing.T) {
	assignee := uint(7)
	for _, tc := range []struct {
		name, global, member string
		assignee             *uint
		user                 uint
		statusOnly, want     bool
	}{
		{"admin-any-task", "admin", "", nil, 99, false, true},
		{"lead-project-task", "participant", "lead", nil, 99, false, true},
		{"participant-assigned-status", "participant", "participant", &assignee, 7, true, true},
		{"participant-assigned-description", "participant", "participant", &assignee, 7, false, false},
		{"participant-unassigned", "participant", "participant", nil, 7, true, false},
		{"participant-other-assignee", "participant", "participant", &assignee, 8, true, false},
		{"teacher-read-only", "teacher", "teacher", &assignee, 7, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanUpdateTask(tc.global, tc.member, tc.assignee, tc.user, tc.statusOnly); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanComment(t *testing.T) {
	for _, tc := range []struct {
		global, member string
		want           bool
	}{
		{"admin", "", true}, {"participant", "lead", true}, {"participant", "participant", true},
		{"teacher", "teacher", false}, {"participant", "", false},
	} {
		if got := CanComment(tc.global, tc.member); got != tc.want {
			t.Errorf("CanComment(%q,%q)=%v, want %v", tc.global, tc.member, got, tc.want)
		}
	}
}

func TestCanUpload(t *testing.T) {
	for _, tc := range []struct {
		global, member string
		assignee       *uint
		user           uint
		want           bool
	}{
		{"admin", "", nil, 1, true}, {"participant", "lead", nil, 1, true},
		{"participant", "participant", uintPtr(5), 5, true},
		{"participant", "participant", uintPtr(5), 6, false},
		{"participant", "participant", nil, 5, false}, {"teacher", "teacher", uintPtr(5), 5, false},
	} {
		if got := CanUpload(tc.global, tc.member, tc.assignee, tc.user); got != tc.want {
			t.Errorf("CanUpload(%q,%q)=%v, want %v", tc.global, tc.member, got, tc.want)
		}
	}
}
