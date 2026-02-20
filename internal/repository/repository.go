package repository

import (
	"github.com/SergeiKhy/org-structure-api/internal/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB возвращает базовый экземпляр gorm.DB для транзакций
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// WithTx создает новый экземпляр Repository с использованием транзакционной БД
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// Department Methods
func (r *Repository) CreateDepartment(dept *model.Department) error {
	return r.db.Create(dept).Error
}

func (r *Repository) GetDepartmentByID(id int) (*model.Department, error) {
	var dept model.Department
	err := r.db.First(&dept, id).Error
	return &dept, err
}

func (r *Repository) UpdateDepartment(dept *model.Department) error {
	return r.db.Save(dept).Error
}

func (r *Repository) DeleteDepartment(id int) error {
	return r.db.Delete(&model.Department{}, id).Error
}

func (r *Repository) CheckUniqueName(parentID *int, name string, excludeID int) (bool, error) {
	var count int64
	query := r.db.Model(&model.Department{}).Where("name = ? AND id != ?", name, excludeID)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.Count(&count).Error
	return count == 0, err
}

func (r *Repository) GetParentChain(id int) ([]int, error) {
	var parents []int
	currentID := id
	for {
		var dept model.Department
		if err := r.db.Select("parent_id").First(&dept, currentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return nil, err
		}
		if dept.ParentID == nil {
			break
		}
		parents = append(parents, *dept.ParentID)
		currentID = *dept.ParentID
	}
	return parents, nil
}

func (r *Repository) GetChildrenIDs(id int) ([]int, error) {
	var ids []int
	// Рекурсивный запрос для плоского списка детей
	// Выбираем всех, у кого parent_id = id
	var depts []model.Department
	if err := r.db.Where("parent_id = ?", id).Find(&depts).Error; err != nil {
		return nil, err
	}
	for _, d := range depts {
		ids = append(ids, d.ID)
		childIDs, err := r.GetChildrenIDs(d.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, childIDs...)
	}
	return ids, nil
}

func (r *Repository) ReassignDepartments(oldParentID int, newParentID int) error {
	return r.db.Model(&model.Department{}).Where("parent_id = ?", oldParentID).Update("parent_id", newParentID).Error
}

func (r *Repository) ReassignEmployees(oldDeptID int, newDeptID int) error {
	return r.db.Model(&model.Employee{}).Where("department_id = ?", oldDeptID).Update("department_id", newDeptID).Error
}

// Employee Methods
func (r *Repository) CreateEmployee(emp *model.Employee) error {
	return r.db.Create(emp).Error
}

func (r *Repository) GetEmployeesByDeptID(deptID int) ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Where("department_id = ?", deptID).Order("created_at ASC").Find(&employees).Error
	return employees, err
}

func (r *Repository) GetDepartmentWithChildren(id int, depth int) (*model.Department, error) {
	var dept model.Department
	if err := r.db.First(&dept, id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

// DeleteChildrenIDs удаляет подразделения по их ID в рамках транзакции
func (r *Repository) DeleteChildrenIDs(tx *gorm.DB, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Where("id IN ?", ids).Delete(&model.Department{}).Error
}
