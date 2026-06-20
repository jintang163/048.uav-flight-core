package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin    UserRole = "ADMIN"
	RoleOperator UserRole = "OPERATOR"
	RoleUser     UserRole = "USER"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

type User struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID            string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Username        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email           string         `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Phone           string         `gorm:"type:varchar(20)" json:"phone"`
	PasswordHash    string         `gorm:"type:varchar(255);not null" json:"-"`
	PasswordSalt    string         `gorm:"type:varchar(64);not null" json:"-"`
	FullName        string         `gorm:"type:varchar(100)" json:"full_name"`
	Role            UserRole       `gorm:"type:varchar(20);default:'USER'" json:"role"`
	Status          UserStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	AvatarURL       string         `gorm:"type:varchar(255)" json:"avatar_url"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	LastLoginIP     string         `gorm:"type:varchar(45)" json:"last_login_ip"`
	LoginAttempts   int            `gorm:"default:0" json:"login_attempts"`
	LastFailedAt    *time.Time     `json:"last_failed_at"`
	TwoFactorEnabled bool          `gorm:"default:false" json:"two_factor_enabled"`
	TokenVersion    int            `gorm:"default:1" json:"token_version"`
	RefreshToken    string         `gorm:"type:varchar(500)" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type UserPermission struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"index;not null" json:"user_id"`
	Resource    string    `gorm:"type:varchar(100);not null" json:"resource"`
	Action      string    `gorm:"type:varchar(50);not null" json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

func (UserPermission) TableName() string {
	return "user_permissions"
}

type LoginLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	Username   string    `gorm:"type:varchar(50);not null" json:"username"`
	Success    bool      `json:"success"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  string    `gorm:"type:varchar(500)" json:"user_agent"`
	Location   string    `gorm:"type:varchar(200)" json:"location"`
	Reason     string    `gorm:"type:varchar(200)" json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
