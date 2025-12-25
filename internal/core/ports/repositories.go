package ports

import "construct-backend/internal/core/domain"

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(id string) (*domain.User, error)
	VerifyUserName(username string) (*domain.UsernameVerification, error)
	UpdateUsername(userID, username string) error
	GetUsername(userID string) (string, error)
	GetPublicProfile(username string) (*domain.PublicProfile, error)
}

type ProjectRepository interface {
	CreateProject(project *domain.Project) error
	GetAllProjects(userID string) ([]domain.Project, error)
	GetProjectByID(id string) (*domain.Project, error)
	UpdateProject(project *domain.Project) error
	DeleteProject(id string) error
	AddTask(task *domain.Task) error
	AddSubtask(subtask *domain.Subtask) error
	UpdateTask(task *domain.Task) error
	UpdateSubtask(subtask *domain.Subtask) error
	DeleteTask(id string) error
	DeleteSubtask(id string) error
	GetTaskByID(id string) (*domain.Task, error)
	GetSubtaskByID(id string) (*domain.Subtask, error)
	GetTasksByProjectID(projectID string) ([]domain.Task, error)
}

type LinkRepository interface {
	CreateLink(link *domain.Link) error
	GetAllLinks(userID string) ([]domain.Link, error)
	UpdateLink(link *domain.Link) error
	DeleteLink(id string) error
}

type ClientRepository interface {
	CreateClient(client *domain.Client) error
	GetClientByID(id string) (*domain.Client, error)
	GetAllClients() ([]domain.Client, error)
	UpdateClient(client *domain.Client) error
	DeleteClient(id string) error
	AddComment(comment *domain.Comment) error
}
