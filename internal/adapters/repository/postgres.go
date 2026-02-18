package repository

import (
	"construct-backend/internal/core/domain"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// UserRepository Implementation

func (r *PostgresRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *PostgresRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByID(id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) VerifyUserName(username string) (*domain.UsernameVerification, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.UsernameVerification{Username: user.Username}, nil
}

func (r *PostgresRepository) UpdateUsername(userID, username string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("username", username).Error
}

func (r *PostgresRepository) GetUsername(userID string) (string, error) {
	var user domain.User
	if err := r.db.Select("username").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	return user.Username, nil
}

func (r *PostgresRepository) GetPublicProfile(username string) (*domain.PublicProfile, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	var links []domain.Link
	if err := r.db.Where("user_id = ?", user.ID).Find(&links).Error; err != nil {
		return nil, err
	}

	return &domain.PublicProfile{
		ID:        user.ID,
		Username:  user.Username,
		Name:      user.Name,
		Links:     links,
		Bio:       user.Bio,
		Avatar:    user.Avatar,
		CompanyID: user.CompanyID,
	}, nil
}

func (r *PostgresRepository) UpdateProfile(user *domain.User) error {
	return r.db.Model(user).Select("Name", "Email", "Phone", "CompanyID").Updates(user).Error
}

func (r *PostgresRepository) UpdatePassword(userID, password string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("password", password).Error
}

func (r *PostgresRepository) ListUsersByCompanyID(companyID string) ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Where("company_id = ?", companyID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ProjectRepository Implementation

func (r *PostgresRepository) CreateProject(project *domain.Project) error {
	return r.db.Create(project).Error
}

func (r *PostgresRepository) GetAllProjects(companyID string) ([]domain.Project, error) {
	var projects []domain.Project
	if err := r.db.Preload("Tasks.Subtasks").Preload("Client").Where("company_id = ?", companyID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *PostgresRepository) GetProjectByID(id, companyID string) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.Preload("Tasks.Subtasks").Preload("Client").Where("id = ? AND company_id = ?", id, companyID).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *PostgresRepository) GetPublicProjectByID(id string) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.Preload("Tasks.Subtasks").Preload("Client").Where("id = ? AND is_public = true", id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *PostgresRepository) UpdateProject(project *domain.Project) error {
	return r.db.Where("user_id = ?", project.UserID).Save(project).Error
}

func (r *PostgresRepository) DeleteProject(id, companyID string) error {
	return r.db.Delete(&domain.Project{}, "id = ? AND company_id = ?", id, companyID).Error
}

func (r *PostgresRepository) AddTask(task *domain.Task) error {
	return r.db.Create(task).Error
}

func (r *PostgresRepository) AddSubtask(subtask *domain.Subtask) error {
	return r.db.Create(subtask).Error
}

func (r *PostgresRepository) UpdateTask(task *domain.Task) error {
	return r.db.Where("company_id = ?", task.CompanyID).Save(task).Error
}

func (r *PostgresRepository) UpdateSubtask(subtask *domain.Subtask) error {
	return r.db.Where("company_id = ?", subtask.CompanyID).Save(subtask).Error
}

func (r *PostgresRepository) DeleteTask(id, companyID string) error {
	return r.db.Delete(&domain.Task{}, "id = ? AND company_id = ?", id, companyID).Error
}

func (r *PostgresRepository) DeleteSubtask(id, companyID string) error {
	return r.db.Delete(&domain.Subtask{}, "id = ? AND company_id = ?", id, companyID).Error
}

func (r *PostgresRepository) GetTaskByID(id, companyID string) (*domain.Task, error) {
	var task domain.Task
	if err := r.db.Preload("Subtasks").Where("id = ? AND company_id = ?", id, companyID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *PostgresRepository) GetSubtaskByID(id, companyID string) (*domain.Subtask, error) {
	var subtask domain.Subtask
	if err := r.db.Where("id = ? AND company_id = ?", id, companyID).First(&subtask).Error; err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (r *PostgresRepository) GetTasksByProjectID(projectID string) ([]domain.Task, error) {
	var tasks []domain.Task
	if err := r.db.Preload("Subtasks").Where("project_id = ?", projectID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// LinkRepository Implementation

func (r *PostgresRepository) CreateLink(link *domain.Link) error {
	return r.db.Create(link).Error
}

func (r *PostgresRepository) GetAllLinks(companyID string) ([]domain.Link, error) {
	var links []domain.Link
	if err := r.db.Select("links.*, (SELECT COUNT(*) FROM link_clicks WHERE link_clicks.link_id = links.id) as count").Where("company_id = ?", companyID).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (r *PostgresRepository) UpdateLink(link *domain.Link) error {
	return r.db.Where("company_id = ?", link.CompanyID).Save(link).Error
}

func (r *PostgresRepository) DeleteLink(id, companyID string) error {
	return r.db.Delete(&domain.Link{}, "id = ? AND company_id = ?", id, companyID).Error
}

func (r *PostgresRepository) RegisterClick(linkID string) error {
	click := &domain.LinkClick{
		ID:        uuid.New().String(),
		LinkID:    linkID,
		CreatedAt: time.Now(),
	}
	return r.db.Create(click).Error
}

// ClientRepository Implementation

func (r *PostgresRepository) CreateClient(client *domain.Client) error {
	return r.db.Create(client).Error
}

func (r *PostgresRepository) GetClientByID(id, companyID string) (*domain.Client, error) {
	var client domain.Client
	if err := r.db.Preload("Comments").Where("id = ? AND company_id = ?", id, companyID).First(&client).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *PostgresRepository) GetAllClients(companyID string) ([]domain.Client, error) {
	var clients []domain.Client
	err := r.db.Where("company_id = ?", companyID).Find(&clients).Error
	return clients, err
}

func (r *PostgresRepository) UpdateClient(client *domain.Client) error {
	return r.db.Where("company_id = ?", client.CompanyID).Save(client).Error
}

func (r *PostgresRepository) DeleteClient(id, companyID string) error {
	return r.db.Delete(&domain.Client{}, "id = ? AND company_id = ?", id, companyID).Error
}

func (r *PostgresRepository) AddComment(comment *domain.Comment) error {
	return r.db.Create(comment).Error
}

func (r *PostgresRepository) UpdateSubtaskByTaskID(taskID string) error {
	return r.db.Model(&domain.Subtask{}).Where("task_id = ?", taskID).Update("status", "Completed").Error
}

func (r *PostgresRepository) UpdateBio(userID, bio string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("bio", bio).Error
}

// CompanyRepository Implementation

func (r *PostgresRepository) CreateCompany(company *domain.Company) error {
	return r.db.Create(company).Error
}

func (r *PostgresRepository) GetCompanyByID(id string) (*domain.Company, error) {
	var company domain.Company
	if err := r.db.Where("id = ?", id).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *PostgresRepository) UpdateCompany(company *domain.Company) error {
	return r.db.Save(company).Error
}
