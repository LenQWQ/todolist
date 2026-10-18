package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"todo/internal/middleware"
	"todo/internal/svc"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"golang.org/x/crypto/bcrypt"
)

// ---------------- 测试常量 ----------------

const (
	cookieName = "token"

	testUserID = 1
	testToken  = "token-for-alice"

	bobUserID = 2
	bobToken  = "token-for-bob"

	testPass = "secret123"
)

func TestMain(m *testing.M) {
	httpx.SetErrorHandler(HandleError)
	os.Exit(m.Run())
}

func doRequest(h http.HandlerFunc, method, path, body string, vars map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if len(vars) > 0 {
		req = pathvar.WithVars(req, vars)
	}
	w := httptest.NewRecorder()
	h(w, req)
	return w
}

func newTestSvc() (*svc.ServiceContext, *fakeTodoRepo, *fakeTokenRepo) {
	todoRepo := newFakeTodoRepo()
	tokenRepo := newFakeTokenRepo()
	userRepo := newFakeUserRepo()

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

	expire := time.Now().Add(time.Hour).Unix()
	tokenRepo.data[testToken] = &types.Token{UserId: testUserID, ExpiresAt: expire}
	tokenRepo.data[bobToken] = &types.Token{UserId: bobUserID, ExpiresAt: expire}

	svcCtx := &svc.ServiceContext{TodoRepo: todoRepo, UserRepo: userRepo, TokenRepo: tokenRepo}
	return svcCtx, todoRepo, tokenRepo
}

func doAs(svcCtx *svc.ServiceContext, token string, h http.HandlerFunc,
	method, path, body string, vars map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: token})
	}
	if len(vars) > 0 {
		req = pathvar.WithVars(req, vars)
	}
	w := httptest.NewRecorder()
	middleware.Auth(svcCtx)(h)(w, req) // 关键：套上鉴权中间件
	return w
}

func doAuthed(svcCtx *svc.ServiceContext, h http.HandlerFunc,
	method, path, body string, vars map[string]string) *httptest.ResponseRecorder {
	return doAs(svcCtx, testToken, h, method, path, body, vars)
}

func TestCreateTodo_Success(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	wantTitle := "学GO"
	wantContent := "内容"
	body := fmt.Sprintf(`{"titles":"%s","content":"%s"}`, wantTitle, wantContent)
	w := doAuthed(svcCtx, CreateTodoHandler(svcCtx), http.MethodPost, "/todos", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200,实际%v,body=%v", w.Code, w.Body.String())
	}
	var got types.Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败:%v,body=%s", err, w.Body.String())
	}
	if got.Id == 0 {
		t.Errorf("期望id非0，实际%d", got.Id)
	}
	if got.Titles != "学GO" {
		t.Errorf("期望titles=学GO,实际 %q", got.Titles)
	}
}

func TestCreateTodo_EmptyBody(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	body := `{"titles":"","content":" 内容"}`
	w := doAuthed(svcCtx, CreateTodoHandler(svcCtx), http.MethodPost, "/todos", body, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestCreateTodo_BadJSON(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	body := `{BadJSON}`
	w := doAuthed(svcCtx, CreateTodoHandler(svcCtx), http.MethodPost, "/todos", body, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestCreateTodo_MissingTitle(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	body := `{"content":"Titles不存在"}`
	w := doAuthed(svcCtx, CreateTodoHandler(svcCtx), http.MethodPost, "/todos", body, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestGetTodoList_Empty(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	w := doAuthed(svcCtx, GetTodoListHandler(svcCtx), http.MethodGet, "/todos", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200,实际%v,body=%v", w.Code, w.Body.String())
	}
	var got types.TodoListResp
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败:%v,body=%s", err, w.Body.String())
	}
	if len(got.Todos) != 0 {
		t.Errorf("期望0条,实际%v条", len(got.Todos))
	}
}

func TestGetTodoList_Multiple(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "学 Go", Content: "内容1", Done: false})
	repo.seed(testUserID, &types.Todo{Id: 2, Titles: "学 SQL", Content: "内容2", Done: true})

	w := doAuthed(svcCtx, GetTodoListHandler(svcCtx), http.MethodGet, "/todos", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际%v,body=%v", w.Code, w.Body.String())
	}
	var got types.TodoListResp
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败:%v,body=%s", err, w.Body.String())
	}
	if len(got.Todos) != 2 {
		t.Errorf("期望2条,实际%v条", len(got.Todos))
	}
	byId := make(map[int]types.Todo, len(got.Todos))
	for _, td := range got.Todos {
		byId[td.Id] = td
	}
	if byId[1].Titles != "学 Go" {
		t.Errorf("期望 id=1 titles=%q，实际 %q", "学 Go", byId[1].Titles)
	}
	if byId[2].Titles != "学 SQL" {
		t.Errorf("期望 id=2 titles=%q，实际 %q", "学 SQL", byId[2].Titles)
	}
	if byId[2].Done != true {
		t.Errorf("期望 id=2 done=true，实际 %v", byId[2].Done)
	}
}

func TestGetTodoInfo_Success(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	wantTitle := "学GO"
	wantContent := "内容"
	body := fmt.Sprintf(`{"titles":"%s","content":"%s"}`, wantTitle, wantContent)
	w := doAuthed(svcCtx, CreateTodoHandler(svcCtx), http.MethodPost, "/todos", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("创建失败: %v, body=%s", w.Code, w.Body.String())
	}
	w = doAuthed(svcCtx, GetTodoInfoHandler(svcCtx), http.MethodGet, "/todos/1", "", map[string]string{"id": "1"})
	if w.Code != http.StatusOK {
		t.Fatalf("期望200,实际%v,body=%v", w.Code, w.Body.String())
	}
	var got types.Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败:%v,body=%s", err, w.Body.String())
	}
	if got.Id != 1 {
		t.Errorf("期望id=1，实际%d", got.Id)
	}
	if got.Titles != wantTitle {
		t.Errorf("期望titles=%q,实际 %q", wantTitle, got.Titles)
	}
}

func TestGetTodoInfo_NotFound(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	w := doAuthed(svcCtx, GetTodoInfoHandler(svcCtx), http.MethodGet, "/todos/999", "", map[string]string{"id": "999"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestGetTodoInfo_BadId(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	w := doAuthed(svcCtx, GetTodoInfoHandler(svcCtx), http.MethodGet, "/todos/abc", "", map[string]string{"id": "abc"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestUpdateTodo_Success(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "旧标题", Content: "旧内容", Done: false})
	wantTitle := "新标题"
	wantContent := "新内容"
	body := fmt.Sprintf(`{"titles":%q,"content":%q,"done":false}`, wantTitle, wantContent)
	w := doAuthed(svcCtx, UpdateTodoHandler(svcCtx), http.MethodPut, "/todos/1", body, map[string]string{"id": "1"})
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var got types.Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败: %v，body=%s", err, w.Body.String())
	}
	if got.Id != 1 {
		t.Errorf("期望 id=1，实际 %d", got.Id)
	}
	if got.Titles != wantTitle {
		t.Errorf("期望 titles=%q，实际 %q", wantTitle, got.Titles)
	}
	if got.Content != wantContent {
		t.Errorf("期望 content=%q，实际 %q", wantContent, got.Content)
	}
	if got.Done != false {
		t.Errorf("期望 done=false，实际 %v", got.Done)
	}
	if repo.data[1].Titles != wantTitle {
		t.Errorf("假 repo 里的数据没更新，实际 %q", repo.data[1].Titles)
	}
}

func TestUpdateTodo_NotFound(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "旧标题", Content: "旧内容", Done: false})
	body := `{"titles":"新标题","content":"新内容","done":false}`
	w := doAuthed(svcCtx, UpdateTodoHandler(svcCtx), http.MethodPut, "/todos/999", body, map[string]string{"id": "999"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404,实际 %v", w.Code)
	}
}

func TestUpdateTodo_Empty(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "旧标题", Content: "旧内容", Done: false})
	body := `{"titles":"","content":"新内容","done":false}`
	w := doAuthed(svcCtx, UpdateTodoHandler(svcCtx), http.MethodPut, "/todos/1", body, map[string]string{"id": "1"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateTodo_DoneTrue(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "旧标题", Content: "旧内容", Done: false})
	wantTitle := "新标题"
	wantContent := "新内容"
	body := fmt.Sprintf(`{"titles":%q,"content":%q,"done":true}`, wantTitle, wantContent)
	w := doAuthed(svcCtx, UpdateTodoHandler(svcCtx), http.MethodPut, "/todos/1", body, map[string]string{"id": "1"})
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var got types.Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败: %v，body=%s", err, w.Body.String())
	}
	if got.Done != true {
		t.Errorf("期望 done=true，实际 %v", got.Done)
	}
}

func TestUpdateTodo_DoneFalse(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "旧标题", Content: "旧内容", Done: true})
	wantTitle := "新标题"
	wantContent := "新内容"
	body := fmt.Sprintf(`{"titles":%q,"content":%q,"done":false}`, wantTitle, wantContent)
	w := doAuthed(svcCtx, UpdateTodoHandler(svcCtx), http.MethodPut, "/todos/1", body, map[string]string{"id": "1"})
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var got types.Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败: %v，body=%s", err, w.Body.String())
	}
	if got.Done != false {
		t.Errorf("期望 done=false，实际 %v", got.Done)
	}
}

// ---------------- 删除 ----------------

func TestDeleteTodo_Success(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "旧标题", Content: "旧内容", Done: true})
	w := doAuthed(svcCtx, DeleteTodoHandler(svcCtx), http.MethodDelete, "/todos/1", "", map[string]string{"id": "1"})
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	w = doAuthed(svcCtx, GetTodoInfoHandler(svcCtx), http.MethodGet, "/todos/1", "", map[string]string{"id": "1"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("删除后再查应 404,实际%v,body=%v", w.Code, w.Body.String())
	}
}

func TestDeleteTodo_NotFound(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	w := doAuthed(svcCtx, DeleteTodoHandler(svcCtx), http.MethodDelete, "/todos/999", "", map[string]string{"id": "999"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

// ================= 鉴权与数据归属（安全测试） =================

// 未登录访问受保护接口 → 401
func TestTodoList_Unauthorized(t *testing.T) {
	svcCtx, _, _ := newTestSvc()
	w := doAs(svcCtx, "", GetTodoListHandler(svcCtx), http.MethodGet, "/todos", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未登录应 401，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

// 列表只返回自己的数据
func TestTodoList_OnlyOwn(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "alice的待办", Content: "a"})
	repo.seed(bobUserID, &types.Todo{Id: 2, Titles: "bob的待办", Content: "b"})

	w := doAs(svcCtx, bobToken, GetTodoListHandler(svcCtx), http.MethodGet, "/todos", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var got types.TodoListResp
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析响应失败: %v，body=%s", err, w.Body.String())
	}
	if len(got.Todos) != 1 {
		t.Fatalf("bob 应只看到自己的 1 条，实际 %d 条", len(got.Todos))
	}
	if got.Todos[0].Id != 2 {
		t.Errorf("期望只看到 id=2（自己的），实际 id=%d", got.Todos[0].Id)
	}
}

// 访问别人的详情 → 404（不暴露"存在"这个事实）
func TestTodoInfo_OtherUserDenied(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "alice的待办", Content: "a"})

	w := doAs(svcCtx, bobToken, GetTodoInfoHandler(svcCtx), http.MethodGet, "/todos/1", "", map[string]string{"id": "1"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("越权访问应 404，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

// 修改别人的待办 → 404 且数据未被篡改
func TestTodoUpdate_OtherUserDenied(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "alice的待办", Content: "a", Done: false})

	body := `{"titles":"HACKED","content":"x","done":true}`
	w := doAs(svcCtx, bobToken, UpdateTodoHandler(svcCtx), http.MethodPut, "/todos/1", body, map[string]string{"id": "1"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("越权修改应 404，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if repo.data[1].Titles != "alice的待办" {
		t.Errorf("数据被篡改了：%q（应保持 alice的待办）", repo.data[1].Titles)
	}
}

// 删除别人的待办 → 404 且数据仍在
func TestTodoDelete_OtherUserDenied(t *testing.T) {
	svcCtx, repo, _ := newTestSvc()
	repo.seed(testUserID, &types.Todo{Id: 1, Titles: "alice的待办", Content: "a"})

	w := doAs(svcCtx, bobToken, DeleteTodoHandler(svcCtx), http.MethodDelete, "/todos/1", "", map[string]string{"id": "1"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("越权删除应 404，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if _, ok := repo.data[1]; !ok {
		t.Error("alice 的待办被删掉了（不应发生）")
	}
}
