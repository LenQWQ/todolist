package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"todo/internal/svc"
	"todo/internal/types"

	"golang.org/x/crypto/bcrypt"
)

func TestUserRegister_Success(t *testing.T) {
	svcCtx := &svc.ServiceContext{UserRepo: newFakeUserRepo()}
	wantUserName := "LenQWQ"
	wantEmail := "123@.com"
	wantPassword := "123456"
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, wantUserName, wantEmail, wantPassword)
	w := doRequest(UserRegisterHandler(svcCtx), http.MethodPost, "/user/register", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200,实际%v,body=%v", w.Code, w.Body.String())
	}
	var got types.UserInfo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败:%v,body=%s", err, w.Body.String())
	}
	if got.Id == 0 {
		t.Errorf("期望id非0，实际%d", got.Id)
	}
	if got.Username != "LenQWQ" {
		t.Errorf("期望Username=LenQWQ,实际 %q", got.Username)
	}
	if got.Email != wantEmail {
		t.Errorf("期望Email=123@.com,实际 %q", got.Email)
	}
}

func TestUserRegister_DuplicateUsername(t *testing.T) {
	svcCtx := &svc.ServiceContext{UserRepo: newFakeUserRepo()}
	wantUserName := "LenQWQ"
	wantEmail := "123@.com"
	wantPassword := "123456"
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, wantUserName, wantEmail, wantPassword)
	doRequest(UserRegisterHandler(svcCtx), http.MethodPost, "/user/register", body, nil)
	anoEmail := "233@.com"
	body2 := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, wantUserName, anoEmail, wantPassword)
	w := doRequest(UserRegisterHandler(svcCtx), http.MethodPost, "/user/register", body2, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestUserRegister_DuplicateEmail(t *testing.T) {
	svcCtx := &svc.ServiceContext{UserRepo: newFakeUserRepo()}
	wantUserName := "LenQWQ"
	wantEmail := "123@.com"
	wantPassword := "123456"
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, wantUserName, wantEmail, wantPassword)
	doRequest(UserRegisterHandler(svcCtx), http.MethodPost, "/user/register", body, nil)
	anoUsername := "LenQAQ"
	body2 := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, anoUsername, wantEmail, wantPassword)
	w := doRequest(UserRegisterHandler(svcCtx), http.MethodPost, "/user/register", body2, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestUserRegister_EmptyField(t *testing.T) {
	svcCtx := &svc.ServiceContext{UserRepo: newFakeUserRepo()}
	wantUserName := ""
	wantEmail := "123@.com"
	wantPassword := "123456"
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, wantUserName, wantEmail, wantPassword)
	w := doRequest(UserRegisterHandler(svcCtx), http.MethodPost, "/user/register", body, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestUserLogin_Success(t *testing.T) {
	userRepo := newFakeUserRepo()
	tokenRepo := newFakeTokenRepo()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("生成散列失败: %v", err)
	}
	userRepo.data["alice"] = &types.User{
		Id: 1, Username: "alice", Email: "a@b.com", PasswordHash: string(hash),
	}
	svcCtx := &svc.ServiceContext{UserRepo: userRepo, TokenRepo: tokenRepo}
	body := `{"username":"alice","password":"secret123"}`
	w := doRequest(UserLoginHandler(svcCtx), http.MethodPost, "/user/login", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望200 实际%v,body=%v", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("期望设置 Set-Cookie，实际没有")
	}
	if cookies[0].Name != "token" {
		t.Fatalf("token不匹配")
	}
	if !cookies[0].HttpOnly {
		t.Error("Cookie 应该是 HttpOnly")
	}
	if strings.Contains(w.Body.String(), cookies[0].Value) {
		t.Error("响应体泄露了 token")
	}
	if len(tokenRepo.data) != 1 {
		t.Fatalf("期望假 repo 里存了 1 条 token，实际 %d", len(tokenRepo.data))
	}
	for _, tk := range tokenRepo.data {
		if tk.UserId != 1 {
			t.Errorf("期望 token.UserId=1，实际 %d", tk.UserId)
		}
	}
}

func TestUserLogin_WrongPassword(t *testing.T) {
	userRepo := newFakeUserRepo()
	tokenRepo := newFakeTokenRepo()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("生成散列失败: %v", err)
	}
	userRepo.data["alice"] = &types.User{
		Id: 1, Username: "alice", Email: "a@b.com", PasswordHash: string(hash),
	}
	svcCtx := &svc.ServiceContext{UserRepo: userRepo, TokenRepo: tokenRepo}
	body := `{"username":"alice","password":"secret234"}`
	w := doRequest(UserLoginHandler(svcCtx), http.MethodPost, "/user/login", body, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望401 实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestUserLogin_UserNotFound(t *testing.T) {
	userRepo := newFakeUserRepo()
	tokenRepo := newFakeTokenRepo()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("生成散列失败: %v", err)
	}
	userRepo.data[""] = &types.User{
		Id: 1, Username: "", Email: "a@b.com", PasswordHash: string(hash),
	}
	svcCtx := &svc.ServiceContext{UserRepo: userRepo, TokenRepo: tokenRepo}
	body := `{"username":"alice","password":"secret123"}`
	w := doRequest(UserLoginHandler(svcCtx), http.MethodPost, "/user/login", body, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望401 实际%v,body=%v", w.Code, w.Body.String())
	}
}
