package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	*BaseRepository
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *UserRepository) Create(user *models.User) error {
	user.UUID = utils.GenerateUUID()
	salt := utils.GenerateSalt()
	hash, err := utils.HashPassword(user.PasswordHash, salt)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.PasswordSalt = salt
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint64) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUUID(uuid string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(pagination *utils.Pagination, role models.UserRole, status models.UserStatus) ([]models.User, int64, error) {
	var users []models.User
	query := r.db.Model(&models.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) UpdatePassword(id uint64, newPassword string) error {
	salt := utils.GenerateSalt()
	hash, err := utils.HashPassword(newPassword, salt)
	if err != nil {
		return err
	}
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash": hash,
		"password_salt": salt,
		"token_version": gorm.Expr("token_version + 1"),
	}).Error
}

func (r *UserRepository) UpdateStatus(id uint64, status models.UserStatus) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *UserRepository) UpdateRole(id uint64, role models.UserRole) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *UserRepository) UpdateLoginInfo(id uint64, ip string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ip,
		"login_attempts": 0,
	}).Error
}

func (r *UserRepository) IncrementLoginAttempts(username string) error {
	return r.db.Model(&models.User{}).Where("username = ?", username).Updates(map[string]interface{}{
		"login_attempts": gorm.Expr("login_attempts + 1"),
		"last_failed_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *UserRepository) UpdateRefreshToken(id uint64, refreshToken string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("refresh_token", refreshToken).Error
}

func (r *UserRepository) IncrementTokenVersion(id uint64) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("token_version", gorm.Expr("token_version + 1")).Error
}

func (r *UserRepository) CreateLoginLog(log *models.LoginLog) error {
	return r.db.Create(log).Error
}

func (r *UserRepository) GetLoginLogs(userID uint64, pagination *utils.Pagination) ([]models.LoginLog, int64, error) {
	var logs []models.LoginLog
	query := r.db.Model(&models.LoginLog{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *UserRepository) GetUserPermissions(userID uint64) ([]models.UserPermission, error) {
	var permissions []models.UserPermission
	err := r.db.Where("user_id = ?", userID).Find(&permissions).Error
	return permissions, err
}

func (r *UserRepository) LockUser(id uint64) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("status", models.UserStatusLocked).Error
}

func (r *UserRepository) UnlockUser(id uint64) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         models.UserStatusActive,
		"login_attempts": 0,
	}).Error
}

func (r *UserRepository) CheckUsernameExists(username string, excludeID uint64) (bool, error) {
	query := r.db.Model(&models.User{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) CheckEmailExists(email string, excludeID uint64) (bool, error) {
	query := r.db.Model(&models.User{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
