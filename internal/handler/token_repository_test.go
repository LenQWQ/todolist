package handler

import (
	"todo/internal/errs"
	"todo/internal/repository"
	"todo/internal/types"
)

var _ repository.TokenRepository = (*fakeTokenRepo)(nil)

type fakeTokenRepo struct {
	data map[string]*types.Token
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{data: make(map[string]*types.Token)}
}

func (f *fakeTokenRepo) Create(token *types.Token) error {
	f.data[token.Token] = token
	return nil
}

func (f *fakeTokenRepo) FindByToken(token string) (*types.Token, error) {
	t, ok := f.data[token]
	if !ok {
		return nil, errs.ErrorInvalidToken
	}
	return t, nil
}

func (f *fakeTokenRepo) DeleteByToken(token string) error {
	delete(f.data, token)
	return nil
}
