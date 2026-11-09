package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/httpx"
	"taskflow/backend/internal/model"
	"taskflow/backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB  *gorm.DB
	Cfg config.Config
}

func New(db *gorm.DB, cfg config.Config) *Handler { return &Handler{DB: db, Cfg: cfg} }
func uid(c *gin.Context) uint                     { v, _ := c.Get("userID"); return v.(uint) }
func fail(c *gin.Context, code, msg string)       { httpx.Error(c, 400, code, msg) }
func role(c *gin.Context) string                  { v, _ := c.Get("role"); r, _ := v.(string); return r }
func isAdmin(c *gin.Context) bool                 { return role(c) == "admin" }
func (h *Handler) projectRole(c *gin.Context, projectID string) string {
	if isAdmin(c) {
		return "admin"
	}
	var m model.ProjectMember
	if h.DB.Where("project_id = ? AND user_id = ?", projectID, uid(c)).First(&m).Error != nil {
		return ""
	}
	return m.Role
}
func (h *Handler) requireProjectRole(c *gin.Context, projectID string, allowed ...string) bool {
	r := h.projectRole(c, projectID)
	for _, a := range allowed {
		if r == a {
			return true
		}
	}
	httpx.Error(c, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав для операции в проекте")
	return false
}

func (h *Handler) Register(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil || !strings.Contains(in.Email, "@") || len(in.Password) < 8 || strings.TrimSpace(in.Name) == "" {
		fail(c, "VALIDATION_ERROR", "Нужны имя, корректный email и пароль минимум из 8 символов")
		return
	}
	u := model.User{Email: strings.ToLower(strings.TrimSpace(in.Email)), Name: strings.TrimSpace(in.Name), PasswordHash: service.HashPassword(in.Password), Role: "participant"}
	if err := h.DB.Create(&u).Error; err != nil {
		httpx.Error(c, 409, "EMAIL_EXISTS", "Email уже используется")
		return
	}
	httpx.OK(c, gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role})
}
func (h *Handler) Login(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, "VALIDATION_ERROR", "Некорректные данные")
		return
	}
	var u model.User
	if h.DB.Where("email = ?", strings.ToLower(in.Email)).First(&u).Error != nil || !service.CheckPassword(u.PasswordHash, in.Password) || u.IsBlocked {
		httpx.Error(c, 401, "LOGIN_FAILED", "Неверные учётные данные или аккаунт заблокирован")
		return
	}
	access, err := service.MakeJWT(h.Cfg.JWTSecret, u.ID, u.Role, "access", time.Now().Add(time.Duration(h.Cfg.AccessMinutes)*time.Minute))
	if err != nil {
		httpx.Error(c, 500, "TOKEN_ERROR", "Не удалось создать токен")
		return
	}
	refresh, err := service.MakeJWT(h.Cfg.JWTSecret, u.ID, u.Role, "refresh", time.Now().Add(time.Duration(h.Cfg.RefreshDays)*24*time.Hour))
	if err != nil {
		httpx.Error(c, 500, "TOKEN_ERROR", "Не удалось создать токен")
		return
	}
	h.DB.Create(&model.RefreshToken{UserID: u.ID, TokenHash: service.TokenHash(refresh), ExpiresAt: time.Now().Add(time.Duration(h.Cfg.RefreshDays) * 24 * time.Hour)})
	httpx.OK(c, gin.H{"access_token": access, "refresh_token": refresh, "user": gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role}})
}
func (h *Handler) Refresh(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, "VALIDATION_ERROR", "refresh_token обязателен")
		return
	}
	id, role, err := service.ParseJWT(h.Cfg.JWTSecret, in.RefreshToken, "refresh")
	if err != nil {
		httpx.Error(c, 401, "INVALID_TOKEN", "Недействительный refresh-токен")
		return
	}
	var row model.RefreshToken
	if h.DB.Where("token_hash = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", service.TokenHash(in.RefreshToken), id, time.Now()).First(&row).Error != nil {
		httpx.Error(c, 401, "INVALID_TOKEN", "Токен отозван или истёк")
		return
	}
	now := time.Now()
	h.DB.Model(&row).Update("revoked_at", &now)
	access, _ := service.MakeJWT(h.Cfg.JWTSecret, id, role, "access", now.Add(time.Duration(h.Cfg.AccessMinutes)*time.Minute))
	refresh, _ := service.MakeJWT(h.Cfg.JWTSecret, id, role, "refresh", now.Add(time.Duration(h.Cfg.RefreshDays)*24*time.Hour))
	h.DB.Create(&model.RefreshToken{UserID: id, TokenHash: service.TokenHash(refresh), ExpiresAt: now.Add(time.Duration(h.Cfg.RefreshDays) * 24 * time.Hour)})
	httpx.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}
func (h *Handler) Logout(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&in)
	if in.RefreshToken != "" {
		now := time.Now()
		h.DB.Model(&model.RefreshToken{}).Where("user_id = ? AND token_hash = ?", uid(c), service.TokenHash(in.RefreshToken)).Update("revoked_at", &now)
	}
	c.JSON(200, gin.H{"data": gin.H{"logged_out": true}})
}
func (h *Handler) Me(c *gin.Context) {
	var u model.User
	if h.DB.First(&u, uid(c)).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Пользователь не найден")
		return
	}
	httpx.OK(c, u)
}
func (h *Handler) UpdateMe(c *gin.Context) {
	var in struct {
		Name string `json:"name"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" {
		fail(c, "VALIDATION_ERROR", "Имя обязательно")
		return
	}
	h.DB.Model(&model.User{}).Where("id = ?", uid(c)).Update("name", strings.TrimSpace(in.Name))
	h.Me(c)
}
func (h *Handler) ChangePassword(c *gin.Context) {
	var in struct {
		Current string `json:"current_password"`
		New     string `json:"new_password"`
	}
	if c.ShouldBindJSON(&in) != nil || len(in.New) < 8 {
		fail(c, "VALIDATION_ERROR", "Новый пароль должен содержать минимум 8 символов")
		return
	}
	var u model.User
	h.DB.First(&u, uid(c))
	if !service.CheckPassword(u.PasswordHash, in.Current) {
		httpx.Error(c, 400, "INVALID_PASSWORD", "Текущий пароль неверен")
		return
	}
	h.DB.Model(&u).Update("password_hash", service.HashPassword(in.New))
	c.JSON(200, gin.H{"data": gin.H{"changed": true}})
}
func (h *Handler) ForgotPassword(c *gin.Context) {
	var in struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, "VALIDATION_ERROR", "Email обязателен")
		return
	}
	var u model.User
	if h.DB.Where("email = ?", strings.ToLower(in.Email)).First(&u).Error == nil {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		token := hex.EncodeToString(b)
		h.DB.Create(&model.PasswordResetToken{UserID: u.ID, TokenHash: service.TokenHash(token), ExpiresAt: time.Now().Add(time.Hour)})
		fmt.Printf("PASSWORD RESET for %s: http://localhost:8080/reset-password?token=%s\n", u.Email, token)
	}
	c.JSON(200, gin.H{"data": gin.H{"message": "Если аккаунт существует, ссылка восстановления записана в лог сервера"}})
}
func (h *Handler) ResetPassword(c *gin.Context) {
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil || len(in.Password) < 8 {
		fail(c, "VALIDATION_ERROR", "Токен и пароль минимум из 8 символов обязательны")
		return
	}
	var row model.PasswordResetToken
	if h.DB.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", service.TokenHash(in.Token), time.Now()).First(&row).Error != nil {
		httpx.Error(c, 400, "INVALID_RESET_TOKEN", "Токен недействителен или истёк")
		return
	}
	if err := h.DB.Model(&model.User{}).Where("id = ?", row.UserID).Update("password_hash", service.HashPassword(in.Password)).Error; err != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR", "Ошибка обновления пароля")
		return
	}
	now := time.Now()
	h.DB.Model(&row).Update("used_at", &now)
	c.JSON(200, gin.H{"data": gin.H{"changed": true}})
}
func (h *Handler) CreateProject(c *gin.Context) {
	if !service.CanCreateProject(role(c)) {
		httpx.Error(c, 403, "FORBIDDEN", "Создавать проекты может руководитель или администратор")
		return
	}
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" {
		fail(c, "VALIDATION_ERROR", "Название проекта обязательно")
		return
	}
	p := model.Project{Name: strings.TrimSpace(in.Name), Description: in.Description, OwnerID: uid(c)}
	if h.DB.Create(&p).Error != nil {
		httpx.Error(c, 500, "DB_ERROR", "Не удалось создать проект")
		return
	}
	h.DB.Create(&model.ProjectMember{ProjectID: p.ID, UserID: uid(c), Role: "lead"})
	httpx.OK(c, p)
}
func (h *Handler) ListProjects(c *gin.Context) {
	var ps []model.Project
	q := h.DB.Model(&model.Project{})
	if !isAdmin(c) {
		q = q.Joins("JOIN project_members pm ON pm.project_id = projects.id").Where("pm.user_id = ?", uid(c))
	}
	if s := c.Query("search"); s != "" {
		q = q.Where("projects.name ILIKE ?", "%"+s+"%")
	}
	q.Order("projects.created_at DESC").Find(&ps)
	httpx.OK(c, ps)
}
func (h *Handler) Project(c *gin.Context) {
	id := c.Param("id")
	if !h.canAccessProject(c, id) {
		return
	}
	var p model.Project
	if h.DB.First(&p, id).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Проект не найден")
		return
	}
	httpx.OK(c, p)
}
func (h *Handler) canAccessProject(c *gin.Context, projectID string) bool {
	if isAdmin(c) {
		return true
	}
	var n int64
	h.DB.Model(&model.ProjectMember{}).Where("project_id = ? AND user_id = ?", projectID, uid(c)).Count(&n)
	if n == 0 {
		httpx.Error(c, 403, "FORBIDDEN", "Нет доступа к проекту")
		return false
	}
	return true
}
func (h *Handler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	if !h.canAccessProject(c, id) {
		return
	}
	if !h.requireProjectRole(c, id, "lead", "admin") {
		return
	}
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" {
		fail(c, "VALIDATION_ERROR", "Название обязательно")
		return
	}
	h.DB.Model(&model.Project{}).Where("id = ?", id).Updates(map[string]any{"name": strings.TrimSpace(in.Name), "description": in.Description})
	h.Project(c)
}
func (h *Handler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if !h.canAccessProject(c, id) {
		return
	}
	var p model.Project
	if h.DB.First(&p, id).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Проект не найден")
		return
	}
	// A project lead may manage and delete their project even when they
	// were assigned the lead role through project membership rather than
	// being the original creator. Global admins retain access to all projects.
	if !h.requireProjectRole(c, id, "lead", "admin") {
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var tasks []model.Task
		if err := tx.Where("project_id = ?", id).Find(&tasks).Error; err != nil {
			return err
		}
		ids := make([]uint, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		if len(ids) > 0 {
			if err := tx.Where("task_id IN ?", ids).Delete(&model.Comment{}).Error; err != nil {
				return err
			}
			if err := tx.Where("task_id IN ?", ids).Delete(&model.TaskHistory{}).Error; err != nil {
				return err
			}
			if err := tx.Where("task_id IN ?", ids).Delete(&model.SprintTask{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.ProjectMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.Task{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.Sprint{}).Error; err != nil {
			return err
		}
		return tx.Delete(&p).Error
	})
	if err != nil {
		httpx.Error(c, 500, "DB_ERROR", "Не удалось удалить проект")
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"deleted": true}})
}
func (h *Handler) ListTasks(c *gin.Context) {
	pid := c.Param("id")
	if !h.canAccessProject(c, pid) {
		return
	}
	page := parseInt(c.Query("page"), 1)
	limit := parseInt(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	q := h.DB.Model(&model.Task{}).Where("project_id = ?", pid)
	if v := c.Query("status"); v != "" {
		if !service.ValidTaskStatus(v) {
			fail(c, "VALIDATION_ERROR", "Недопустимый статус")
			return
		}
		q = q.Where("status = ?", v)
	}
	if v := c.Query("priority"); v != "" {
		if !service.ValidPriority(v) {
			fail(c, "VALIDATION_ERROR", "Недопустимый приоритет")
			return
		}
		q = q.Where("priority = ?", v)
	}
	if v := c.Query("assignee_id"); v != "" {
		q = q.Where("assignee_id = ?", v)
	}
	if v := c.Query("search"); v != "" {
		q = q.Where("(title ILIKE ? OR description ILIKE ?)", "%"+v+"%", "%"+v+"%")
	}
	sortField := map[string]string{"created_at": "created_at", "updated_at": "updated_at", "due_date": "due_date", "priority": "priority", "title": "title"}[c.Query("sort")]
	if sortField == "" {
		sortField = "created_at"
	}
	order := "DESC"
	if strings.EqualFold(c.Query("order"), "asc") {
		order = "ASC"
	}
	var total int64
	q.Count(&total)
	var tasks []model.Task
	q.Order(sortField + " " + order).Offset((page - 1) * limit).Limit(limit).Find(&tasks)
	httpx.OK(c, gin.H{"items": tasks, "pagination": gin.H{"page": page, "limit": limit, "total": total, "pages": (total + int64(limit) - 1) / int64(limit)}})
}
func (h *Handler) CreateTask(c *gin.Context) {
	pid := c.Param("id")
	if !h.canAccessProject(c, pid) {
		return
	}
	if !h.requireProjectRole(c, pid, "lead", "admin") {
		return
	}
	var in struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Status      string     `json:"status"`
		Priority    string     `json:"priority"`
		AssigneeID  *uint      `json:"assignee_id"`
		DueDate     *time.Time `json:"due_date"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Title) == "" {
		fail(c, "VALIDATION_ERROR", "Название задачи обязательно")
		return
	}
	if in.Status == "" {
		in.Status = "todo"
	}
	if in.Priority == "" {
		in.Priority = "medium"
	}
	if !service.ValidTaskStatus(in.Status) || !service.ValidPriority(in.Priority) {
		fail(c, "VALIDATION_ERROR", "Недопустимый статус или приоритет")
		return
	}
	pID, _ := strconv.ParseUint(pid, 10, 64)
	t := model.Task{ProjectID: uint(pID), Title: strings.TrimSpace(in.Title), Description: in.Description, Status: in.Status, Priority: in.Priority, AssigneeID: in.AssigneeID, CreatorID: uid(c), DueDate: in.DueDate}
	if h.DB.Create(&t).Error != nil {
		httpx.Error(c, 500, "DB_ERROR", "Не удалось создать задачу")
		return
	}
	h.DB.Create(&model.TaskHistory{TaskID: t.ID, ActorID: uid(c), Field: "created", NewValue: t.Status})
	httpx.OK(c, t)
}
func (h *Handler) Task(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	if !h.canAccessProject(c, strconv.FormatUint(uint64(t.ProjectID), 10)) {
		return
	}
	httpx.OK(c, t)
}
func (h *Handler) UpdateTask(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	pid := strconv.FormatUint(uint64(t.ProjectID), 10)
	if !h.canAccessProject(c, pid) {
		return
	}
	pr := h.projectRole(c, pid)
	var in struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		Status      *string    `json:"status"`
		Priority    *string    `json:"priority"`
		AssigneeID  *uint      `json:"assignee_id"`
		DueDate     *time.Time `json:"due_date"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, "VALIDATION_ERROR", "Некорректные данные")
		return
	}
	statusOnly := in.Status != nil && in.Title == nil && in.Description == nil && in.Priority == nil && in.AssigneeID == nil && in.DueDate == nil
	if !service.CanUpdateTask(role(c), pr, t.AssigneeID, uid(c), statusOnly) {
		httpx.Error(c, 403, "FORBIDDEN", "Недостаточно прав для изменения этой задачи")
		return
	}
	updates := map[string]any{}
	if in.Title != nil {
		if strings.TrimSpace(*in.Title) == "" {
			fail(c, "VALIDATION_ERROR", "Название обязательно")
			return
		}
		updates["title"] = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.Status != nil {
		if !service.ValidTaskStatus(*in.Status) {
			fail(c, "VALIDATION_ERROR", "Недопустимый статус")
			return
		}
		if t.Status != *in.Status {
			updates["status"] = *in.Status
			h.DB.Create(&model.TaskHistory{TaskID: t.ID, ActorID: uid(c), Field: "status", OldValue: t.Status, NewValue: *in.Status})
		}
	}
	if in.Priority != nil {
		if !service.ValidPriority(*in.Priority) {
			fail(c, "VALIDATION_ERROR", "Недопустимый приоритет")
			return
		}
		updates["priority"] = *in.Priority
	}
	if in.AssigneeID != nil {
		updates["assignee_id"] = *in.AssigneeID
	}
	if in.DueDate != nil {
		updates["due_date"] = *in.DueDate
	}
	if len(updates) > 0 {
		h.DB.Model(&t).Updates(updates)
	}
	h.Task(c)
}
func (h *Handler) DeleteTask(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	pid := strconv.FormatUint(uint64(t.ProjectID), 10)
	if !h.canAccessProject(c, pid) {
		return
	}
	if !h.requireProjectRole(c, pid, "lead", "admin") {
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", t.ID).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", t.ID).Delete(&model.TaskHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", t.ID).Delete(&model.SprintTask{}).Error; err != nil {
			return err
		}
		return tx.Delete(&t).Error
	})
	if err != nil {
		httpx.Error(c, 500, "DB_ERROR", "Не удалось удалить задачу")
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"deleted": true}})
}
func (h *Handler) AddComment(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	pid := strconv.FormatUint(uint64(t.ProjectID), 10)
	if !h.canAccessProject(c, pid) {
		return
	}
	if !service.CanComment(role(c), h.projectRole(c, pid)) {
		httpx.Error(c, 403, "FORBIDDEN", "Недостаточно прав для комментария")
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Body) == "" {
		fail(c, "VALIDATION_ERROR", "Комментарий обязателен")
		return
	}
	cm := model.Comment{TaskID: t.ID, UserID: uid(c), Body: strings.TrimSpace(in.Body)}
	h.DB.Create(&cm)
	httpx.OK(c, cm)
}
func (h *Handler) ListComments(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	if !h.canAccessProject(c, strconv.FormatUint(uint64(t.ProjectID), 10)) {
		return
	}
	var rows []model.Comment
	h.DB.Where("task_id = ?", t.ID).Order("created_at ASC").Find(&rows)
	httpx.OK(c, rows)
}
func (h *Handler) CreateSprint(c *gin.Context) {
	pid := c.Param("id")
	if !h.canAccessProject(c, pid) {
		return
	}
	if !h.requireProjectRole(c, pid, "lead", "admin") {
		return
	}
	var in struct {
		Name      string    `json:"name"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" || in.EndDate.Before(in.StartDate) {
		fail(c, "VALIDATION_ERROR", "Укажите название и корректный период спринта")
		return
	}
	n, _ := strconv.ParseUint(pid, 10, 64)
	s := model.Sprint{ProjectID: uint(n), Name: strings.TrimSpace(in.Name), StartDate: in.StartDate, EndDate: in.EndDate}
	h.DB.Create(&s)
	httpx.OK(c, s)
}
func (h *Handler) ListSprints(c *gin.Context) {
	pid := c.Param("id")
	if !h.canAccessProject(c, pid) {
		return
	}
	var rows []model.Sprint
	h.DB.Where("project_id = ?", pid).Order("start_date DESC").Find(&rows)
	httpx.OK(c, rows)
}
func (h *Handler) Dashboard(c *gin.Context) {
	pid := c.Param("id")
	if !h.canAccessProject(c, pid) {
		return
	}
	var total, done, overdue int64
	h.DB.Model(&model.Task{}).Where("project_id = ?", pid).Count(&total)
	h.DB.Model(&model.Task{}).Where("project_id = ? AND status = ?", pid, "done").Count(&done)
	h.DB.Model(&model.Task{}).Where("project_id = ? AND due_date < ? AND status <> ?", pid, time.Now(), "done").Count(&overdue)
	var byStatus []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	h.DB.Model(&model.Task{}).Select("status, count(*) as count").Where("project_id = ?", pid).Group("status").Scan(&byStatus)
	httpx.OK(c, gin.H{"total_tasks": total, "completed_tasks": done, "overdue_tasks": overdue, "completion_percent": percent(done, total), "by_status": byStatus})
}
func (h *Handler) AdminUsers(c *gin.Context) {
	var rows []model.User
	h.DB.Select("id,email,name,role,is_blocked,created_at").Order("id ASC").Find(&rows)
	httpx.OK(c, rows)
}
func (h *Handler) AdminRole(c *gin.Context) {
	var in struct {
		Role string `json:"role"`
	}
	if c.ShouldBindJSON(&in) != nil || !service.ValidRole(in.Role) {
		fail(c, "VALIDATION_ERROR", "Недопустимая роль")
		return
	}
	if h.DB.Model(&model.User{}).Where("id = ?", c.Param("id")).Update("role", in.Role).RowsAffected == 0 {
		httpx.Error(c, 404, "NOT_FOUND", "Пользователь не найден")
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"updated": true}})
}
func (h *Handler) AdminBlock(c *gin.Context) {
	var in struct {
		Blocked bool `json:"blocked"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, "VALIDATION_ERROR", "Поле blocked обязательно")
		return
	}
	if h.DB.Model(&model.User{}).Where("id = ?", c.Param("id")).Update("is_blocked", in.Blocked).RowsAffected == 0 {
		httpx.Error(c, 404, "NOT_FOUND", "Пользователь не найден")
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"updated": true}})
}
func (h *Handler) Upload(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	pid := strconv.FormatUint(uint64(t.ProjectID), 10)
	if !h.canAccessProject(c, pid) {
		return
	}
	pr := h.projectRole(c, pid)
	if !service.CanUpload(role(c), pr, t.AssigneeID, uid(c)) {
		httpx.Error(c, 403, "FORBIDDEN", "Нет права прикреплять файл к этой задаче")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(h.Cfg.MaxUploadMB)*1024*1024)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httpx.Error(c, 400, "UPLOAD_ERROR", "Файл не передан или превышен лимит размера")
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".pdf": true, ".txt": true, ".docx": true}
	if !allowed[ext] {
		httpx.Error(c, 400, "INVALID_FILE_TYPE", "Разрешены PNG, JPG, PDF, TXT и DOCX")
		return
	}
	if err = os.MkdirAll(h.Cfg.UploadDir, 0750); err != nil {
		httpx.Error(c, 500, "UPLOAD_ERROR", "Не удалось сохранить файл")
		return
	}
	name := fmt.Sprintf("task-%d-%d%s", t.ID, time.Now().UnixNano(), ext)
	out, err := os.OpenFile(filepath.Join(h.Cfg.UploadDir, name), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0640)
	if err != nil {
		httpx.Error(c, 500, "UPLOAD_ERROR", "Не удалось сохранить файл")
		return
	}
	_, copyErr := io.Copy(out, file)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		httpx.Error(c, 500, "UPLOAD_ERROR", "Ошибка записи файла")
		return
	}
	h.DB.Model(&t).Update("attachment_path", "/uploads/"+name)
	httpx.OK(c, gin.H{"path": "/uploads/" + name, "filename": header.Filename})
}
func (h *Handler) History(c *gin.Context) {
	var t model.Task
	if h.DB.First(&t, c.Param("id")).Error != nil {
		httpx.Error(c, 404, "NOT_FOUND", "Задача не найдена")
		return
	}
	if !h.canAccessProject(c, strconv.FormatUint(uint64(t.ProjectID), 10)) {
		return
	}
	var rows []model.TaskHistory
	h.DB.Where("task_id = ?", t.ID).Order("created_at DESC").Find(&rows)
	httpx.OK(c, rows)
}
func parseInt(s string, d int) int {
	n, e := strconv.Atoi(s)
	if e != nil || n < 1 {
		return d
	}
	return n
}
func percent(a, b int64) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) * 100 / float64(b)
}

var _ = errors.Is
