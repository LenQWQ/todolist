package repository

import "todo/internal/types"

type TokenRepository interface {
	Create(user *types.Token) error
	FindByToken(token string) (*types.Token, error)
	DeleteByToken(token string) error
}
