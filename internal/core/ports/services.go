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
	GetProject(id, userID string) (*domain.Project, error)
	GetPublicProject(id string) (*domain.Project, error)
	UpdateProject(id, name, clientID, address, summary, startDate string, isPublic bool, userID string) (*domain.Project, error)
	DeleteProject(id, userID string) error
	AddTask(projectID, name, status, dueDate, userID string) (*domain.Task, error)
	AddSubtask(taskID, name, status, userID string) (*domain.Subtask, error)
	UpdateTask(id, userID string) (*domain.Task, error)
	UpdateSubtask(id, userID string) (*domain.Subtask, error)
	DeleteTask(id, userID string) error
	DeleteSubtask(id, userID string) error
	GetTask(id, userID string) (*domain.Task, error)
	GetSubtask(id, userID string) (*domain.Subtask, error)
	ListTasks(projectID string) ([]domain.Task, error)
}

type LinkService interface {
	CreateLink(userID, url, description string) (*domain.Link, error)
	UpdateLink(userID, url, description, id string) (*domain.Link, error)
	ListLinks(userID string) ([]domain.Link, error)
	DeleteLink(id, userID string) error
	TrackLinkClick(id string) error
}

type UserService interface {
	VerifyUserName(username string) error
	UpdateUsername(userID, username string) error
	GetUsername(userID string) (string, error)
	GetPublicProfile(username string) (*domain.PublicProfile, error)
	UpdateBio(userID, bio string) error
	GetProfile(userID string) (*domain.User, error)
}

type ClientService interface {
	CreateClient(userID, name, phone, address, summary string) (*domain.Client, error)
	GetClient(id, userID string) (*domain.Client, error)
	ListClients(userID string) ([]domain.Client, error)
	UpdateClient(id, name, phone, address, summary, userID string) (*domain.Client, error)
	DeleteClient(id string) error
	AddComment(clientID, content string) (*domain.Comment, error)
}
