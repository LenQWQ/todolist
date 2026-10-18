// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"todo/internal/errs"
	"todo/internal/middleware"

	"todo/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogOutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogOutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogOutLogic {
	return &UserLogOutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogOutLogic) UserLogOut() error {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return errs.ErrorInvalidToken
	}
	err := l.svcCtx.TokenRepo.DeleteByToken(info.Token)
	if err != nil {
		return err
	}
	return nil
}
