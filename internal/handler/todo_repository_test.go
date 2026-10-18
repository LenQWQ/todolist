package handler

import (
	"todo/internal/errs"
	"todo/internal/repository"
	"todo/internal/types"
)

var _ repository.TodoRepository = (*fakeTodoRepo)(nil)

type fakeTodoRepo struct {
	data  map[int]*types.Todo
	owner map[int]int
	next  int
}

func newFakeTodoRepo() *fakeTodoRepo {
	return &fakeTodoRepo{
		data:  make(map[int]*types.Todo),
		owner: make(map[int]int),
		next:  1,
	}
}

// seed 预置一条属于 userID 的待办（测试辅助）
func (f *fakeTodoRepo) seed(userID int, todo *types.Todo) {
	f.data[todo.Id] = todo
	f.owner[todo.Id] = userID
	if todo.Id >= f.next {
		f.next = todo.Id + 1
	}
}

func (f *fakeTodoRepo) Create(userID int, todo *types.Todo) error {
	todo.Id = f.next
	f.next++
	f.data[todo.Id] = todo
	f.owner[todo.Id] = userID
	return nil
}

func (f *fakeTodoRepo) FindById(id int, userID int) (*types.Todo, error) {
	t, ok := f.data[id]
	if !ok || f.owner[id] != userID {
		// 不存在 或 不属于该用户 → 都当作"不存在"（不暴露数据是否存在）
		return nil, errs.ErrorTodoNotFound
	}
	return t, nil
}

func (f *fakeTodoRepo) FindAll(userID int) ([]*types.Todo, error) {
	result := make([]*types.Todo, 0, len(f.data))
	for id, t := range f.data {
		if f.owner[id] == userID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (f *fakeTodoRepo) Update(userID int, todo *types.Todo) error {
	if _, ok := f.data[todo.Id]; !ok || f.owner[todo.Id] != userID {
		return errs.ErrorTodoNotFound
	}
	f.data[todo.Id] = todo
	return nil
}

func (f *fakeTodoRepo) Delete(id int, userID int) error {
	if _, ok := f.data[id]; !ok || f.owner[id] != userID {
		return errs.ErrorTodoNotFound
	}
	delete(f.data, id)
	delete(f.owner, id)
	return nil
}
