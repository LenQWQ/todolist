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

type GetTodoInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTodoInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTodoInfoLogic {
	return &GetTodoInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTodoInfoLogic) GetTodoInfo(req *types.GetTodoReq) (resp *types.Todo, err error) {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return nil, errs.ErrorInvalidToken
	}
	id := req.Id
	temp, err := l.svcCtx.TodoRepo.FindById(info.UserID, id)
	if err != nil {
		return nil, errs.ErrorTodoNotFound
	}
	return temp, nil
}
