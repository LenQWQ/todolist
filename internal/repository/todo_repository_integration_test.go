package repository

import (
	"errors"
	"os"
	"testing"

	"todo/internal/errs"
	"todo/internal/types"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================================================
// 集成测试：连真实 MySQL，验证【仓库层】的行为（SQL 条件、影响行数、错误映射）
//
// 为什么需要它：
//   单元测试用的是 fakeTodoRepo（内存假实现），它永远不会执行真实 SQL。
//   所以像 "First(&id)" 这种写错的查询，单测全绿也发现不了。
//
// 运行方式：
//   1) 先建一个独立的测试库（不要用开发库，测试会清表）：
//        CREATE DATABASE IF NOT EXISTS todo_test CHARACTER SET utf8mb4;
//   2) 设置环境变量并运行：
//        $env:TEST_MYSQL_DSN="root:你的密码@tcp(127.0.0.1:3306)/todo_test?charset=utf8mb4&parseTime=True"
//        go test ./internal/repository/ -run Integration -v
//   3) 未设置 TEST_MYSQL_DSN 时自动跳过（保证 CI 无数据库也能跑其余测试）
// ============================================================================

const integrationDSNEnv = "TEST_MYSQL_DSN"

// setupTestDB 连接测试库并建表，每个用例开始前清空数据
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv(integrationDSNEnv)
	if dsn == "" {
		t.Skipf("未设置 %s，跳过集成测试", integrationDSNEnv)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 安静点，别刷 SQL
	})
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}

	if err := db.AutoMigrate(&Todo{}, &User{}, &Token{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	if err := db.Exec("DELETE FROM todos").Error; err != nil {
		t.Fatalf("清空 todos 失败: %v", err)
	}
	return db
}

// TestIntegration_TodoCRUD 走一遍完整的仓库层往返：
// Create → FindById → FindAll → Update → Delete，并覆盖归属隔离。
func TestIntegration_TodoCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlTodoRepository(db)

	const (
		alice = 1
		bob   = 2
	)

	// ---------- ① Create ----------
	todo := &types.Todo{Titles: "集成测试标题", Content: "原始内容", Done: false}
	if err := repo.Create(alice, todo); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if todo.Id == 0 {
		t.Fatal("Create 应回填自增 id")
	}

	// ---------- ② FindById：必须按 id 精确查询 ----------
	// 这一条就能抓到 "First(&id)" 这类 bug：
	//   - 查询条件里没有 id → 返回的可能是另一条记录
	//   - 结果没扫进 todo → 字段全是零值
	got, err := repo.FindById(todo.Id, alice)
	if err != nil {
		t.Fatalf("FindById(自己的) 应成功，实际报错: %v", err)
	}
	if got.Id != todo.Id {
		t.Errorf("FindById 返回了错误的记录: 期望 id=%d，实际 id=%d", todo.Id, got.Id)
	}
	if got.Titles != "集成测试标题" {
		t.Errorf("标题不对: 期望 %q，实际 %q", "集成测试标题", got.Titles)
	}
	if got.Content != "原始内容" {
		t.Errorf("内容不对: 期望 %q，实际 %q", "原始内容", got.Content)
	}

	// ---------- ③ 不存在的 id → 应返回"不存在"而不是空对象 ----------
	if _, err := repo.FindById(999999, alice); !errors.Is(err, errs.ErrorTodoNotFound) {
		t.Errorf("不存在的 id 应返回 ErrorTodoNotFound，实际: %v", err)
	}

	// ---------- ④ 别人的 id → 同样"不存在"（数据隔离） ----------
	if _, err := repo.FindById(todo.Id, bob); !errors.Is(err, errs.ErrorTodoNotFound) {
		t.Errorf("访问他人待办应返回 ErrorTodoNotFound，实际: %v", err)
	}

	// ---------- ⑤ FindAll：只返回自己的 ----------
	// 再造一条 bob 的数据，确保过滤真的生效（否则删掉 WHERE 也能通过）
	bobTodo := &types.Todo{Titles: "bob 的待办"}
	if err := repo.Create(bob, bobTodo); err != nil {
		t.Fatalf("为 bob 创建失败: %v", err)
	}

	aliceList, err := repo.FindAll(alice)
	if err != nil {
		t.Fatalf("FindAll(alice) 失败: %v", err)
	}
	if len(aliceList) != 1 {
		t.Fatalf("alice 应只看到自己的 1 条，实际 %d 条", len(aliceList))
	}
	if aliceList[0].Id != todo.Id {
		t.Errorf("alice 看到的应是 id=%d，实际 id=%d", todo.Id, aliceList[0].Id)
	}

	bobList, err := repo.FindAll(bob)
	if err != nil {
		t.Fatalf("FindAll(bob) 失败: %v", err)
	}
	if len(bobList) != 1 || bobList[0].Id != bobTodo.Id {
		t.Errorf("bob 应只看到自己的 1 条(id=%d)，实际 %+v", bobTodo.Id, bobList)
	}

	// ---------- ⑥ Update ----------
	todo.Titles = "改后的标题"
	todo.Content = "改后的内容"
	todo.Done = true
	if err := repo.Update(alice, todo); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}
	updated, err := repo.FindById(todo.Id, alice)
	if err != nil {
		t.Fatalf("Update 后查询失败: %v", err)
	}
	if updated.Titles != "改后的标题" || updated.Content != "改后的内容" || !updated.Done {
		t.Errorf("Update 未生效: %+v", updated)
	}

	// 别人不能改（改了也不该影响原数据）
	todo.Titles = "被篡改"
	if err := repo.Update(bob, todo); !errors.Is(err, errs.ErrorTodoNotFound) {
		t.Errorf("他人 Update 应返回 ErrorTodoNotFound，实际: %v", err)
	}
	check, _ := repo.FindById(todo.Id, alice)
	if check.Titles == "被篡改" {
		t.Error("数据被越权修改了")
	}

	// ---------- ⑦ Delete：别人的删不掉 ----------
	if err := repo.Delete(todo.Id, bob); !errors.Is(err, errs.ErrorTodoNotFound) {
		t.Errorf("他人 Delete 应返回 ErrorTodoNotFound，实际: %v", err)
	}
	if _, err := repo.FindById(todo.Id, alice); err != nil {
		t.Error("越权删除后，alice 的数据不应消失")
	}

	// ---------- ⑧ Delete：自己的能删 ----------
	if err := repo.Delete(todo.Id, alice); err != nil {
		t.Fatalf("删除自己的待办失败: %v", err)
	}
	if _, err := repo.FindById(todo.Id, alice); !errors.Is(err, errs.ErrorTodoNotFound) {
		t.Errorf("删除后再查应返回 ErrorTodoNotFound，实际: %v", err)
	}

	// ---------- ⑨ 再删一次 → 仍然"不存在"（验证 RowsAffected 判断） ----------
	if err := repo.Delete(todo.Id, alice); !errors.Is(err, errs.ErrorTodoNotFound) {
		t.Errorf("重复删除应返回 ErrorTodoNotFound，实际: %v", err)
	}
}
