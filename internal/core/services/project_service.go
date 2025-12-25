package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"time"

	"github.com/google/uuid"
)

type ProjectService struct {
	projectRepo ports.ProjectRepository
}

func NewProjectService(projectRepo ports.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
	}
}

func (s *ProjectService) CreateProject(userID, name, clientID, address, summary string, startDate string) (*domain.Project, error) {
	parsedStartDate, _ := time.Parse(time.RFC3339, startDate)

	project := &domain.Project{
		ID:        uuid.New().String(),
		Name:      name,
		ClientID:  clientID,
		Address:   address,
		Summary:   summary,
		StartDate: parsedStartDate,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.projectRepo.CreateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) ListProjects(userID string) ([]domain.Project, error) {
	return s.projectRepo.GetAllProjects(userID)
}

func (s *ProjectService) GetProject(id string) (*domain.Project, error) {
	return s.projectRepo.GetProjectByID(id)
}

func (s *ProjectService) UpdateProject(id, name, clientID, address, summary string, startDate string) (*domain.Project, error) {
	project, err := s.projectRepo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	parsedStartDate, _ := time.Parse(time.RFC3339, startDate)

	project.Name = name
	project.ClientID = clientID
	project.Address = address
	project.Summary = summary
	project.StartDate = parsedStartDate
	project.UpdatedAt = time.Now()

	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(id string) error {
	return s.projectRepo.DeleteProject(id)
}

func (s *ProjectService) AddTask(projectID, name, status string, dueDate string) (*domain.Task, error) {
	parsedDueDate, _ := time.Parse(time.RFC3339, dueDate)

	task := &domain.Task{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Name:      name,
		Status:    status,
		DueDate:   parsedDueDate,
		CreatedAt: time.Now(),
	}

	if err := s.projectRepo.AddTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *ProjectService) AddSubtask(taskID, name, status string) (*domain.Subtask, error) {
	subtask := &domain.Subtask{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Name:      name,
		Status:    status,
		CreatedAt: time.Now(),
	}

	if err := s.projectRepo.AddSubtask(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *ProjectService) UpdateTask(id, name, status string, dueDate string) (*domain.Task, error) {
	task, err := s.projectRepo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	parsedDueDate, _ := time.Parse(time.RFC3339, dueDate)

	task.Name = name
	task.Status = status
	task.DueDate = parsedDueDate

	if err := s.projectRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *ProjectService) DeleteTask(id string) error {
	return s.projectRepo.DeleteTask(id)
}

func (s *ProjectService) DeleteSubtask(id string) error {
	return s.projectRepo.DeleteSubtask(id)
}

func (s *ProjectService) UpdateSubtask(id, name, status string) (*domain.Subtask, error) {
	subtask, err := s.projectRepo.GetSubtaskByID(id)
	if err != nil {
		return nil, err
	}

	subtask.Name = name
	subtask.Status = status

	return subtask, nil
}

func (s *ProjectService) GetTask(id string) (*domain.Task, error) {
	return s.projectRepo.GetTaskByID(id)
}

func (s *ProjectService) GetSubtask(id string) (*domain.Subtask, error) {
	return s.projectRepo.GetSubtaskByID(id)
}

func (s *ProjectService) ListTasks(projectID string) ([]domain.Task, error) {
	return s.projectRepo.GetTasksByProjectID(projectID)
}
