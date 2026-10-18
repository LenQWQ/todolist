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

type DeleteTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTodoLogic {
	return &DeleteTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTodoLogic) DeleteTodo(req *types.DeleteTodoReq) error {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return errs.ErrorInvalidToken
	}
	id := req.Id
	err := l.svcCtx.TodoRepo.Delete(info.UserID, id)
	if err != nil {
		return errs.ErrorTodoNotFound
	} else {
		return nil
	}
}
