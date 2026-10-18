// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"todo/internal/errs"
	"todo/internal/middleware"
	"todo/internal/svc"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTodoLogic {
	return &CreateTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTodoLogic) CreateTodo(req *types.CreateTodoReq) (resp *types.Todo, err error) {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return nil, errs.ErrorInvalidToken
	}
	if req.Titles == "" {
		return nil, errs.ErrorTodoNull
	}
	temp := &types.Todo{
		Titles:  req.Titles,
		Done:    false,
		Content: req.Content,
	}
	err = l.svcCtx.TodoRepo.Create(info.UserID, temp)
	return temp, err
}
