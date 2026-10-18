// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"todo/internal/config"
	"todo/internal/repository"
)

type ServiceContext struct {
	Config    config.Config
	TodoRepo  repository.TodoRepository
	UserRepo  repository.UserRepository
	TokenRepo repository.TokenRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := repository.InitDB(c.Mysql.DataSource)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&repository.Todo{}, &repository.User{}, &repository.Token{})
	return &ServiceContext{
		Config:    c,
		TodoRepo:  repository.NewMysqlTodoRepository(db),
		UserRepo:  repository.NewMysqlUserRepository(db),
		TokenRepo: repository.NewMysqlTokenRepository(db),
	}
}
