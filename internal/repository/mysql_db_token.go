package repository

import (
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
	"todo/internal/errs"
	"todo/internal/types"
)

type MysqlTokenRepository struct {
	db *gorm.DB
}

func NewMysqlTokenRepository(db *gorm.DB) *MysqlTokenRepository {
	return &MysqlTokenRepository{db: db}
}

func (M *MysqlTokenRepository) Create(token *types.Token) error {
	m := &Token{
		Token:     token.Token,
		UserId:    token.UserId,
		ExpiresAt: time.Unix(token.ExpiresAt, 0),
	}
	if err := M.db.Create(m).Error; err != nil {
		logx.Error(err)
		return errs.ErrorService
	}
	token.Id = m.Id
	token.CreatedAt = m.CreatedAt.Unix()
	return nil
}

func (M *MysqlTokenRepository) FindByToken(token string) (*types.Token, error) {
	var temp Token
	err := M.db.Where("token = ?", token).First(&temp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrorInvalidToken
	}
	if err != nil {
		logx.Error(err)
		return nil, errs.ErrorService
	}
	return &types.Token{
		Id:        temp.Id,
		Token:     temp.Token,
		UserId:    temp.UserId,
		ExpiresAt: temp.ExpiresAt.Unix(),
		CreatedAt: temp.CreatedAt.Unix(),
	}, nil
}

func (M *MysqlTokenRepository) DeleteByToken(token string) error {
	var temp Token
	err := M.db.Where("token = ?", token).First(&temp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrorInvalidToken
	}
	if err != nil {
		logx.Error(err)
		return errs.ErrorService
	}
	err = M.db.Delete(&temp).Error
	if err != nil {
		logx.Error(err)
		return errs.ErrorService
	}
	return nil
}
