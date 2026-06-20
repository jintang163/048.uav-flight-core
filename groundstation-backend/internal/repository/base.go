package repository

import (
	"groundstation-backend/internal/config"
	"groundstation-backend/pkg/utils"

	"gorm.io/gorm"
)

type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository() *BaseRepository {
	return &BaseRepository{
		db: config.DB,
	}
}

func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepository) Create(model interface{}) error {
	return r.db.Create(model).Error
}

func (r *BaseRepository) Update(model interface{}) error {
	return r.db.Save(model).Error
}

func (r *BaseRepository) Delete(model interface{}, id uint64) error {
	return r.db.Delete(model, id).Error
}

func (r *BaseRepository) SoftDelete(model interface{}, id uint64) error {
	return r.db.Delete(model, id).Error
}

func (r *BaseRepository) FindByID(model interface{}, id uint64) error {
	return r.db.First(model, id).Error
}

func (r *BaseRepository) FindByUUID(model interface{}, uuid string) error {
	return r.db.Where("uuid = ?", uuid).First(model).Error
}

func (r *BaseRepository) FindAll(models interface{}) error {
	return r.db.Find(models).Error
}

func (r *BaseRepository) FindWithPage(models interface{}, pagination *utils.Pagination, where interface{}, args ...interface{}) (int64, error) {
	var total int64
	query := r.db.Model(models)
	if where != nil {
		query = query.Where(where, args...)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).Find(models).Error
	return total, err
}

func (r *BaseRepository) Exists(model interface{}, where interface{}, args ...interface{}) (bool, error) {
	var count int64
	err := r.db.Model(model).Where(where, args...).Count(&count).Error
	return count > 0, err
}

func (r *BaseRepository) Count(model interface{}, where interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := r.db.Model(model).Where(where, args...).Count(&count).Error
	return count, err
}
