package ports

import "construct-backend/internal/core/domain"

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(id string) (*domain.User, error)
}

type ProjectRepository interface {
	CreateProject(project *domain.Project) error
	GetAllProjects(userID string) ([]domain.Project, error)
	GetProjectByID(id string) (*domain.Project, error)
	UpdateProject(project *domain.Project) error
	DeleteProject(id string) error
}

type LinkRepository interface {
	CreateLink(link *domain.Link) error
	GetAllLinks(projectID string) ([]domain.Link, error)
	DeleteLink(id string) error
}
