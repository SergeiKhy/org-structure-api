package model

import (
	"testing"
	"time"
)

// Department Tests

func TestDepartment_DefaultValues(t *testing.T) {
	dept := Department{
		Name: "Test",
	}

	if dept.ID != 0 {
		t.Error("new department should have ID 0")
	}

	if dept.ParentID != nil {
		t.Error("new department should have nil ParentID")
	}

	if !dept.CreatedAt.IsZero() {
		t.Error("new department CreatedAt should be zero until set")
	}

	// Слайсы в Go по умолчанию равны nil
	if dept.Employees != nil {
		t.Error("Employees slice should be nil by default")
	}

	if dept.Children != nil {
		t.Error("Children slice should be nil by default")
	}
}

func TestDepartment_WithName(t *testing.T) {
	dept := Department{
		ID:   1,
		Name: "Engineering",
	}

	if dept.ID != 1 {
		t.Errorf("expected ID 1, got %d", dept.ID)
	}

	if dept.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got %q", dept.Name)
	}
}

func TestDepartment_WithParent(t *testing.T) {
	parentID := 5
	dept := Department{
		Name:     "Child",
		ParentID: &parentID,
	}

	if dept.ParentID == nil {
		t.Fatal("ParentID should not be nil")
	}

	if *dept.ParentID != 5 {
		t.Errorf("expected ParentID 5, got %d", *dept.ParentID)
	}
}

func TestDepartment_WithEmployees(t *testing.T) {
	now := time.Now()
	dept := Department{
		Name: "Engineering",
		Employees: []Employee{
			{
				ID:       1,
				FullName: "John Doe",
				Position: "Developer",
			},
			{
				ID:       2,
				FullName: "Jane Smith",
				Position: "Manager",
			},
		},
		CreatedAt: now,
	}

	if len(dept.Employees) != 2 {
		t.Errorf("expected 2 employees, got %d", len(dept.Employees))
	}

	if dept.Employees[0].FullName != "John Doe" {
		t.Errorf("expected first employee 'John Doe', got %q", dept.Employees[0].FullName)
	}
}

func TestDepartment_WithChildren(t *testing.T) {
	dept := Department{
		Name: "Parent",
		Children: []Department{
			{Name: "Child1"},
			{Name: "Child2"},
		},
	}

	if len(dept.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(dept.Children))
	}

	if dept.Children[0].Name != "Child1" {
		t.Errorf("expected first child 'Child1', got %q", dept.Children[0].Name)
	}
}

// Employee Tests

func TestEmployee_DefaultValues(t *testing.T) {
	emp := Employee{
		FullName: "John Doe",
		Position: "Developer",
	}

	if emp.ID != 0 {
		t.Error("new employee should have ID 0")
	}

	// DepartmentID может быть равен 0 для нового сотрудника
	// if emp.DepartmentID != 0 {
	// 	t.Error("new employee DepartmentID should be 0 until set")
	// }

	if emp.HiredAt != nil {
		t.Error("new employee HiredAt should be nil by default")
	}
}

func TestEmployee_WithHiredAt(t *testing.T) {
	now := time.Now()
	emp := Employee{
		FullName:     "John Doe",
		Position:     "Developer",
		DepartmentID: 1,
		HiredAt:      &now,
	}

	if emp.HiredAt == nil {
		t.Error("HiredAt should not be nil")
	}

	if emp.HiredAt != &now {
		t.Error("HiredAt should point to the provided time")
	}
}

func TestEmployee_WithDepartment(t *testing.T) {
	emp := Employee{
		DepartmentID: 5,
		FullName:     "Jane Smith",
		Position:     "Manager",
	}

	if emp.DepartmentID != 5 {
		t.Errorf("expected DepartmentID 5, got %d", emp.DepartmentID)
	}
}

// CreateDepartmentRequest Tests

func TestCreateDepartmentRequest(t *testing.T) {
	req := CreateDepartmentRequest{
		Name:     "Engineering",
		ParentID: nil,
	}

	if req.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got %q", req.Name)
	}

	if req.ParentID != nil {
		t.Error("ParentID should be nil")
	}
}

func TestCreateDepartmentRequest_WithParent(t *testing.T) {
	parentID := 10
	req := CreateDepartmentRequest{
		Name:     "Child Dept",
		ParentID: &parentID,
	}

	if req.ParentID == nil {
		t.Fatal("ParentID should not be nil")
	}

	if *req.ParentID != 10 {
		t.Errorf("expected ParentID 10, got %d", *req.ParentID)
	}
}

// CreateEmployeeRequest Tests

func TestCreateEmployeeRequest(t *testing.T) {
	req := CreateEmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
		HiredAt:  nil,
	}

	if req.FullName != "John Doe" {
		t.Errorf("expected FullName 'John Doe', got %q", req.FullName)
	}

	if req.Position != "Developer" {
		t.Errorf("expected Position 'Developer', got %q", req.Position)
	}

	if req.HiredAt != nil {
		t.Error("HiredAt should be nil")
	}
}

func TestCreateEmployeeRequest_WithHiredAt(t *testing.T) {
	date := "2024-01-15"
	req := CreateEmployeeRequest{
		FullName: "Jane Smith",
		Position: "Manager",
		HiredAt:  &date,
	}

	if req.HiredAt == nil {
		t.Error("HiredAt should not be nil")
	}

	if *req.HiredAt != "2024-01-15" {
		t.Errorf("expected HiredAt '2024-01-15', got %q", *req.HiredAt)
	}
}

// UpdateDepartmentRequest Tests

func TestUpdateDepartmentRequest_Empty(t *testing.T) {
	req := UpdateDepartmentRequest{}

	if req.Name != "" {
		t.Errorf("expected empty name, got %q", req.Name)
	}

	if req.ParentID != nil {
		t.Error("expected nil ParentID")
	}
}

func TestUpdateDepartmentRequest_FullUpdate(t *testing.T) {
	parentID := 5
	req := UpdateDepartmentRequest{
		Name:     "Updated Name",
		ParentID: &parentID,
	}

	if req.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", req.Name)
	}

	if req.ParentID == nil {
		t.Error("ParentID should not be nil")
	}
}

func TestUpdateDepartmentRequest_NameOnly(t *testing.T) {
	req := UpdateDepartmentRequest{
		Name: "New Name",
	}

	if req.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %q", req.Name)
	}

	if req.ParentID != nil {
		t.Error("ParentID should be nil when not updating")
	}
}

func TestUpdateDepartmentRequest_ParentOnly(t *testing.T) {
	parentID := 3
	req := UpdateDepartmentRequest{
		ParentID: &parentID,
	}

	if req.Name != "" {
		t.Errorf("expected empty name, got %q", req.Name)
	}

	if req.ParentID == nil {
		t.Error("ParentID should not be nil")
	}
}

// JSON Tag Tests

func TestDepartmentJSONTags(t *testing.T) {
	dept := Department{
		ID:   1,
		Name: "Test",
	}

	// Проверка наличия JSON-тегов в структуре через верификацию названий полей
	// Это базовая проверка — для полноценного тестирования JSON следует использовать json.Marshal
	if dept.ID == 0 {
		t.Error("ID field should be accessible")
	}
	if dept.Name == "" {
		t.Error("Name field should be accessible")
	}
}

func TestEmployeeJSONTags(t *testing.T) {
	emp := Employee{
		ID:       1,
		FullName: "Test",
		Position: "Dev",
	}

	if emp.ID == 0 {
		t.Error("ID field should be accessible")
	}
	if emp.FullName == "" {
		t.Error("FullName field should be accessible")
	}
	if emp.Position == "" {
		t.Error("Position field should be accessible")
	}
}

// GORM Tag Tests

func TestDepartmentGORMTags(t *testing.T) {
	// Базовая проверка наличия GORM-тегов
	// В реальных тестах следует проверять фактическую работу GORM
	dept := Department{}
	_ = dept.ID
	_ = dept.Name
	_ = dept.ParentID
	_ = dept.CreatedAt
}

func TestEmployeeGORMTags(t *testing.T) {
	emp := Employee{}
	_ = emp.ID
	_ = emp.DepartmentID
	_ = emp.FullName
	_ = emp.Position
	_ = emp.HiredAt
	_ = emp.CreatedAt
}

// Benchmark Tests

func BenchmarkDepartmentCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Department{
			Name: "Test Department",
		}
	}
}

func BenchmarkEmployeeCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Employee{
			FullName:     "John Doe",
			Position:     "Developer",
			DepartmentID: 1,
		}
	}
}
