// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"errors"
	"todo/internal/errs"

	"todo/internal/svc"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRegisterLogic {
	return &UserRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRegisterLogic) UserRegister(req *types.UserRegisterReq) (resp *types.UserInfo, err error) {
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return nil, errs.ErrorUserNull
	}
	_, err1 := l.svcCtx.UserRepo.FindByUsername(req.Username)
	if err1 == nil {
		return nil, errs.ErrorUserAlreadyExists
	}
	if !errors.Is(err1, errs.ErrorUserNotFound) {
		return nil, err1
	}
	_, err2 := l.svcCtx.UserRepo.FindByEmail(req.Email)
	if err2 == nil {
		return nil, errs.ErrorUserAlreadyExists
	}
	if !errors.Is(err2, errs.ErrorUserNotFound) {
		return nil, err2
	}
	HashPassword, err3 := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err3 != nil {
		return nil, err3
	}
	temp := &types.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(HashPassword),
	}
	err = l.svcCtx.UserRepo.Create(temp)
	if err != nil {
		return nil, err
	}
	return &types.UserInfo{Id: temp.Id, Username: temp.Username, Email: temp.Email, CreatedAt: temp.CreatedAt, UpdatedAt: temp.UpdatedAt}, nil
}
