package repository

import (
	"errors"
	"todo/internal/errs"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type MysqlTodoRepository struct {
	db *gorm.DB
}

func NewMysqlTodoRepository(db *gorm.DB) *MysqlTodoRepository {
	return &MysqlTodoRepository{db: db}
}

func (M *MysqlTodoRepository) Create(userId int, todo *types.Todo) error {
	m := &Todo{UserId: userId, Id: todo.Id, Titles: todo.Titles, Done: todo.Done, Content: todo.Content}
	err := M.db.Create(m).Error
	if err != nil {
		logx.Error(err)
		return errs.ErrorService
	}
	todo.Id = m.Id
	todo.CreatedAt = m.CreatedAt.Unix()
	todo.UpdatedAt = m.UpdatedAt.Unix()
	return nil
}

func (M *MysqlTodoRepository) FindById(userId int, id int) (*types.Todo, error) {
	var todo Todo
	err := M.db.Where("user_id = ?", userId).First(&id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrorTodoNotFound
	}
	if err != nil {
		logx.Error(err)
		return nil, errs.ErrorService
	}
	return &types.Todo{Id: todo.Id, Titles: todo.Titles, Done: todo.Done, Content: todo.Content, CreatedAt: todo.CreatedAt.Unix(), UpdatedAt: todo.UpdatedAt.Unix()}, nil
}

func (M *MysqlTodoRepository) FindAll(userId int) ([]*types.Todo, error) {
	var todos []*Todo
	err := M.db.Where("user_id = ?", userId).Find(&todos).Error
	if err != nil {
		logx.Error(err)
		return nil, errs.ErrorService
	}
	result := make([]*types.Todo, 0, len(todos))
	for _, v := range todos {
		result = append(result, &types.Todo{Id: v.Id, Titles: v.Titles, Done: v.Done, Content: v.Content, CreatedAt: v.CreatedAt.Unix(), UpdatedAt: v.UpdatedAt.Unix()})
	}
	return result, nil
}

func (M *MysqlTodoRepository) Update(userId int, todo *types.Todo) error {
	var m Todo
	err := M.db.Where("user_id = ?", userId).First(&m, todo.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrorTodoNotFound
	}
	if err != nil {
		logx.Error(err)
		return errs.ErrorService
	}
	m.Titles = todo.Titles
	m.Content = todo.Content
	m.Done = todo.Done
	err2 := M.db.Model(&m).Select("Titles", "Done", "Content").Updates(&m).Error
	if err2 != nil {
		logx.Error(err2)
		return errs.ErrorService
	}
	return nil
}

func (M *MysqlTodoRepository) Delete(userId int, id int) error {
	var todo Todo
	err := M.db.Where("user_id = ?", userId).First(&todo, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrorTodoNotFound
	}
	if err != nil {
		logx.Error(err)
		return errs.ErrorService
	}
	err2 := M.db.Delete(&Todo{}, id).Error
	if err2 != nil {
		logx.Error(err2)
		return errs.ErrorService
	}
	return nil
}
