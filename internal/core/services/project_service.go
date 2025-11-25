package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"

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

func (s *ProjectService) CreateProject(userID, title, description string) (*domain.Project, error) {
	project := &domain.Project{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		UserID:      userID,
	}

	if err := s.projectRepo.CreateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) ListProjects(userID string) ([]domain.Project, error) {
	return s.projectRepo.GetAllProjects(userID)
}

func (s *ProjectService) UpdateProject(id, title, description string) (*domain.Project, error) {
	project, err := s.projectRepo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	project.Title = title
	project.Description = description

	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(id string) error {
	return s.projectRepo.DeleteProject(id)
}
