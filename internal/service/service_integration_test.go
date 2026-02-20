//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/SergeiKhy/org-structure-api/internal/model"
	"github.com/SergeiKhy/org-structure-api/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestContainer создаёт контейнер с PostgreSQL для тестов
func setupTestContainer(t testing.TB) (*tcpostgres.PostgresContainer, *gorm.DB, context.Context) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(time.Minute)),
	)
	if err != nil {
		t.Fatalf("ошибка запуска контейнера: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("ошибка получения connection string: %v", err)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка подключения к БД: %v", err)
	}

	err = db.AutoMigrate(&model.Department{}, &model.Employee{})
	if err != nil {
		t.Fatalf("ошибка миграции: %v", err)
	}

	return pgContainer, db, ctx
}

// TestService_CreateDepartment_Integration тестирует создание подразделения
func TestService_CreateDepartment_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	req := model.CreateDepartmentRequest{
		Name:     "Engineering",
		ParentID: nil,
	}

	dept, err := svc.CreateDepartment(req)
	if err != nil {
		t.Fatalf("ошибка создания: %v", err)
	}

	if dept.ID == 0 {
		t.Error("ожидался ненулевой ID")
	}
	if dept.Name != "Engineering" {
		t.Errorf("ожидалось имя 'Engineering', получено %q", dept.Name)
	}
}

// TestService_CreateDepartment_Duplicate_Integration тестирует защиту от дубликатов
func TestService_CreateDepartment_Duplicate_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём первое
	req1 := model.CreateDepartmentRequest{Name: "Engineering"}
	_, err := svc.CreateDepartment(req1)
	if err != nil {
		t.Fatalf("ошибка создания первого: %v", err)
	}

	// Пытаемся создать дубликат
	req2 := model.CreateDepartmentRequest{Name: "Engineering"}
	_, err = svc.CreateDepartment(req2)
	if err != ErrDuplicateName {
		t.Errorf("ожидалась ошибка ErrDuplicateName, получено %v", err)
	}
}

// TestService_UpdateDepartment_Integration тестирует обновление
func TestService_UpdateDepartment_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём
	req := model.CreateDepartmentRequest{Name: "Original"}
	dept, _ := svc.CreateDepartment(req)

	// Обновляем
	updateReq := model.UpdateDepartmentRequest{Name: "Updated"}
	updated, err := svc.UpdateDepartment(dept.ID, updateReq)
	if err != nil {
		t.Fatalf("ошибка обновления: %v", err)
	}

	if updated.Name != "Updated" {
		t.Errorf("ожидалось имя 'Updated', получено %q", updated.Name)
	}
}

// TestService_UpdateDepartment_SelfParent_Integration тестирует защиту от self-parent
func TestService_UpdateDepartment_SelfParent_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	req := model.CreateDepartmentRequest{Name: "Dept"}
	dept, _ := svc.CreateDepartment(req)

	// Пытаемся сделать родителем самого себя
	updateReq := model.UpdateDepartmentRequest{ParentID: &dept.ID}
	_, err := svc.UpdateDepartment(dept.ID, updateReq)
	if err != ErrSelfParent {
		t.Errorf("ожидалась ошибка ErrSelfParent, получено %v", err)
	}
}

// TestService_DeleteDepartment_Cascade_Integration тестирует каскадное удаление
func TestService_DeleteDepartment_Cascade_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём подразделение
	req := model.CreateDepartmentRequest{Name: "ToDelete"}
	dept, _ := svc.CreateDepartment(req)

	// Создаём сотрудника
	empReq := model.CreateEmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
	}
	svc.CreateEmployee(dept.ID, empReq)

	// Удаляем
	err := svc.DeleteDepartment(dept.ID, "cascade", nil)
	if err != nil {
		t.Fatalf("ошибка удаления: %v", err)
	}

	// Проверяем что удалено
	_, err = svc.GetDepartmentTree(dept.ID, 1, true)
	if err != ErrNotFound {
		t.Error("ожидалась ошибка NotFound для удалённого подразделения")
	}
}

// TestService_DeleteDepartment_Reassign_Integration тестирует переназначение
func TestService_DeleteDepartment_Reassign_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём подразделения
	oldDept, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "Old"})
	newDept, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "New"})

	// Создаём сотрудника в старом
	empReq := model.CreateEmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
	}
	svc.CreateEmployee(oldDept.ID, empReq)

	// Удаляем с переназначением
	reassignTo := newDept.ID
	err := svc.DeleteDepartment(oldDept.ID, "reassign", &reassignTo)
	if err != nil {
		t.Fatalf("ошибка удаления: %v", err)
	}

	// Проверяем что сотрудник переназначен
	var emp model.Employee
	db.Where("full_name = ?", "John Doe").First(&emp)
	if emp.DepartmentID != newDept.ID {
		t.Errorf("ожидался department_id %d, получен %d", newDept.ID, emp.DepartmentID)
	}
}

// TestService_CreateEmployee_Integration тестирует создание сотрудника
func TestService_CreateEmployee_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём подразделение
	dept, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "Engineering"})

	// Создаём сотрудника
	req := model.CreateEmployeeRequest{
		FullName: "John Doe",
		Position: "Senior Developer",
		HiredAt:  strPtr("2024-01-15"),
	}

	emp, err := svc.CreateEmployee(dept.ID, req)
	if err != nil {
		t.Fatalf("ошибка создания сотрудника: %v", err)
	}

	if emp.ID == 0 {
		t.Error("ожидался ненулевой ID")
	}
	if emp.FullName != "John Doe" {
		t.Errorf("ожидалось full_name 'John Doe', получено %q", emp.FullName)
	}
}

// TestService_CreateEmployee_InvalidDate_Integration тестирует невалидную дату
func TestService_CreateEmployee_InvalidDate_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	dept, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "Engineering"})

	req := model.CreateEmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
		HiredAt:  strPtr("invalid-date"),
	}

	_, err := svc.CreateEmployee(dept.ID, req)
	if err == nil {
		t.Error("ожидалась ошибка для невалидной даты")
	}
}

// TestService_GetDepartmentTree_Integration тестирует получение дерева
func TestService_GetDepartmentTree_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём иерархию
	parent, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "Parent"})
	child, _ := svc.CreateDepartment(model.CreateDepartmentRequest{
		Name:     "Child",
		ParentID: &parent.ID,
	})

	// Создаём сотрудников
	svc.CreateEmployee(parent.ID, model.CreateEmployeeRequest{
		FullName: "Parent Employee",
		Position: "Manager",
	})

	svc.CreateEmployee(child.ID, model.CreateEmployeeRequest{
		FullName: "Child Employee",
		Position: "Developer",
	})

	// Получаем дерево
	tree, err := svc.GetDepartmentTree(parent.ID, 2, true)
	if err != nil {
		t.Fatalf("ошибка получения дерева: %v", err)
	}

	if len(tree.Children) != 1 {
		t.Errorf("ожидался 1 ребёнок, получено %d", len(tree.Children))
	}

	if len(tree.Employees) != 1 {
		t.Errorf("ожидался 1 сотрудник у родителя, получено %d", len(tree.Employees))
	}
}

// TestService_CycleDetection_Integration тестирует обнаружение цикла
func TestService_CycleDetection_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	// Создаём: Parent -> Child
	parent, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "Parent"})
	child, _ := svc.CreateDepartment(model.CreateDepartmentRequest{
		Name:     "Child",
		ParentID: &parent.ID,
	})

	// Пытаемся сделать Parent ребёнком Child (цикл!)
	updateReq := model.UpdateDepartmentRequest{ParentID: &child.ID}
	_, err := svc.UpdateDepartment(parent.ID, updateReq)
	if err != ErrCycleDetected {
		t.Errorf("ожидалась ошибка ErrCycleDetected, получено %v", err)
	}
}

// TestService_DepthClamping_Integration тестирует ограничение глубины
func TestService_DepthClamping_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	dept, _ := svc.CreateDepartment(model.CreateDepartmentRequest{Name: "Test"})

	// depth = 0 должен стать 1
	tree, err := svc.GetDepartmentTree(dept.ID, 0, true)
	if err != nil {
		t.Fatalf("ошибка при depth=0: %v", err)
	}
	if tree == nil {
		t.Error("ожидалось дерево")
	}

	// depth = 10 должен стать 5
	tree, err = svc.GetDepartmentTree(dept.ID, 10, true)
	if err != nil {
		t.Fatalf("ошибка при depth=10: %v", err)
	}
	if tree == nil {
		t.Error("ожидалось дерево")
	}
}

// TestService_NameTrimming_Integration тестирует обрезку пробелов
func TestService_NameTrimming_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	req := model.CreateDepartmentRequest{
		Name: "  TrimmedName  ",
	}

	dept, err := svc.CreateDepartment(req)
	if err != nil {
		t.Fatalf("ошибка создания: %v", err)
	}

	if dept.Name != "TrimmedName" {
		t.Errorf("ожидалось имя 'TrimmedName', получено %q", dept.Name)
	}
}

// BenchmarkService_CreateDepartment_Integration бенчмарк
func BenchmarkService_CreateDepartment_Integration(b *testing.B) {
	pgContainer, db, ctx := setupTestContainer(b)
	defer pgContainer.Terminate(ctx)

	repo := repository.NewRepository(db)
	svc := NewService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := model.CreateDepartmentRequest{Name: "Benchmark Dept"}
		svc.CreateDepartment(req)
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
