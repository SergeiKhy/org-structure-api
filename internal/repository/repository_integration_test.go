//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/SergeiKhy/org-structure-api/internal/model"
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

	// Запускаем PostgreSQL контейнер
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

	// Получаем connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("ошибка получения connection string: %v", err)
	}

	// Подключаемся через GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка подключения к БД: %v", err)
	}

	// Создаём таблицы
	err = db.AutoMigrate(&model.Department{}, &model.Employee{})
	if err != nil {
		t.Fatalf("ошибка миграции: %v", err)
	}

	return pgContainer, db, ctx
}

// TestRepository_CreateDepartment_Integration тестирует создание подразделения
func TestRepository_CreateDepartment_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Тест создания без родителя
	dept := &model.Department{
		Name:      "Engineering",
		ParentID:  nil,
		CreatedAt: time.Now(),
	}

	err := repo.CreateDepartment(dept)
	if err != nil {
		t.Fatalf("ошибка создания подразделения: %v", err)
	}

	if dept.ID == 0 {
		t.Error("ожидался ненулевой ID после создания")
	}

	// Проверяем что сохранилось
	saved, err := repo.GetDepartmentByID(dept.ID)
	if err != nil {
		t.Fatalf("ошибка получения: %v", err)
	}

	if saved.Name != "Engineering" {
		t.Errorf("ожидалось имя 'Engineering', получено %q", saved.Name)
	}
}

// TestRepository_CreateDepartment_WithParent тестирует создание с родителем
func TestRepository_CreateDepartment_WithParent_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Создаём родителя
	parent := &model.Department{Name: "Company"}
	repo.CreateDepartment(parent)

	// Создаём ребёнка
	child := &model.Department{
		Name:      "Engineering",
		ParentID:  &parent.ID,
		CreatedAt: time.Now(),
	}

	err := repo.CreateDepartment(child)
	if err != nil {
		t.Fatalf("ошибка создания подразделения: %v", err)
	}

	// Проверяем parent_id
	saved, _ := repo.GetDepartmentByID(child.ID)
	if saved.ParentID == nil {
		t.Fatal("ожидался ненулевой parent_id")
	}
	if *saved.ParentID != parent.ID {
		t.Errorf("ожидался parent_id %d, получен %d", parent.ID, *saved.ParentID)
	}
}

// TestRepository_CheckUniqueName_Integration тестирует проверку уникальности
func TestRepository_CheckUniqueName_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Создаём подразделение
	repo.CreateDepartment(&model.Department{Name: "Engineering"})

	// Проверяем дубликат
	isUnique, err := repo.CheckUniqueName(nil, "Engineering", 0)
	if err != nil {
		t.Fatalf("ошибка проверки: %v", err)
	}
	if isUnique {
		t.Error("ожидалось false для дубликата")
	}

	// Проверяем уникальное имя
	isUnique, _ = repo.CheckUniqueName(nil, "Marketing", 0)
	if !isUnique {
		t.Error("ожидалось true для уникального имени")
	}
}

// TestRepository_GetParentChain_Integration тестирует получение цепочки родителей
func TestRepository_GetParentChain_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Создаём иерархию: Grandparent -> Parent -> Child
	grandparent := &model.Department{Name: "Grandparent"}
	repo.CreateDepartment(grandparent)

	parent := &model.Department{
		Name:     "Parent",
		ParentID: &grandparent.ID,
	}
	repo.CreateDepartment(parent)

	child := &model.Department{
		Name:     "Child",
		ParentID: &parent.ID,
	}
	repo.CreateDepartment(child)

	// Получаем цепочку
	chain, err := repo.GetParentChain(child.ID)
	if err != nil {
		t.Fatalf("ошибка получения цепочки: %v", err)
	}

	if len(chain) != 2 {
		t.Fatalf("ожидалось 2 родителя, получено %d", len(chain))
	}

	if chain[0] != parent.ID || chain[1] != grandparent.ID {
		t.Errorf("неверная цепочка: %v", chain)
	}
}

// TestRepository_GetChildrenIDs_Integration тестирует получение дочерних элементов
func TestRepository_GetChildrenIDs_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Создаём родителя и детей
	parent := &model.Department{Name: "Parent"}
	repo.CreateDepartment(parent)

	child1 := &model.Department{Name: "Child1", ParentID: &parent.ID}
	repo.CreateDepartment(child1)

	child2 := &model.Department{Name: "Child2", ParentID: &parent.ID}
	repo.CreateDepartment(child2)

	// Получаем детей
	children, err := repo.GetChildrenIDs(parent.ID)
	if err != nil {
		t.Fatalf("ошибка получения детей: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("ожидалось 2 детей, получено %d", len(children))
	}
}

// TestRepository_ReassignEmployees_Integration тестирует переназначение сотрудников
func TestRepository_ReassignEmployees_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Создаём подразделения
	oldDept := &model.Department{Name: "OldDept"}
	repo.CreateDepartment(oldDept)

	newDept := &model.Department{Name: "NewDept"}
	repo.CreateDepartment(newDept)

	// Создаём сотрудника
	emp := &model.Employee{
		DepartmentID: oldDept.ID,
		FullName:     "John Doe",
		Position:     "Developer",
	}
	repo.CreateEmployee(emp)

	// Переназначаем
	err := repo.ReassignEmployees(oldDept.ID, newDept.ID)
	if err != nil {
		t.Fatalf("ошибка переназначения: %v", err)
	}

	// Проверяем
	var updated model.Employee
	db.First(&updated, emp.ID)
	if updated.DepartmentID != newDept.ID {
		t.Errorf("ожидался department_id %d, получен %d", newDept.ID, updated.DepartmentID)
	}
}

// TestRepository_DeleteDepartment_Cascade_Integration тестирует каскадное удаление
func TestRepository_DeleteDepartment_Cascade_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Создаём родителя с ребёнком
	parent := &model.Department{Name: "Parent"}
	repo.CreateDepartment(parent)

	child := &model.Department{Name: "Child", ParentID: &parent.ID}
	repo.CreateDepartment(child)

	// Создаём сотрудника у родителя
	emp := &model.Employee{
		DepartmentID: parent.ID,
		FullName:     "John Doe",
		Position:     "Dev",
	}
	repo.CreateEmployee(emp)

	// Удаляем родителя
	err := repo.DeleteDepartment(parent.ID)
	if err != nil {
		t.Fatalf("ошибка удаления: %v", err)
	}

	// Проверяем что удалилось
	_, err = repo.GetDepartmentByID(parent.ID)
	if err == nil {
		t.Error("ожидалась ошибка для удаленного подразделения")
	}

	// Ребёнок тоже должен удалиться (CASCADE)
	_, err = repo.GetDepartmentByID(child.ID)
	if err == nil {
		t.Error("ожидалась ошибка для удаленного дочернего подразделения")
	}
}

// TestRepository_Transaction_Integration тестирует транзакции
func TestRepository_Transaction_Integration(t *testing.T) {
	pgContainer, db, ctx := setupTestContainer(t)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	// Начинаем транзакцию
	tx := repo.DB().Begin()
	txRepo := repo.WithTx(tx)

	// Создаём в транзакции
	dept := &model.Department{Name: "In Transaction"}
	err := txRepo.CreateDepartment(dept)
	if err != nil {
		tx.Rollback()
		t.Fatalf("ошибка создания: %v", err)
	}

	// Проверяем внутри транзакции
	_, err = txRepo.GetDepartmentByID(dept.ID)
	if err != nil {
		tx.Rollback()
		t.Fatal("не найдено внутри транзакции")
	}

	// Коммитим
	tx.Commit()

	// Проверяем после коммита
	_, err = repo.GetDepartmentByID(dept.ID)
	if err != nil {
		t.Error("не найдено после коммита")
	}
}

// BenchmarkRepository_CreateDepartment_Integration бенчмарк создания
func BenchmarkRepository_CreateDepartment_Integration(b *testing.B) {
	pgContainer, db, ctx := setupTestContainer(b)
	defer pgContainer.Terminate(ctx)

	repo := NewRepository(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dept := &model.Department{Name: "Benchmark Dept"}
		repo.CreateDepartment(dept)
	}
}
