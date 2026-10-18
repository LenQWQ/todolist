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
	"golang.org/x/crypto/bcrypt"
)

type UserChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserChangePasswordLogic {
	return &UserChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserChangePasswordLogic) UserChangePassword(req *types.UserChangePasswordReq) error {
	info, ok := middleware.AuthInfoFrom(l.ctx)
	if !ok {
		return errs.ErrorInvalidToken
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return errs.ErrorUserNull
	}
	user, err := l.svcCtx.UserRepo.FindById(info.UserID)
	if err != nil {
		return err
	}
	err2 := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err2 != nil {
		return &errs.CodeError{
			Code: 400,
			Msg:  "原密码错误",
		}
	}
	password, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err4 := l.svcCtx.UserRepo.UpdatePassword(info.UserID, string(password))
	if err4 != nil {
		return err4
	}
	return nil
}
