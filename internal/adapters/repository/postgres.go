package repository

import (
	"construct-backend/internal/core/domain"
	"errors"

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
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) UpdateUsername(userID, username string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("name", username).Error
}

func (r *PostgresRepository) GetUsername(userID string) (string, error) {
	var user domain.User
	if err := r.db.Select("name").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	return user.Name, nil
}

func (r *PostgresRepository) GetPublicProfile(username string) (*domain.PublicProfile, error) {
	var user domain.User
	if err := r.db.Where("name = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	var links []domain.Link
	if err := r.db.Where("user_id = ?", user.ID).Find(&links).Error; err != nil {
		return nil, err
	}

	return &domain.PublicProfile{
		ID:       user.ID,
		Username: user.Name,
		Name:     user.Name,
		Links:    links,
		Bio:      user.Bio,
		Avatar:   user.Avatar,
	}, nil
}

// ProjectRepository Implementation

func (r *PostgresRepository) CreateProject(project *domain.Project) error {
	return r.db.Create(project).Error
}

func (r *PostgresRepository) GetAllProjects(userID string) ([]domain.Project, error) {
	var projects []domain.Project
	if err := r.db.Preload("Tasks.Subtasks").Preload("Client").Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *PostgresRepository) GetProjectByID(id string) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.Preload("Tasks.Subtasks").Preload("Client").Where("id = ?", id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *PostgresRepository) UpdateProject(project *domain.Project) error {
	return r.db.Save(project).Error
}

func (r *PostgresRepository) DeleteProject(id string) error {
	return r.db.Delete(&domain.Project{}, "id = ?", id).Error
}

func (r *PostgresRepository) AddTask(task *domain.Task) error {
	return r.db.Create(task).Error
}

func (r *PostgresRepository) AddSubtask(subtask *domain.Subtask) error {
	return r.db.Create(subtask).Error
}

func (r *PostgresRepository) UpdateTask(task *domain.Task) error {
	return r.db.Save(task).Error
}

func (r *PostgresRepository) UpdateSubtask(subtask *domain.Subtask) error {
	return r.db.Save(subtask).Error
}

func (r *PostgresRepository) DeleteTask(id string) error {
	return r.db.Delete(&domain.Task{}, "id = ?", id).Error
}

func (r *PostgresRepository) DeleteSubtask(id string) error {
	return r.db.Delete(&domain.Subtask{}, "id = ?", id).Error
}

func (r *PostgresRepository) GetTaskByID(id string) (*domain.Task, error) {
	var task domain.Task
	if err := r.db.Preload("Subtasks").Where("id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *PostgresRepository) GetSubtaskByID(id string) (*domain.Subtask, error) {
	var subtask domain.Subtask
	if err := r.db.Where("id = ?", id).First(&subtask).Error; err != nil {
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

func (r *PostgresRepository) GetAllLinks(userID string) ([]domain.Link, error) {
	var links []domain.Link
	if err := r.db.Where("user_id = ?", userID).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (r *PostgresRepository) UpdateLink(link *domain.Link) error {
	return r.db.Save(link).Error
}

func (r *PostgresRepository) DeleteLink(id string) error {
	return r.db.Delete(&domain.Link{}, "id = ?", id).Error
}

// ClientRepository Implementation

func (r *PostgresRepository) CreateClient(client *domain.Client) error {
	return r.db.Create(client).Error
}

func (r *PostgresRepository) GetClientByID(id string) (*domain.Client, error) {
	var client domain.Client
	if err := r.db.Preload("Comments").Where("id = ?", id).First(&client).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *PostgresRepository) GetAllClients() ([]domain.Client, error) {
	var clients []domain.Client
	if err := r.db.Preload("Comments").Find(&clients).Error; err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *PostgresRepository) UpdateClient(client *domain.Client) error {
	return r.db.Save(client).Error
}

func (r *PostgresRepository) DeleteClient(id string) error {
	return r.db.Delete(&domain.Client{}, "id = ?", id).Error
}

func (r *PostgresRepository) AddComment(comment *domain.Comment) error {
	return r.db.Create(comment).Error
}

func (r *PostgresRepository) UpdateSubtaskByTaskID(taskID string) error {
	return r.db.Model(&domain.Subtask{}).Where("task_id = ?", taskID).Update("status", "Completed").Error
}
