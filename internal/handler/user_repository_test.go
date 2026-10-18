package handler

import (
	"todo/internal/errs"
	"todo/internal/repository"
	"todo/internal/types"
)

var _ repository.UserRepository = (*fakeUserRepo)(nil)

type fakeUserRepo struct {
	data map[string]*types.User
	next int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{data: make(map[string]*types.User), next: 1}
}

func (f *fakeUserRepo) Create(u *types.User) error {
	u.Id = f.next
	f.next++
	f.data[u.Username] = u
	return nil
}

func (f *fakeUserRepo) FindById(id int) (*types.User, error) {
	for _, u := range f.data {
		if u.Id == id {
			return u, nil
		}
	}
	return nil, errs.ErrorUserNotFound
}

func (f *fakeUserRepo) FindByUsername(username string) (*types.User, error) {
	u, ok := f.data[username]
	if !ok {
		return nil, errs.ErrorUserNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*types.User, error) {
	for _, u := range f.data {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errs.ErrorUserNotFound
}

func (f *fakeUserRepo) UpdateEmail(id int, email string) error {
	for _, u := range f.data {
		if u.Id == id {
			u.Email = email
			return nil
		}
	}
	return errs.ErrorUserNotFound
}

func (f *fakeUserRepo) UpdatePassword(id int, passwordHash string) error {
	for _, u := range f.data {
		if u.Id == id {
			u.PasswordHash = passwordHash
			return nil
		}
	}
	return errs.ErrorUserNotFound
}
