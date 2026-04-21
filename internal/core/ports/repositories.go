package ports

import (
	"construct-backend/internal/core/domain"
	"time"
)

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(id string) (*domain.User, error)
	VerifyUserName(username string) (*domain.UsernameVerification, error)
	UpdateUsername(userID, username string) error
	UpdateUserCompany(userID, companyID, role string) error
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
	GetProjectsByClientID(clientID, companyID string) ([]domain.Project, error)
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
	CreateDiaryEntry(entry *domain.DiaryEntry) error
	GetDiaryEntriesByProject(projectID, companyID string) ([]domain.DiaryEntry, error)
	GetPublicDiaryEntriesByProject(projectID string) ([]domain.DiaryEntry, error)
	GetDiaryEntryByID(id, projectID, companyID string) (*domain.DiaryEntry, error)
	UpdateDiaryEntry(entry *domain.DiaryEntry) error
	DeleteDiaryEntry(id, projectID, companyID string) error
}

type LinkRepository interface {
	CreateLink(link *domain.Link) error
	GetAllLinks(companyID string) ([]domain.Link, error)
	GetLinkAnalytics(companyID string, startDate, endDate *time.Time) ([]domain.LinkAnalyticsItem, error)
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
	GetCompanyBySlug(slug string) (*domain.Company, error)
	UpdateCompany(company *domain.Company) error
	UpdateCompanyPlan(companyID, plan, status, subscriptionID string, expiresAt *time.Time) error
}

type SubscriptionRepository interface {
	CountProjectsByCompany(companyID string) (int64, error)
}

type DashboardRepository interface {
	CountProjectsInProgress(companyID string) (int64, error)
	CountCompletedProjects(companyID string) (int64, error)
	CountActiveTasks(companyID string) (int64, error)
	CountLinkClicksByCompany(companyID string) (int64, error)
	CountClientsByCompany(companyID string) (int64, error)
}
