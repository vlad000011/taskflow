package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Name         string    `gorm:"size:120;not null" json:"name"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"size:24;not null;default:participant;index" json:"role"`
	IsBlocked    bool      `gorm:"not null;default:false" json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:160;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     uint      `gorm:"not null;index" json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"uniqueIndex:idx_project_user;not null;index" json:"project_id"`
	UserID    uint      `gorm:"uniqueIndex:idx_project_user;not null;index" json:"user_id"`
	Role      string    `gorm:"size:24;not null;default:participant" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
type Task struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ProjectID      uint       `gorm:"not null;index" json:"project_id"`
	SprintID       *uint      `gorm:"index" json:"sprint_id"`
	Title          string     `gorm:"size:200;not null;index" json:"title"`
	Description    string     `gorm:"type:text" json:"description"`
	Status         string     `gorm:"size:24;not null;default:todo;index" json:"status"`
	Priority       string     `gorm:"size:24;not null;default:medium;index" json:"priority"`
	AssigneeID     *uint      `gorm:"index" json:"assignee_id"`
	CreatorID      uint       `gorm:"not null;index" json:"creator_id"`
	DueDate        *time.Time `json:"due_date"`
	AttachmentPath string     `gorm:"size:500" json:"attachment_path,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"not null;index" json:"task_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
type Sprint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	Name      string    `gorm:"size:160;not null" json:"name"`
	StartDate time.Time `gorm:"not null" json:"start_date"`
	EndDate   time.Time `gorm:"not null" json:"end_date"`
	IsClosed  bool      `gorm:"not null;default:false" json:"is_closed"`
	CreatedAt time.Time `json:"created_at"`
}
type SprintTask struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	SprintID uint `gorm:"uniqueIndex:idx_sprint_task;not null;index" json:"sprint_id"`
	TaskID   uint `gorm:"uniqueIndex:idx_sprint_task;not null;index" json:"task_id"`
}
type TaskHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"not null;index" json:"task_id"`
	ActorID   uint      `gorm:"not null;index" json:"actor_id"`
	Field     string    `gorm:"size:60;not null" json:"field"`
	OldValue  string    `gorm:"type:text" json:"old_value"`
	NewValue  string    `gorm:"type:text" json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName maps TaskHistory to the singular table created by the SQL migration.
func (TaskHistory) TableName() string {
	return "task_history"
}

type RefreshToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	TokenHash string     `gorm:"uniqueIndex;size:64;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
}
type PasswordResetToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	TokenHash string     `gorm:"uniqueIndex;size:64;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
