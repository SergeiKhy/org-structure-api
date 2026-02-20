package service

import (
	"errors"
	"strings"
	"time"

	"github.com/SergeiKhy/org-structure-api/internal/model"
	"github.com/SergeiKhy/org-structure-api/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrCycleDetected = errors.New("cycle detected")
	ErrDuplicateName = errors.New("duplicate name within parent")
	ErrSelfParent    = errors.New("cannot be parent of itself")
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// Валидация имени
func validateName(name string) string {
	return strings.TrimSpace(name)
}

func (s *Service) CreateDepartment(req model.CreateDepartmentRequest) (*model.Department, error) {
	name := validateName(req.Name)
	if name == "" || len(name) > 200 {
		return nil, errors.New("invalid name")
	}

	// Проверка уникальности
	ok, err := s.repo.CheckUniqueName(req.ParentID, name, 0)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrDuplicateName
	}

	// Проверка на цикл, если parent_id указан
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			// Обработка кейса, когда фронт может прислать 0 вместо nil
			req.ParentID = nil
		} else {
			// Проверка существования родителя
			_, err = s.repo.GetDepartmentByID(*req.ParentID)
			if err != nil {
				return nil, ErrNotFound
			}
		}
	}

	dept := &model.Department{
		Name:      name,
		ParentID:  req.ParentID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateDepartment(dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *Service) UpdateDepartment(id int, req model.UpdateDepartmentRequest) (*model.Department, error) {
	dept, err := s.repo.GetDepartmentByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	if req.Name != "" {
		name := validateName(req.Name)
		if len(name) > 200 {
			return nil, errors.New("invalid name")
		}
		// Проверка уникальности нового имени
		ok, err := s.repo.CheckUniqueName(req.ParentID, name, id)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrDuplicateName
		}
		dept.Name = name
	}

	if req.ParentID != nil {
		if *req.ParentID == 0 {
			req.ParentID = nil
		} else {
			// Нельзя сделать родителем самого себя
			if *req.ParentID == id {
				return nil, ErrSelfParent
			}
			// Проверка на цикл (новый родитель не должен быть потомком текущего)
			parents, err := s.repo.GetParentChain(*req.ParentID)
			if err != nil {
				return nil, err
			}
			for _, pID := range parents {
				if pID == id {
					return nil, ErrCycleDetected
				}
			}
			// Проверка сущестования родителя
			_, err = s.repo.GetDepartmentByID(*req.ParentID)
			if err != nil {
				return nil, ErrNotFound
			}
		}
		dept.ParentID = req.ParentID
	}
	if err := s.repo.UpdateDepartment(dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *Service) DeleteDepartment(id int, mode string, reassignToID *int) error {
	// Проверка на существование
	_, err := s.repo.GetDepartmentByID(id)
	if err != nil {
		return ErrNotFound
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// Создание репозитория с транзакционной БД
		txRepo := s.repo.WithTx(tx)

		if mode == "reassign" {
			if reassignToID == nil {
				return errors.New("reassign_to_department_id is required")
			}
			// Проверка существования целевого департамента
			_, err := txRepo.GetDepartmentByID(*reassignToID)
			if err != nil {
				return ErrNotFound
			}
			// Перевод сотрудников
			if err := txRepo.ReassignEmployees(id, *reassignToID); err != nil {
				return err
			}
		} else {
			// Удаляем дочерние департаменты рекурсивно
			childrenIDs, err := txRepo.GetChildrenIDs(id)
			if err != nil {
				return err
			}
			// Удаляем детей
			if err := txRepo.DeleteChildrenIDs(tx, childrenIDs); err != nil {
				return err
			}
		}

		// удаляем департамент
		return txRepo.DeleteDepartment(id)
	})
}

func (s *Service) CreateEmployee(deptID int, req model.CreateEmployeeRequest) (*model.Employee, error) {
	// Проверка существования департамента
	_, err := s.repo.GetDepartmentByID(deptID)
	if err != nil {
		return nil, ErrNotFound
	}

	fullName := validateName(req.FullName)
	position := validateName(req.Position)

	if fullName == "" || position == "" || len(fullName) > 200 || len(position) > 200 {
		return nil, errors.New("invalid fields")
	}

	var hiredAt *time.Time
	if req.HiredAt != nil {
		t, err := time.Parse("2006-01-02", *req.HiredAt)
		if err != nil {
			return nil, errors.New("invalid date format")
		}
		hiredAt = &t
	}

	emp := &model.Employee{
		DepartmentID: deptID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateEmployee(emp); err != nil {
		return nil, err
	}
	return emp, nil
}

// Рекурсивное построение дерева
func (s *Service) GetDepartmentTree(id int, depth int, includeEmployees bool) (*model.Department, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}

	dept, err := s.repo.GetDepartmentWithChildren(id, depth)
	if err != nil {
		return nil, ErrNotFound
	}

	if includeEmployees {
		emps, err := s.repo.GetEmployeesByDeptID(id)
		if err != nil {
			return nil, err
		}
		dept.Employees = emps
	}

	if depth > 1 {
		// Загружаем детей
		var children []model.Department
		if err := s.repo.DB().Where("parent_id = ?", id).Find(&children).Error; err != nil {
			return nil, err
		}

		for i := range children {
			childTree, err := s.GetDepartmentTree(children[i].ID, depth-1, includeEmployees)
			if err != nil {
				continue
			}
			dept.Children = append(dept.Children, *childTree)
		}
	}
	return dept, nil
}
