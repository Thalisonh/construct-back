package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"fmt"
	"strings"
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

func (s *ProjectService) CreateProject(companyID, userID, name, clientID, address, summary string, startDate string) (*domain.Project, error) {
	parsedStartDate, errParse := time.Parse(time.RFC3339, startDate)
	if errParse != nil {
		parsedStartDate, _ = time.Parse("2006-01-02", startDate)
	}

	project := &domain.Project{
		ID:        uuid.New().String(),
		Name:      name,
		ClientID:  clientID,
		Address:   address,
		Summary:   summary,
		StartDate: parsedStartDate,
		UserID:    userID,
		CompanyID: companyID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.projectRepo.CreateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) ListProjects(companyID string) ([]domain.Project, error) {
	return s.projectRepo.GetAllProjects(companyID)
}

func (s *ProjectService) ListProjectsByClient(clientID, companyID string) ([]domain.Project, error) {
	return s.projectRepo.GetProjectsByClientID(clientID, companyID)
}

func (s *ProjectService) GetProject(id, companyID string) (*domain.Project, error) {
	return s.projectRepo.GetProjectByID(id, companyID)
}

func (s *ProjectService) GetPublicProject(id, pin string) (*domain.Project, error) {
	project, err := s.projectRepo.GetPublicProjectByID(id)
	if err != nil {
		return nil, err
	}

	if err := validatePublicProjectPin(project, pin); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) VerifyPublicProjectPin(id, pin string) error {
	project, err := s.projectRepo.GetPublicProjectByID(id)
	if err != nil {
		return fmt.Errorf("invalid public project access")
	}

	return validatePublicProjectPin(project, pin)
}

func (s *ProjectService) UpdateProject(id, name, clientID, address, summary, startDate string, isPublic bool, companyID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetProjectByID(id, companyID)
	if err != nil {
		return nil, err
	}

	parsedStartDate, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		parsedStartDate, _ = time.Parse("2006-01-02", startDate)
	}

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

func (s *ProjectService) DeleteProject(id, companyID string) error {
	return s.projectRepo.DeleteProject(id, companyID)
}

func (s *ProjectService) AddTask(projectID, name, status, dueDate, companyID, userID string) (*domain.Task, error) {
	parsedDueDate, errParse := time.Parse(time.RFC3339, dueDate)
	if errParse != nil {
		parsedDueDate, _ = time.Parse("2006-01-02", dueDate)
	}

	task := &domain.Task{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Name:      name,
		Status:    status,
		DueDate:   parsedDueDate,
		UserID:    userID,
		CompanyID: companyID,
		CreatedAt: time.Now(),
	}

	// Verify project ownership before adding task
	_, err := s.projectRepo.GetProjectByID(projectID, companyID)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied")
	}

	if err := s.projectRepo.AddTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *ProjectService) AddSubtask(taskID, name, status, companyID, userID string) (*domain.Subtask, error) {
	// Verify task ownership
	_, err := s.projectRepo.GetTaskByID(taskID, companyID)
	if err != nil {
		return nil, fmt.Errorf("task not found or access denied")
	}

	subtask := &domain.Subtask{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Name:      name,
		Status:    status,
		UserID:    userID,
		CompanyID: companyID,
		CreatedAt: time.Now(),
	}

	if err := s.projectRepo.AddSubtask(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *ProjectService) UpdateTask(id, companyID, status string) (*domain.Task, error) {
	task, err := s.projectRepo.GetTaskByID(id, companyID)
	if err != nil {
		return nil, err
	}

	if status != "" {
		task.Status = status
	} else {
		// Fallback to toggle if no status provided (for backward compatibility if needed)
		if task.Status == "Completed" {
			task.Status = "Pending"
		} else {
			task.Status = "Completed"
		}
	}

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

func (s *ProjectService) DeleteTask(id, companyID string) error {
	return s.projectRepo.DeleteTask(id, companyID)
}

func (s *ProjectService) DeleteSubtask(id, companyID string) error {
	return s.projectRepo.DeleteSubtask(id, companyID)
}

func (s *ProjectService) UpdateSubtask(id, companyID string) (*domain.Subtask, error) {
	subtask, err := s.projectRepo.GetSubtaskByID(id, companyID)
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

func (s *ProjectService) GetTask(id, companyID string) (*domain.Task, error) {
	return s.projectRepo.GetTaskByID(id, companyID)
}

func (s *ProjectService) GetSubtask(id, companyID string) (*domain.Subtask, error) {
	return s.projectRepo.GetSubtaskByID(id, companyID)
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

func (s *ProjectService) CreateDiaryEntry(projectID, companyID, userID, entryDate, title string, items []domain.DiaryItem) (*domain.DiaryEntry, error) {
	_, err := s.projectRepo.GetProjectByID(projectID, companyID)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied")
	}

	parsedEntryDate, err := time.Parse("2006-01-02", entryDate)
	if err != nil {
		parsedEntryDate, err = time.Parse(time.RFC3339, entryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid entry date")
		}
	}

	now := time.Now()
	entry := &domain.DiaryEntry{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    userID,
		CompanyID: companyID,
		EntryDate: parsedEntryDate,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for index, item := range items {
		entry.Items = append(entry.Items, domain.DiaryItem{
			ID:           uuid.New().String(),
			DiaryEntryID: entry.ID,
			Type:         item.Type,
			Label:        item.Label,
			Content:      item.Content,
			Visibility:   item.Visibility,
			SortOrder:    index,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	if err := s.projectRepo.CreateDiaryEntry(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *ProjectService) ListDiaryEntries(projectID, companyID string) ([]domain.DiaryEntry, error) {
	_, err := s.projectRepo.GetProjectByID(projectID, companyID)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied")
	}

	return s.projectRepo.GetDiaryEntriesByProject(projectID, companyID)
}

func (s *ProjectService) ListPublicDiaryEntries(projectID, pin string) ([]domain.DiaryEntry, error) {
	project, err := s.projectRepo.GetPublicProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or not public")
	}

	if err := validatePublicProjectPin(project, pin); err != nil {
		return nil, err
	}

	return s.projectRepo.GetPublicDiaryEntriesByProject(projectID)
}

func (s *ProjectService) UpdateDiaryEntry(entryID, projectID, companyID, entryDate, title string, items []domain.DiaryItem) (*domain.DiaryEntry, error) {
	entry, err := s.projectRepo.GetDiaryEntryByID(entryID, projectID, companyID)
	if err != nil {
		return nil, fmt.Errorf("diary entry not found")
	}

	parsedEntryDate, err := time.Parse("2006-01-02", entryDate)
	if err != nil {
		parsedEntryDate, err = time.Parse(time.RFC3339, entryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid entry date")
		}
	}

	now := time.Now()
	entry.EntryDate = parsedEntryDate
	entry.Title = title
	entry.UpdatedAt = now
	entry.Items = nil

	for index, item := range items {
		entry.Items = append(entry.Items, domain.DiaryItem{
			ID:           uuid.New().String(),
			DiaryEntryID: entry.ID,
			Type:         item.Type,
			Label:        item.Label,
			Content:      item.Content,
			Visibility:   item.Visibility,
			SortOrder:    index,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	if err := s.projectRepo.UpdateDiaryEntry(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *ProjectService) DeleteDiaryEntry(entryID, projectID, companyID string) error {
	_, err := s.projectRepo.GetDiaryEntryByID(entryID, projectID, companyID)
	if err != nil {
		return fmt.Errorf("diary entry not found")
	}

	return s.projectRepo.DeleteDiaryEntry(entryID, projectID, companyID)
}

func validatePublicProjectPin(project *domain.Project, pin string) error {
	if len(pin) != 4 {
		return fmt.Errorf("invalid public project access")
	}

	for _, digit := range pin {
		if digit < '0' || digit > '9' {
			return fmt.Errorf("invalid public project access")
		}
	}

	if project == nil || project.Client == nil {
		return fmt.Errorf("invalid public project access")
	}

	digits := onlyDigits(project.Client.Phone)
	if len(digits) < 4 {
		return fmt.Errorf("invalid public project access")
	}

	if pin != digits[len(digits)-4:] {
		return fmt.Errorf("invalid public project access")
	}

	return nil
}

func onlyDigits(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}
