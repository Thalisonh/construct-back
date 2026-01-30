package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"fmt"
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

func (s *ProjectService) GetProject(id, userID string) (*domain.Project, error) {
	return s.projectRepo.GetProjectByID(id, userID)
}

func (s *ProjectService) GetPublicProject(id string) (*domain.Project, error) {
	return s.projectRepo.GetPublicProjectByID(id)
}

func (s *ProjectService) UpdateProject(id, name, clientID, address, summary, startDate string, isPublic bool, userID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetProjectByID(id, userID)
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
	project.IsPublic = isPublic

	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(id, userID string) error {
	return s.projectRepo.DeleteProject(id, userID)
}

func (s *ProjectService) AddTask(projectID, name, status, dueDate, userID string) (*domain.Task, error) {
	parsedDueDate, _ := time.Parse(time.RFC3339, dueDate)

	task := &domain.Task{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Name:      name,
		Status:    status,
		DueDate:   parsedDueDate,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	// Verify project ownership before adding task?
	// GetProjectByID check is safest, but AddTask relies on projectID.
	// Ideally we should check if projectID belongs to userID.
	// But repository.AddTask just inserts.
	// Let's rely on the handler passing a valid projectID for the user?
	// No, user can pass any projectID. We must verify ownership.
	_, err := s.projectRepo.GetProjectByID(projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied")
	}

	if err := s.projectRepo.AddTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *ProjectService) AddSubtask(taskID, name, status, userID string) (*domain.Subtask, error) {
	// Verify task ownership
	_, err := s.projectRepo.GetTaskByID(taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task not found or access denied")
	}

	subtask := &domain.Subtask{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Name:      name,
		Status:    status,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.projectRepo.AddSubtask(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *ProjectService) UpdateTask(id, userID string) (*domain.Task, error) {
	task, err := s.projectRepo.GetTaskByID(id, userID)
	if err != nil {
		return nil, err
	}

	status := task.Status
	if status == "Completed" {
		status = "Pending"
	} else {
		status = "Completed"
	}

	task.Status = status

	if err := s.projectRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	if status == "Completed" {
		if err := s.projectRepo.UpdateSubtaskByTaskID(task.ID); err != nil {
			return nil, err
		}
	}

	return task, nil
}

func (s *ProjectService) DeleteTask(id, userID string) error {
	return s.projectRepo.DeleteTask(id, userID)
}

func (s *ProjectService) DeleteSubtask(id, userID string) error {
	return s.projectRepo.DeleteSubtask(id, userID)
}

func (s *ProjectService) UpdateSubtask(id, userID string) (*domain.Subtask, error) {
	subtask, err := s.projectRepo.GetSubtaskByID(id, userID)
	if err != nil {
		return nil, err
	}

	status := subtask.Status
	if status == "Completed" {
		status = "Pending"
	} else {
		status = "Completed"
	}

	subtask.Status = status

	if err := s.projectRepo.UpdateSubtask(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *ProjectService) GetTask(id, userID string) (*domain.Task, error) {
	return s.projectRepo.GetTaskByID(id, userID)
}

func (s *ProjectService) GetSubtask(id, userID string) (*domain.Subtask, error) {
	return s.projectRepo.GetSubtaskByID(id, userID)
}

func (s *ProjectService) ListTasks(projectID string) ([]domain.Task, error) {
	tasks, err := s.projectRepo.GetTasksByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	totalTasks := len(tasks)
	if totalTasks > 0 {
		completedTasks := 0
		for _, task := range tasks {
			if task.Status == "Completed" {
				completedTasks++
			}
		}
		percentage := (float64(completedTasks) / float64(totalTasks)) * 100
		fmt.Printf("Project %s: %.2f%% tasks completed\n", projectID, percentage)
	}

	return tasks, nil
}
