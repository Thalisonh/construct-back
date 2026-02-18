package ports

import "construct-backend/internal/core/domain"

type AuthService interface {
	Signup(email, password, name, companyName, cnpj string) (string, error)
	Login(email, password string) (string, error)
	LoginWithGoogle(idToken string) (string, error)
	VerifyToken(token string) error
}

type ProjectService interface {
	CreateProject(companyID, userID, name, clientID, address, summary string, startDate string) (*domain.Project, error)
	ListProjects(companyID string) ([]domain.Project, error)
	GetProject(id, companyID string) (*domain.Project, error)
	GetPublicProject(id string) (*domain.Project, error)
	UpdateProject(id, name, clientID, address, summary, startDate string, isPublic bool, companyID string) (*domain.Project, error)
	DeleteProject(id, companyID string) error
	AddTask(projectID, name, status, dueDate, companyID, userID string) (*domain.Task, error)
	AddSubtask(taskID, name, status, companyID, userID string) (*domain.Subtask, error)
	UpdateTask(id, companyID string) (*domain.Task, error)
	UpdateSubtask(id, companyID string) (*domain.Subtask, error)
	DeleteTask(id, companyID string) error
	DeleteSubtask(id, companyID string) error
	GetTask(id, companyID string) (*domain.Task, error)
	GetSubtask(id, companyID string) (*domain.Subtask, error)
	ListTasks(projectID string) ([]domain.Task, error)
}

type LinkService interface {
	CreateLink(companyID, userID, url, description string) (*domain.Link, error)
	UpdateLink(companyID, url, description, id string) (*domain.Link, error)
	ListLinks(companyID string) ([]domain.Link, error)
	DeleteLink(id, companyID string) error
	TrackLinkClick(id string) error
}

type UserService interface {
	VerifyUserName(username string) error
	UpdateUsername(userID, username string) error
	GetUsername(userID string) (string, error)
	GetPublicProfile(username string) (*domain.PublicProfile, error)
	UpdateBio(userID, bio string) error
	GetProfile(userID string) (*domain.User, error)
	UpdateProfile(userID, name, email, phone, companyID string) error
	UpdatePassword(userID, oldPassword, newPassword string) error
	GetCompanyMembers(companyID string) ([]domain.User, error)
	AddCompanyMember(companyID, email, name, password, role string) (*domain.User, error)
}

type ClientService interface {
	CreateClient(companyID, userID, name, phone, address, summary string) (*domain.Client, error)
	GetClient(id, companyID string) (*domain.Client, error)
	ListClients(companyID string) ([]domain.Client, error)
	UpdateClient(id, name, phone, address, summary, companyID string) (*domain.Client, error)
	DeleteClient(id, companyID string) error
	AddComment(clientID, content string) (*domain.Comment, error)
}

type CompanyService interface {
	CreateCompany(name, cnpj, email, phone, address string) (*domain.Company, error)
	GetCompany(id string) (*domain.Company, error)
	UpdateCompany(id, name, email, phone, address string) (*domain.Company, error)
}
