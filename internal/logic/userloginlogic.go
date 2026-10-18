// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"
	"todo/internal/errs"

	"todo/internal/svc"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLoginLogic) UserLogin(req *types.UserLoginReq) (resp *types.UserLoginResp, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, errs.ErrorUserNull
	}
	user, err1 := l.svcCtx.UserRepo.FindByUsername(req.Username)
	if errors.Is(err1, errs.ErrorUserNotFound) {
		return nil, errs.ErrorInvalidCredentials
	}
	if err1 != nil {
		return nil, err1
	}
	err2 := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err2 != nil {
		return nil, errs.ErrorInvalidCredentials
	}
	b := make([]byte, 32)
	rand.Read(b)
	tokenStr := hex.EncodeToString(b)
	expiresAt := time.Now().Unix() + int64(l.svcCtx.Config.Token.ExpireSeconds)
	temp := &types.Token{
		UserId:    user.Id,
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}
	err3 := l.svcCtx.TokenRepo.Create(temp)
	if err3 != nil {
		return nil, err3
	}
	resp = &types.UserLoginResp{
		Id:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     tokenStr,
	}
	return resp, nil
}
