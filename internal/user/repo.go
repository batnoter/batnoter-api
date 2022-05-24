package user

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Repo represents a user repository.
// It provides methods to retrieve and manage user records from the database.
//go:generate mockgen -source=repo.go -package=user -destination=mock_repo.go
type Repo interface {
	Get(userID uint) (User, error)
	GetByEmail(email string) (User, error)
	Save(user User) (uint, error)
	Delete(userID uint) error
}

type repoImpl struct {
	db *gorm.DB
}

// NewRepository creates and returns a new instance of user repository.
func NewRepository(db *gorm.DB) Repo {
	return &repoImpl{
		db: db,
	}
}

// Get returns a user record by user-id.
func (r *repoImpl) Get(userID uint) (User, error) {
	var user User
	if err := r.db.Where("id = ?", userID).Preload("DefaultRepo").First(&user).Error; err != nil {
		return user, errors.Wrap(err, "retrieving user from database failed")
	}
	return user, nil
}

// GetByEmail returns a user record matching the provided email.
func (r *repoImpl) GetByEmail(email string) (User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, nil
	}
	if err != nil {
		return user, errors.Wrap(err, "retrieving user from database failed")
	}
	return user, nil
}

// Save stores a given user record to database.
func (r *repoImpl) Save(user User) (uint, error) {
	if err := r.db.Save(&user).Error; err != nil {
		return 0, errors.Wrap(err, "storing user to database failed")
	}
	return user.ID, nil
}

// Delete marks a user record matching provided user-id as deleted in database.
func (r *repoImpl) Delete(userID uint) error {
	var user User
	if err := r.db.Where("id = ?", userID).Delete(&user).Error; err != nil {
		return errors.Wrap(err, "deleting user from database failed")
	}
	return nil
}
