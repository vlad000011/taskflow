package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"taskflow/backend/internal/config"
	"taskflow/backend/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// These tests exercise the HTTP handlers against a real PostgreSQL database.
// Use a dedicated disposable database via TASKFLOW_TEST_DATABASE_URL.
func openRBACIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TASKFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TASKFLOW_TEST_DATABASE_URL to a dedicated PostgreSQL test database to run handler integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Project{}, &model.ProjectMember{}, &model.Task{},
		&model.Comment{}, &model.Sprint{}, &model.SprintTask{}, &model.TaskHistory{},
		&model.RefreshToken{}, &model.PasswordResetToken{},
	); err != nil {
		t.Fatalf("migrate PostgreSQL test database: %v", err)
	}
	return db
}

type rbacFixture struct {
	db      *gorm.DB
	project model.Project
	users   map[string]model.User
}

func newRBACFixture(t *testing.T, db *gorm.DB) *rbacFixture {
	t.Helper()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	f := &rbacFixture{db: db, users: make(map[string]model.User)}
	for _, roleName := range []string{"admin", "lead", "participant", "teacher", "outsider"} {
		u := model.User{
			Email:        roleName + "-" + suffix + "@rbac.test",
			Name:         "RBAC " + roleName,
			PasswordHash: "test-only-not-a-real-password-hash",
			Role:         roleName,
		}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("create %s fixture user: %v", roleName, err)
		}
		f.users[roleName] = u
	}
	f.project = model.Project{Name: "RBAC integration " + suffix, Description: "test fixture", OwnerID: f.users["lead"].ID}
	if err := db.Create(&f.project).Error; err != nil {
		t.Fatalf("create fixture project: %v", err)
	}
	for _, roleName := range []string{"lead", "participant", "teacher"} {
		m := model.ProjectMember{ProjectID: f.project.ID, UserID: f.users[roleName].ID, Role: roleName}
		if err := db.Create(&m).Error; err != nil {
			t.Fatalf("create %s project membership: %v", roleName, err)
		}
	}
	t.Cleanup(func() {
		var tasks []model.Task
		db.Where("project_id = ?", f.project.ID).Find(&tasks)
		ids := make([]uint, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		if len(ids) > 0 {
			db.Where("task_id IN ?", ids).Delete(&model.Comment{})
			db.Where("task_id IN ?", ids).Delete(&model.TaskHistory{})
			db.Where("task_id IN ?", ids).Delete(&model.SprintTask{})
		}
		db.Where("project_id = ?", f.project.ID).Delete(&model.Task{})
		db.Where("project_id = ?", f.project.ID).Delete(&model.Sprint{})
		userIDs := make([]uint, 0, len(f.users))
		for _, u := range f.users {
			userIDs = append(userIDs, u.ID)
		}
		var ownedProjects []model.Project
		db.Where("owner_id IN ?", userIDs).Find(&ownedProjects)
		projectIDs := make([]uint, 0, len(ownedProjects))
		for _, p := range ownedProjects {
			projectIDs = append(projectIDs, p.ID)
		}
		if len(projectIDs) > 0 {
			db.Where("project_id IN ?", projectIDs).Delete(&model.ProjectMember{})
			db.Where("id IN ?", projectIDs).Delete(&model.Project{})
		}
		for _, u := range f.users {
			db.Delete(&u)
		}
	})
	return f
}

func (f *rbacFixture) task(t *testing.T, assignee string) model.Task {
	t.Helper()
	task := model.Task{
		ProjectID: f.project.ID,
		Title:     "RBAC task",
		Status:    "todo",
		Priority:  "medium",
		CreatorID: f.users["lead"].ID,
	}
	if assignee != "" {
		id := f.users[assignee].ID
		task.AssigneeID = &id
	}
	if err := f.db.Create(&task).Error; err != nil {
		t.Fatalf("create fixture task: %v", err)
	}
	return task
}

func rbacRequest(t *testing.T, h *Handler, method, path, body string, user model.User) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Set("role", user.Role)
		c.Next()
	})
	r.PATCH("/tasks/:id", h.UpdateTask)
	r.POST("/tasks/:id/comments", h.AddComment)
	r.POST("/projects", h.CreateProject)
	r.POST("/projects/:id/sprints", h.CreateSprint)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestRBACHTTPTaskUpdateMatrix(t *testing.T) {
	db := openRBACIntegrationDB(t)
	f := newRBACFixture(t, db)
	h := New(db, config.Config{})

	tests := []struct {
		name, globalRole, assignee, body string
		want                             int
	}{
		{"participant_changes_assigned_task_status", "participant", "participant", `{"status":"in_progress"}`, http.StatusOK},
		{"participant_cannot_change_assigned_task_description", "participant", "participant", `{"description":"forbidden"}`, http.StatusForbidden},
		{"participant_cannot_change_task_assigned_to_other", "participant", "lead", `{"status":"done"}`, http.StatusForbidden},
		{"teacher_is_read_only", "teacher", "participant", `{"status":"done"}`, http.StatusForbidden},
		{"lead_can_edit_project_task", "lead", "participant", `{"description":"lead edit"}`, http.StatusOK},
		{"admin_can_edit_project_task", "admin", "participant", `{"description":"admin edit"}`, http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := f.task(t, tc.assignee)
			w := rbacRequest(t, h, http.MethodPatch, fmt.Sprintf("/tasks/%d", task.ID), tc.body, f.users[tc.globalRole])
			if w.Code != tc.want {
				t.Fatalf("status=%d, want %d; body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestRBACHTTPCommentAndSprintMatrix(t *testing.T) {
	db := openRBACIntegrationDB(t)
	f := newRBACFixture(t, db)
	h := New(db, config.Config{})
	task := f.task(t, "participant")

	for _, tc := range []struct {
		name, role, method, path, body string
		want                           int
	}{
		{"participant_can_comment", "participant", http.MethodPost, fmt.Sprintf("/tasks/%d/comments", task.ID), `{"body":"working on it"}`, http.StatusOK},
		{"teacher_cannot_comment", "teacher", http.MethodPost, fmt.Sprintf("/tasks/%d/comments", task.ID), `{"body":"read only"}`, http.StatusForbidden},
		{"lead_can_create_sprint", "lead", http.MethodPost, fmt.Sprintf("/projects/%d/sprints", f.project.ID), `{"name":"Sprint test","start_date":"2026-11-01T00:00:00Z","end_date":"2026-11-14T00:00:00Z"}`, http.StatusOK},
		{"participant_cannot_create_sprint", "participant", http.MethodPost, fmt.Sprintf("/projects/%d/sprints", f.project.ID), `{"name":"Forbidden sprint","start_date":"2026-11-01T00:00:00Z","end_date":"2026-11-14T00:00:00Z"}`, http.StatusForbidden},
		{"teacher_cannot_create_sprint", "teacher", http.MethodPost, fmt.Sprintf("/projects/%d/sprints", f.project.ID), `{"name":"Forbidden sprint","start_date":"2026-11-01T00:00:00Z","end_date":"2026-11-14T00:00:00Z"}`, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := rbacRequest(t, h, tc.method, tc.path, tc.body, f.users[tc.role])
			if w.Code != tc.want {
				t.Fatalf("status=%d, want %d; body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestRBACHTTPProjectCreationMatrix(t *testing.T) {
	db := openRBACIntegrationDB(t)
	f := newRBACFixture(t, db)
	h := New(db, config.Config{})
	for _, tc := range []struct {
		role string
		want int
	}{
		{"admin", http.StatusOK},
		{"lead", http.StatusOK},
		{"participant", http.StatusForbidden},
		{"teacher", http.StatusForbidden},
	} {
		t.Run(tc.role, func(t *testing.T) {
			w := rbacRequest(t, h, http.MethodPost, "/projects", `{"name":"New RBAC project"}`, f.users[tc.role])
			if w.Code != tc.want {
				t.Fatalf("status=%d, want %d; body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}
