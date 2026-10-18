package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"todo/internal/svc"
	"todo/internal/types"

	"golang.org/x/crypto/bcrypt"
)

func newUserTestSvc() (*svc.ServiceContext, *fakeUserRepo, *fakeTokenRepo) {
	userRepo := newFakeUserRepo()
	tokenRepo := newFakeTokenRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte(testPass), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	userRepo.data["alice"] = &types.User{
		Id: testUserID, Username: "alice", Email: "alice@example.com", PasswordHash: string(hash),
	}
	userRepo.data["bob"] = &types.User{
		Id: bobUserID, Username: "bob", Email: "bob@example.com", PasswordHash: string(hash),
	}
	tokenRepo.data[testToken] = &types.Token{
		UserId: testUserID, ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	svcCtx := &svc.ServiceContext{UserRepo: userRepo, TokenRepo: tokenRepo, TodoRepo: newFakeTodoRepo()}
	return svcCtx, userRepo, tokenRepo
}

func TestGetUserInfo_NoCookie(t *testing.T) {
	svcCtx, _, _ := newUserTestSvc()
	w := doAs(svcCtx, "", GetUserInfoHandler(svcCtx), http.MethodGet, "/user/me", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未登录应 401，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestGetUserInfo_ExpiredToken(t *testing.T) {
	svcCtx, _, tokenRepo := newUserTestSvc()
	tokenRepo.data["expired-token"] = &types.Token{
		UserId: testUserID, ExpiresAt: time.Now().Add(-time.Hour).Unix(), // 已过期
	}
	w := doAs(svcCtx, "expired-token", GetUserInfoHandler(svcCtx), http.MethodGet, "/user/me", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("过期 token 应 401，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestGetUserInfo_Success(t *testing.T) {
	svcCtx, _, _ := newUserTestSvc()
	w := doAuthed(svcCtx, GetUserInfoHandler(svcCtx), http.MethodGet, "/user/me", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var got types.UserInfo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败: %v，body=%s", err, w.Body.String())
	}
	if got.Username != "alice" {
		t.Errorf("期望 username=alice，实际 %q", got.Username)
	}
	if got.Email != "alice@example.com" {
		t.Errorf("期望 email=alice@example.com，实际 %q", got.Email)
	}
}

func TestLogout_ThenOldTokenDenied(t *testing.T) {
	svcCtx, _, tokenRepo := newUserTestSvc()
	w := doAuthed(svcCtx, UserLogOutHandler(svcCtx), http.MethodPost, "/user/logout", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("退出应 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if _, ok := tokenRepo.data[testToken]; ok {
		t.Fatal("退出后 token 仍存在于服务端（假退出）")
	}
	w = doAs(svcCtx, testToken, GetUserInfoHandler(svcCtx), http.MethodGet, "/user/me", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("退出后旧 token 应 401，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateEmail_Success(t *testing.T) {
	svcCtx, userRepo, _ := newUserTestSvc()
	body := `{"email":"newmail@example.com"}`
	w := doAuthed(svcCtx, UserUpdateHandler(svcCtx), http.MethodPut, "/user/me", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var got types.UserInfo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败: %v，body=%s", err, w.Body.String())
	}
	if got.Email != "newmail@example.com" {
		t.Errorf("响应里的 email 应是新值，实际 %q", got.Email)
	}
	if userRepo.data["alice"].Email != "newmail@example.com" {
		t.Errorf("仓库里的 email 没更新，实际 %q", userRepo.data["alice"].Email)
	}
}

func TestUpdateEmail_Duplicate(t *testing.T) {
	svcCtx, _, _ := newUserTestSvc()
	body := `{"email":"bob@example.com"}` // bob 已占用
	w := doAuthed(svcCtx, UserUpdateHandler(svcCtx), http.MethodPut, "/user/me", body, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateEmail_SameAsOwn(t *testing.T) {
	svcCtx, _, _ := newUserTestSvc()
	body := `{"email":"alice@example.com"}` // 自己当前的邮箱：应允许（幂等），不是 409
	w := doAuthed(svcCtx, UserUpdateHandler(svcCtx), http.MethodPut, "/user/me", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("改自己的邮箱应 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestChangePassword_WrongOld(t *testing.T) {
	svcCtx, _, _ := newUserTestSvc()
	body := `{"old_password":"wrongpass","new_password":"newpass999"}`
	w := doAuthed(svcCtx, UserChangePasswordHandler(svcCtx), http.MethodPut, "/user/me/password", body, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("旧密码错误应 400，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "原密码错误") {
		t.Errorf("提示应为『原密码错误』，实际 body=%s", w.Body.String())
	}
}

func TestChangePassword_Success(t *testing.T) {
	svcCtx, userRepo, _ := newUserTestSvc()
	body := `{"old_password":"` + testPass + `","new_password":"newpass999"}`
	w := doAuthed(svcCtx, UserChangePasswordHandler(svcCtx), http.MethodPut, "/user/me/password", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	hash := userRepo.data["alice"].PasswordHash
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("newpass999")); err != nil {
		t.Errorf("新密码应能通过校验，实际失败: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(testPass)); err == nil {
		t.Error("旧密码不应再通过校验")
	}
}
