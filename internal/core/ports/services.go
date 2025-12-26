package ports

import "construct-backend/internal/core/domain"

type AuthService interface {
	Signup(email, password, name string) (string, error)
	Login(email, password string) (string, error)
	LoginWithGoogle(idToken string) (string, error)
	VerifyToken(token string) error
}

type ProjectService interface {
	CreateProject(userID, name, clientID, address, summary string, startDate string) (*domain.Project, error)
	ListProjects(userID string) ([]domain.Project, error)
	GetProject(id string) (*domain.Project, error)
	UpdateProject(id, name, clientID, address, summary string, startDate string) (*domain.Project, error)
	DeleteProject(id string) error
	AddTask(projectID, name, status string, dueDate string) (*domain.Task, error)
	AddSubtask(taskID, name, status string) (*domain.Subtask, error)
	UpdateTask(id string) (*domain.Task, error)
	UpdateSubtask(id string) (*domain.Subtask, error)
	DeleteTask(id string) error
	DeleteSubtask(id string) error
	GetTask(id string) (*domain.Task, error)
	GetSubtask(id string) (*domain.Subtask, error)
	ListTasks(projectID string) ([]domain.Task, error)
}

type LinkService interface {
	CreateLink(userID, url, description string) (*domain.Link, error)
	UpdateLink(id, url, description string) (*domain.Link, error)
	ListLinks(userID string) ([]domain.Link, error)
	DeleteLink(id string) error
}

type UserService interface {
	VerifyUserName(username string) error
	UpdateUsername(userID, username string) error
	GetUsername(userID string) (string, error)
	GetPublicProfile(username string) (*domain.PublicProfile, error)
}

type ClientService interface {
	CreateClient(name, phone, address, summary string) (*domain.Client, error)
	GetClient(id string) (*domain.Client, error)
	ListClients() ([]domain.Client, error)
	UpdateClient(id, name, phone, address, summary string) (*domain.Client, error)
	DeleteClient(id string) error
	AddComment(clientID, content string) (*domain.Comment, error)
}
