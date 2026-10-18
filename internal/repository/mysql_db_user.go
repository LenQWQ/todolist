package repository

import (
	"errors"
	"todo/internal/errs"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type MysqlUserRepository struct {
	db *gorm.DB
}

func NewMysqlUserRepository(db *gorm.DB) *MysqlUserRepository {
	return &MysqlUserRepository{db: db}
}

func (M *MysqlUserRepository) Create(user *types.User) error {
	m := &User{Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash}
	err := M.db.Create(m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.ErrorUserAlreadyExists
		}
		logx.Error(err)
		return errs.ErrorService
	}
	user.Id = m.Id
	user.CreatedAt = m.CreatedAt.Unix()
	user.UpdatedAt = m.UpdatedAt.Unix()
	return nil
}

func (M *MysqlUserRepository) FindById(id int) (*types.User, error) {
	var user User
	err := M.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrorUserNotFound
	}
	if err != nil {
		logx.Error(err)
		return nil, errs.ErrorService
	}
	return &types.User{Id: user.Id, Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash, CreatedAt: user.CreatedAt.Unix(), UpdatedAt: user.UpdatedAt.Unix()}, nil
}

func (M *MysqlUserRepository) FindByUsername(username string) (*types.User, error) {
	var user User
	err := M.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrorUserNotFound
	}
	if err != nil {
		logx.Error(err)
		return nil, errs.ErrorService
	}
	return &types.User{Id: user.Id, Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash, CreatedAt: user.CreatedAt.Unix(), UpdatedAt: user.UpdatedAt.Unix()}, nil
}

func (M *MysqlUserRepository) FindByEmail(email string) (*types.User, error) {
	var user User
	err := M.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrorUserNotFound
	}
	if err != nil {
		logx.Error(err)
		return nil, errs.ErrorService
	}
	return &types.User{Id: user.Id, Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash, CreatedAt: user.CreatedAt.Unix(), UpdatedAt: user.UpdatedAt.Unix()}, nil
}

func (M *MysqlUserRepository) UpdatePassword(id int, passwordHash string) error {
	res := M.db.Model(&User{}).Where("id = ?", id).Update("password_hash", passwordHash)
	if res.Error != nil {
		logx.Error(res.Error)
		return errs.ErrorService
	}
	return nil
}

func (M *MysqlUserRepository) UpdateEmail(id int, email string) error {
	res := M.db.Model(&User{}).Where("id = ?", id).Update("email", email)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
			return errs.ErrorEmailAlreadyExists
		}
		logx.Error(res.Error)
		return errs.ErrorService
	}
	return nil
}
