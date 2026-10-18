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

type UpdateTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTodoLogic {
	return &UpdateTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTodoLogic) UpdateTodo(req *types.UpdateTodoReq) (resp *types.Todo, err error) {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return nil, errs.ErrorInvalidToken
	}
	if req.Titles == "" {
		return nil, errs.ErrorTodoNull
	}
	todo := &types.Todo{
		Id:      req.Id,
		Titles:  req.Titles,
		Done:    req.Done,
		Content: req.Content,
	}
	err = l.svcCtx.TodoRepo.Update(info.UserID, todo)
	if err != nil {
		return nil, err
	} else {
		temp, err2 := l.svcCtx.TodoRepo.FindById(info.UserID, req.Id)
		if err2 != nil {
			return nil, err2
		}
		return temp, nil
	}
}
