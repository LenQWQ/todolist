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

type GetTodoListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTodoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTodoListLogic {
	return &GetTodoListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTodoListLogic) GetTodoList() (resp *types.TodoListResp, err error) {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return nil, errs.ErrorInvalidToken
	}
	list, err := l.svcCtx.TodoRepo.FindAll(info.UserID)
	temp := make([]types.Todo, 0, len(list))
	for _, item := range list {
		temp = append(temp, *item)
	}
	return &types.TodoListResp{Todos: temp}, nil
}
