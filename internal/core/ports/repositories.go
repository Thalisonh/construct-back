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
	UpdateBio(userID, bio string) error
	UpdateProfile(user *domain.User) error
	UpdatePassword(userID, password string) error
	ListUsersByCompanyID(companyID string) ([]domain.User, error)
}

type ProjectRepository interface {
	CreateProject(project *domain.Project) error
	GetAllProjects(companyID string) ([]domain.Project, error)
	GetProjectByID(id, companyID string) (*domain.Project, error)
	GetPublicProjectByID(id string) (*domain.Project, error)
	UpdateProject(project *domain.Project) error
	DeleteProject(id, companyID string) error
	AddTask(task *domain.Task) error
	AddSubtask(subtask *domain.Subtask) error
	UpdateTask(task *domain.Task) error
	UpdateSubtask(subtask *domain.Subtask) error
	UpdateSubtaskByTaskID(taskID string) error
	DeleteTask(id, companyID string) error
	DeleteSubtask(id, companyID string) error
	GetTaskByID(id, companyID string) (*domain.Task, error)
	GetSubtaskByID(id, companyID string) (*domain.Subtask, error)
	GetTasksByProjectID(projectID string) ([]domain.Task, error)
}

type LinkRepository interface {
	CreateLink(link *domain.Link) error
	GetAllLinks(companyID string) ([]domain.Link, error)
	UpdateLink(link *domain.Link) error
	DeleteLink(id, companyID string) error
	RegisterClick(linkID string) error
}

type ClientRepository interface {
	CreateClient(client *domain.Client) error
	GetClientByID(id, companyID string) (*domain.Client, error)
	GetAllClients(companyID string) ([]domain.Client, error)
	UpdateClient(client *domain.Client) error
	DeleteClient(id, companyID string) error
	AddComment(comment *domain.Comment) error
}

type CompanyRepository interface {
	CreateCompany(company *domain.Company) error
	GetCompanyByID(id string) (*domain.Company, error)
	UpdateCompany(company *domain.Company) error
}
