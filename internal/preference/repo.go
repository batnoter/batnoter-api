package preference

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Repo represents a preference repository.
// It provides methods retrieve and save the user preferences from database.
//go:generate mockgen -source=repo.go -package=preference -destination=mock_repo.go
type Repo interface {
	Save(defaultRepo DefaultRepo) error
	GetByUserID(userID uint) (DefaultRepo, error)
}

type repoImpl struct {
	db *gorm.DB
}

// NewRepository creates and returns a new instance of preference repository.
func NewRepository(db *gorm.DB) Repo {
	return &repoImpl{
		db: db,
	}
}

// GetByUserID returns user's default repo preference.
func (r *repoImpl) GetByUserID(userID uint) (DefaultRepo, error) {
	var defaultRepo DefaultRepo
	err := r.db.Where("user_id = ?", userID).First(&defaultRepo).Error
	if err == gorm.ErrRecordNotFound {
		return DefaultRepo{}, nil
	}
	if err != nil {
		return defaultRepo, errors.Wrap(err, "retrieving user's default repo from database failed")
	}
	return defaultRepo, nil
}

// Save creates or updates user's default repository preference.
func (r *repoImpl) Save(defaultRepo DefaultRepo) error {
	if err := r.db.Save(&defaultRepo).Error; err != nil {
		return errors.Wrap(err, "storing user's default repo to database failed")
	}
	return nil
}
