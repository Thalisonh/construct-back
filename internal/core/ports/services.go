package ports

import "construct-backend/internal/core/domain"

type AuthService interface {
	Signup(email, password, name string) (string, error)
	Login(email, password string) (string, error)
	LoginWithGoogle(idToken string) (string, error)
}

type ProjectService interface {
	CreateProject(userID, title, description string) (*domain.Project, error)
	ListProjects(userID string) ([]domain.Project, error)
	UpdateProject(id, title, description string) (*domain.Project, error)
	DeleteProject(id string) error
}

type LinkService interface {
	CreateLink(projectID, url, description string) (*domain.Link, error)
	ListLinks(projectID string) ([]domain.Link, error)
	DeleteLink(id string) error
}
