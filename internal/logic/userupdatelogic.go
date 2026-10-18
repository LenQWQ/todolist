// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"todo/internal/errs"
	"todo/internal/middleware"

	"todo/internal/svc"
	"todo/internal/types"

	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateLogic {
	return &UserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserUpdateLogic) UserUpdate(req *types.UserUpdateReq) (resp *types.UserInfo, err error) {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return nil, errs.ErrorUserNotFound
	}
	if req.Email == "" {
		return nil, errs.ErrorUserNull
	}
	existing, err1 := l.svcCtx.UserRepo.FindByEmail(req.Email)
	if err1 == nil && existing.Id != info.UserID {
		return nil, errs.ErrorEmailAlreadyExists
	}
	if err1 != nil && !errors.Is(err1, errs.ErrorUserNotFound) {
		return nil, err1
	}
	if err2 := l.svcCtx.UserRepo.UpdateEmail(info.UserID, req.Email); err2 != nil {
		return nil, err2
	}
	u, err3 := l.svcCtx.UserRepo.FindById(info.UserID)
	if err3 != nil {
		return nil, err3
	}
	return &types.UserInfo{
		Id: u.Id, Username: u.Username, Email: u.Email,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}, nil
}
