package repository

import "todo/internal/types"

type TodoRepository interface {
	Create(userID int, todo *types.Todo) error
	FindById(id int, userID int) (*types.Todo, error)
	FindAll(userID int) ([]*types.Todo, error)
	Update(userID int, todo *types.Todo) error
	Delete(id int, userID int) error
}
