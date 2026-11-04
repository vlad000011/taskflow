package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/handler"
	"taskflow/backend/internal/middleware"
	"taskflow/backend/internal/model"
	"taskflow/backend/internal/service"
	"time"
)

func main() {
	cfg := config.Load()
	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET должен содержать минимум 32 символа")
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err = runMigrations(db, "./migrations"); err != nil {
		log.Fatal(err)
	}
	if cfg.SeedDemo {
		seed(db)
	}
	if err = os.MkdirAll(cfg.UploadDir, 0750); err != nil {
		log.Fatal(err)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), cors.New(cors.Config{AllowOrigins: []string{cfg.CORSOrigin}, AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}, AllowCredentials: true, MaxAge: 12 * time.Hour}))
	h := handler.New(db, cfg)
	api := r.Group("/api/v1")
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/refresh", h.Refresh)
	api.POST("/auth/forgot-password", h.ForgotPassword)
	api.POST("/auth/reset-password", h.ResetPassword)
	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	auth := api.Group("")
	auth.Use(middleware.Auth(cfg))
	auth.POST("/auth/logout", h.Logout)
	auth.GET("/users/me", h.Me)
	auth.PATCH("/users/me", h.UpdateMe)
	auth.POST("/users/me/password", h.ChangePassword)
	auth.GET("/projects", h.ListProjects)
	auth.POST("/projects", h.CreateProject)
	auth.GET("/projects/:id", h.Project)
	auth.PATCH("/projects/:id", h.UpdateProject)
	auth.DELETE("/projects/:id", h.DeleteProject)
	auth.GET("/projects/:id/tasks", h.ListTasks)
	auth.POST("/projects/:id/tasks", h.CreateTask)
	auth.GET("/projects/:id/dashboard", h.Dashboard)
	auth.GET("/projects/:id/sprints", h.ListSprints)
	auth.POST("/projects/:id/sprints", h.CreateSprint)
	auth.GET("/tasks/:id", h.Task)
	auth.PATCH("/tasks/:id", h.UpdateTask)
	auth.DELETE("/tasks/:id", h.DeleteTask)
	auth.POST("/tasks/:id/comments", h.AddComment)
	auth.GET("/tasks/:id/comments", h.ListComments)
	auth.GET("/tasks/:id/history", h.History)
	auth.POST("/tasks/:id/attachment", h.Upload)
	admin := auth.Group("/admin")
	admin.Use(middleware.Roles("admin"))
	admin.GET("/users", h.AdminUsers)
	admin.PATCH("/users/:id/role", h.AdminRole)
	admin.PATCH("/users/:id/status", h.AdminBlock)
	r.Static("/uploads", cfg.UploadDir)
	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(`<!doctype html><html><head><title>TaskFlow Swagger UI</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:"/swagger/openapi.yaml",dom_id:"#swagger-ui"})</script></body></html>`))
	})
	r.StaticFile("/swagger/openapi.yaml", "./docs/openapi.yaml")
	fmt.Printf("TaskFlow API listening on :%s\n", cfg.Port)
	if err = r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
func seed(db *gorm.DB) {
	var n int64
	db.Model(&model.User{}).Count(&n)
	if n > 0 {
		return
	}
	users := []model.User{{Email: "admin@taskflow.local", Name: "Demo Admin", PasswordHash: service.HashPassword("ChangeMe123!"), Role: "admin"}, {Email: "lead@taskflow.local", Name: "Demo Lead", PasswordHash: service.HashPassword("ChangeMe123!"), Role: "lead"}, {Email: "student@taskflow.local", Name: "Demo Student", PasswordHash: service.HashPassword("ChangeMe123!"), Role: "participant"}, {Email: "teacher@taskflow.local", Name: "Demo Teacher", PasswordHash: service.HashPassword("ChangeMe123!"), Role: "teacher"}}
	for i := range users {
		db.Create(&users[i])
	}
	p := model.Project{Name: "Учебный проект TaskFlow", Description: "Демо-проект для практики", OwnerID: users[1].ID}
	db.Create(&p)
	for _, u := range users {
		role := u.Role
		if role == "lead" {
			role = "lead"
		}
		db.Create(&model.ProjectMember{ProjectID: p.ID, UserID: u.ID, Role: role})
	}
	now := time.Now()
	tasks := []model.Task{{ProjectID: p.ID, Title: "Спроектировать REST API", Description: "Подготовить OpenAPI", Status: "done", Priority: "high", CreatorID: users[1].ID}, {ProjectID: p.ID, Title: "Сделать канбан-доску", Status: "in_progress", Priority: "critical", AssigneeID: &users[2].ID, CreatorID: users[1].ID, DueDate: &now}, {ProjectID: p.ID, Title: "Добавить тесты", Status: "todo", Priority: "medium", CreatorID: users[1].ID}}
	for _, t := range tasks {
		db.Create(&t)
	}
}
