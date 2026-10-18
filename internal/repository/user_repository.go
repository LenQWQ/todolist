package repository

import "todo/internal/types"

type UserRepository interface {
	Create(user *types.User) error
	FindById(id int) (*types.User, error)
	FindByEmail(email string) (*types.User, error)
	FindByUsername(username string) (*types.User, error)
	UpdateEmail(id int, email string) error
	UpdatePassword(id int, passwordHash string) error
}
