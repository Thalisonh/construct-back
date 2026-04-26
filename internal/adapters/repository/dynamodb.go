package repository

import (
	"construct-backend/internal/core/domain"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	entityCompany    = "company"
	entityUser       = "user"
	entityClient     = "client"
	entityComment    = "comment"
	entityProject    = "project"
	entityTask       = "task"
	entitySubtask    = "subtask"
	entityDiaryEntry = "diary_entry"
	entityLink       = "link"
	entityLinkClick  = "link_click"
)

type DynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

type dynamoItem struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	GSI1PK     string `dynamodbav:"GSI1PK,omitempty"`
	GSI1SK     string `dynamodbav:"GSI1SK,omitempty"`
	GSI2PK     string `dynamodbav:"GSI2PK,omitempty"`
	GSI2SK     string `dynamodbav:"GSI2SK,omitempty"`
	GSI3PK     string `dynamodbav:"GSI3PK,omitempty"`
	GSI3SK     string `dynamodbav:"GSI3SK,omitempty"`
	EntityType string `dynamodbav:"entity_type"`
	ID         string `dynamodbav:"id,omitempty"`
	CompanyID  string `dynamodbav:"company_id,omitempty"`
	UserID     string `dynamodbav:"user_id,omitempty"`
	ClientID   string `dynamodbav:"client_id,omitempty"`
	ProjectID  string `dynamodbav:"project_id,omitempty"`
	TaskID     string `dynamodbav:"task_id,omitempty"`
	LinkID     string `dynamodbav:"link_id,omitempty"`
	Status     string `dynamodbav:"status,omitempty"`
	IsPublic   bool   `dynamodbav:"is_public,omitempty"`
	CreatedAt  string `dynamodbav:"created_at,omitempty"`
	EntryDate  string `dynamodbav:"entry_date,omitempty"`

	User       *domain.User       `dynamodbav:"user,omitempty"`
	Company    *domain.Company    `dynamodbav:"company,omitempty"`
	Client     *domain.Client     `dynamodbav:"client,omitempty"`
	Comment    *domain.Comment    `dynamodbav:"comment,omitempty"`
	Project    *domain.Project    `dynamodbav:"project,omitempty"`
	Task       *domain.Task       `dynamodbav:"task,omitempty"`
	Subtask    *domain.Subtask    `dynamodbav:"subtask,omitempty"`
	DiaryEntry *domain.DiaryEntry `dynamodbav:"diary_entry,omitempty"`
	Link       *domain.Link       `dynamodbav:"link,omitempty"`
	LinkClick  *domain.LinkClick  `dynamodbav:"link_click,omitempty"`
}

func NewDynamoRepository(ctx context.Context, tableName string) (*DynamoRepository, error) {
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE is required when REPOSITORY_DRIVER=dynamodb")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &DynamoRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

func NewDynamoRepositoryFromEnv(ctx context.Context) (*DynamoRepository, error) {
	return NewDynamoRepository(ctx, os.Getenv("DYNAMODB_TABLE"))
}

func (r *DynamoRepository) putItem(ctx context.Context, item dynamoItem) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	return err
}

func (r *DynamoRepository) deleteItem(ctx context.Context, pk, sk string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       key(pk, sk),
	})
	return err
}

func (r *DynamoRepository) getItem(ctx context.Context, pk, sk string) (*dynamoItem, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key(pk, sk),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var item dynamoItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DynamoRepository) query(ctx context.Context, keyCondition expression.KeyConditionBuilder, options ...func(*dynamodb.QueryInput)) ([]dynamoItem, error) {
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	for _, option := range options {
		option(input)
	}

	var items []dynamoItem
	paginator := dynamodb.NewQueryPaginator(r.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		var pageItems []dynamoItem
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &pageItems); err != nil {
			return nil, err
		}
		items = append(items, pageItems...)
	}
	return items, nil
}

func (r *DynamoRepository) scanByEntity(ctx context.Context, entityType string) ([]dynamoItem, error) {
	filter := expression.Name("entity_type").Equal(expression.Value(entityType))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	var items []dynamoItem
	paginator := dynamodb.NewScanPaginator(r.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		var pageItems []dynamoItem
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &pageItems); err != nil {
			return nil, err
		}
		items = append(items, pageItems...)
	}
	return items, nil
}

// UserRepository

func (r *DynamoRepository) CreateUser(user *domain.User) error {
	item := dynamoItem{
		PK:         userPK(user.ID),
		SK:         metadataSK(),
		GSI1PK:     companyPK(user.CompanyID),
		GSI1SK:     userPK(user.ID),
		GSI2PK:     emailPK(user.Email),
		GSI2SK:     userPK(user.ID),
		GSI3PK:     usernamePK(user.Username),
		GSI3SK:     userPK(user.ID),
		EntityType: entityUser,
		ID:         user.ID,
		CompanyID:  user.CompanyID,
		User:       user,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) GetUserByEmail(email string) (*domain.User, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI2PK").Equal(expression.Value(emailPK(email))),
		withIndex("GSI2"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].User == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return items[0].User, nil
}

func (r *DynamoRepository) GetUserByID(id string) (*domain.User, error) {
	item, err := r.getItem(context.Background(), userPK(id), metadataSK())
	if err != nil {
		return nil, err
	}
	if item.User == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return item.User, nil
}

func (r *DynamoRepository) VerifyUserName(username string) (*domain.UsernameVerification, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI3PK").Equal(expression.Value(usernamePK(username))),
		withIndex("GSI3"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].User == nil {
		return nil, nil
	}
	return &domain.UsernameVerification{Username: items[0].User.Username}, nil
}

func (r *DynamoRepository) UpdateUsername(userID, username string) error {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.Username = username
	user.UpdatedAt = time.Now()
	return r.CreateUser(user)
}

func (r *DynamoRepository) UpdateUserCompany(userID, companyID, role string) error {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.CompanyID = companyID
	user.Role = role
	user.UpdatedAt = time.Now()
	return r.CreateUser(user)
}

func (r *DynamoRepository) GetUsername(userID string) (string, error) {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (r *DynamoRepository) GetPublicProfile(username string) (*domain.PublicProfile, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI3PK").Equal(expression.Value(usernamePK(username))),
		withIndex("GSI3"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].User == nil {
		return nil, gorm.ErrRecordNotFound
	}
	user := items[0].User
	links, err := r.GetAllLinks(user.CompanyID)
	if err != nil {
		return nil, err
	}
	return &domain.PublicProfile{
		ID:        user.ID,
		Username:  user.Username,
		Name:      user.Name,
		Bio:       user.Bio,
		Avatar:    user.Avatar,
		CompanyID: user.CompanyID,
		Links:     links,
	}, nil
}

func (r *DynamoRepository) UpdateBio(userID, bio string) error {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.Bio = bio
	user.UpdatedAt = time.Now()
	return r.CreateUser(user)
}

func (r *DynamoRepository) UpdateProfile(user *domain.User) error {
	current, err := r.GetUserByID(user.ID)
	if err != nil {
		return err
	}
	current.Name = user.Name
	current.Email = user.Email
	current.Phone = user.Phone
	current.UpdatedAt = time.Now()
	return r.CreateUser(current)
}

func (r *DynamoRepository) UpdatePassword(userID, password string) error {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.Password = password
	user.UpdatedAt = time.Now()
	return r.CreateUser(user)
}

func (r *DynamoRepository) ListUsersByCompanyID(companyID string) ([]domain.User, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI1PK").Equal(expression.Value(companyPK(companyID))).And(expression.Key("GSI1SK").BeginsWith(userPK(""))),
		withIndex("GSI1"),
	)
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(items))
	for _, item := range items {
		if item.User != nil {
			users = append(users, *item.User)
		}
	}
	return users, nil
}

// CompanyRepository

func (r *DynamoRepository) CreateCompany(company *domain.Company) error {
	item := dynamoItem{
		PK:         companyPK(company.ID),
		SK:         metadataSK(),
		GSI2PK:     companySlugPK(company.Slug),
		GSI2SK:     companyPK(company.ID),
		EntityType: entityCompany,
		ID:         company.ID,
		CompanyID:  company.ID,
		Company:    company,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) GetCompanyByID(id string) (*domain.Company, error) {
	item, err := r.getItem(context.Background(), companyPK(id), metadataSK())
	if err != nil {
		return nil, err
	}
	if item.Company == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return item.Company, nil
}

func (r *DynamoRepository) GetCompanyBySlug(slug string) (*domain.Company, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI2PK").Equal(expression.Value(companySlugPK(slug))),
		withIndex("GSI2"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].Company == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return items[0].Company, nil
}

func (r *DynamoRepository) UpdateCompany(company *domain.Company) error {
	company.UpdatedAt = time.Now()
	return r.CreateCompany(company)
}

func (r *DynamoRepository) UpdateCompanyPlan(companyID, plan, status, subscriptionID string, expiresAt *time.Time) error {
	company, err := r.GetCompanyByID(companyID)
	if err != nil {
		return err
	}
	company.Plan = plan
	company.PlanStatus = status
	company.SubscriptionID = subscriptionID
	company.PlanExpiresAt = expiresAt
	company.UpdatedAt = time.Now()
	return r.CreateCompany(company)
}

// ClientRepository

func (r *DynamoRepository) CreateClient(client *domain.Client) error {
	item := dynamoItem{
		PK:         companyPK(client.CompanyID),
		SK:         clientSK(client.ID),
		GSI1PK:     clientPK(client.ID),
		GSI1SK:     companyPK(client.CompanyID),
		EntityType: entityClient,
		ID:         client.ID,
		CompanyID:  client.CompanyID,
		UserID:     client.UserID,
		Client:     client,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) GetClientByID(id, companyID string) (*domain.Client, error) {
	item, err := r.getItem(context.Background(), companyPK(companyID), clientSK(id))
	if err != nil {
		return nil, err
	}
	if item.Client == nil {
		return nil, gorm.ErrRecordNotFound
	}
	client := *item.Client
	comments, err := r.commentsByClient(id)
	if err != nil {
		return nil, err
	}
	client.Comments = comments
	return &client, nil
}

func (r *DynamoRepository) GetAllClients(companyID string) ([]domain.Client, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(companyPK(companyID))).And(expression.Key("SK").BeginsWith(clientSK(""))),
	)
	if err != nil {
		return nil, err
	}
	clients := make([]domain.Client, 0, len(items))
	for _, item := range items {
		if item.Client != nil {
			clients = append(clients, *item.Client)
		}
	}
	return clients, nil
}

func (r *DynamoRepository) UpdateClient(client *domain.Client) error {
	client.UpdatedAt = time.Now()
	return r.CreateClient(client)
}

func (r *DynamoRepository) DeleteClient(id, companyID string) error {
	return r.deleteItem(context.Background(), companyPK(companyID), clientSK(id))
}

func (r *DynamoRepository) AddComment(comment *domain.Comment) error {
	item := dynamoItem{
		PK:         clientPK(comment.ClientID),
		SK:         commentSK(comment.CreatedAt, comment.ID),
		EntityType: entityComment,
		ID:         comment.ID,
		ClientID:   comment.ClientID,
		CreatedAt:  timeKey(comment.CreatedAt),
		Comment:    comment,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) commentsByClient(clientID string) ([]domain.Comment, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(clientPK(clientID))).And(expression.Key("SK").BeginsWith(commentSKPrefix())),
	)
	if err != nil {
		return nil, err
	}
	comments := make([]domain.Comment, 0, len(items))
	for _, item := range items {
		if item.Comment != nil {
			comments = append(comments, *item.Comment)
		}
	}
	return comments, nil
}

// ProjectRepository

func (r *DynamoRepository) CreateProject(project *domain.Project) error {
	item := projectItem(project)
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) GetAllProjects(companyID string) ([]domain.Project, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(companyPK(companyID))).And(expression.Key("SK").BeginsWith(projectSK(""))),
	)
	if err != nil {
		return nil, err
	}
	projects := make([]domain.Project, 0, len(items))
	for _, item := range items {
		if item.Project == nil {
			continue
		}
		project, err := r.enrichProject(*item.Project)
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}
	return projects, nil
}

func (r *DynamoRepository) GetProjectsByClientID(clientID, companyID string) ([]domain.Project, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI3PK").Equal(expression.Value(clientPK(clientID))).And(expression.Key("GSI3SK").BeginsWith(projectPK(""))),
		withIndex("GSI3"),
	)
	if err != nil {
		return nil, err
	}
	projects := make([]domain.Project, 0, len(items))
	for _, item := range items {
		if item.Project == nil || item.Project.CompanyID != companyID {
			continue
		}
		project, err := r.enrichProject(*item.Project)
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}
	return projects, nil
}

func (r *DynamoRepository) GetProjectByID(id, companyID string) (*domain.Project, error) {
	item, err := r.getItem(context.Background(), companyPK(companyID), projectSK(id))
	if err != nil {
		return nil, err
	}
	if item.Project == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.enrichProject(*item.Project)
}

func (r *DynamoRepository) GetPublicProjectByID(id string) (*domain.Project, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI1PK").Equal(expression.Value(projectPK(id))),
		withIndex("GSI1"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].Project == nil || !items[0].Project.IsPublic {
		return nil, gorm.ErrRecordNotFound
	}
	return r.enrichProject(*items[0].Project)
}

func (r *DynamoRepository) UpdateProject(project *domain.Project) error {
	project.UpdatedAt = time.Now()
	return r.CreateProject(project)
}

func (r *DynamoRepository) DeleteProject(id, companyID string) error {
	return r.deleteItem(context.Background(), companyPK(companyID), projectSK(id))
}

func (r *DynamoRepository) AddTask(task *domain.Task) error {
	item := dynamoItem{
		PK:         projectPK(task.ProjectID),
		SK:         taskSK(task.ID),
		GSI1PK:     taskPK(task.ID),
		GSI1SK:     companyPK(task.CompanyID),
		GSI2PK:     companyPK(task.CompanyID),
		GSI2SK:     taskStatusSK(task.Status, task.ID),
		EntityType: entityTask,
		ID:         task.ID,
		CompanyID:  task.CompanyID,
		UserID:     task.UserID,
		ProjectID:  task.ProjectID,
		Status:     task.Status,
		CreatedAt:  timeKey(task.CreatedAt),
		Task:       task,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) AddSubtask(subtask *domain.Subtask) error {
	item := dynamoItem{
		PK:         taskPK(subtask.TaskID),
		SK:         subtaskSK(subtask.ID),
		GSI1PK:     subtaskPK(subtask.ID),
		GSI1SK:     companyPK(subtask.CompanyID),
		EntityType: entitySubtask,
		ID:         subtask.ID,
		CompanyID:  subtask.CompanyID,
		UserID:     subtask.UserID,
		TaskID:     subtask.TaskID,
		Status:     subtask.Status,
		CreatedAt:  timeKey(subtask.CreatedAt),
		Subtask:    subtask,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) UpdateTask(task *domain.Task) error {
	return r.AddTask(task)
}

func (r *DynamoRepository) UpdateSubtask(subtask *domain.Subtask) error {
	return r.AddSubtask(subtask)
}

func (r *DynamoRepository) UpdateSubtaskByTaskID(taskID string) error {
	subtasks, err := r.subtasksByTask(taskID)
	if err != nil {
		return err
	}
	for _, subtask := range subtasks {
		subtask.Status = "Completed"
		if err := r.AddSubtask(&subtask); err != nil {
			return err
		}
	}
	return nil
}

func (r *DynamoRepository) DeleteTask(id, companyID string) error {
	task, err := r.GetTaskByID(id, companyID)
	if err != nil {
		return err
	}
	return r.deleteItem(context.Background(), projectPK(task.ProjectID), taskSK(id))
}

func (r *DynamoRepository) DeleteSubtask(id, companyID string) error {
	subtask, err := r.GetSubtaskByID(id, companyID)
	if err != nil {
		return err
	}
	return r.deleteItem(context.Background(), taskPK(subtask.TaskID), subtaskSK(id))
}

func (r *DynamoRepository) GetTaskByID(id, companyID string) (*domain.Task, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI1PK").Equal(expression.Value(taskPK(id))),
		withIndex("GSI1"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].Task == nil || items[0].Task.CompanyID != companyID {
		return nil, gorm.ErrRecordNotFound
	}
	task := *items[0].Task
	subtasks, err := r.subtasksByTask(id)
	if err != nil {
		return nil, err
	}
	task.Subtasks = subtasks
	return &task, nil
}

func (r *DynamoRepository) GetSubtaskByID(id, companyID string) (*domain.Subtask, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI1PK").Equal(expression.Value(subtaskPK(id))),
		withIndex("GSI1"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].Subtask == nil || items[0].Subtask.CompanyID != companyID {
		return nil, gorm.ErrRecordNotFound
	}
	return items[0].Subtask, nil
}

func (r *DynamoRepository) GetTasksByProjectID(projectID string) ([]domain.Task, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(projectPK(projectID))).And(expression.Key("SK").BeginsWith(taskSK(""))),
	)
	if err != nil {
		return nil, err
	}
	tasks := make([]domain.Task, 0, len(items))
	for _, item := range items {
		if item.Task == nil {
			continue
		}
		task := *item.Task
		subtasks, err := r.subtasksByTask(task.ID)
		if err != nil {
			return nil, err
		}
		task.Subtasks = subtasks
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *DynamoRepository) CreateDiaryEntry(entry *domain.DiaryEntry) error {
	item := diaryEntryItem(entry)
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) GetDiaryEntriesByProject(projectID, companyID string) ([]domain.DiaryEntry, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(projectPK(projectID))).And(expression.Key("SK").BeginsWith(diarySKPrefix())),
	)
	if err != nil {
		return nil, err
	}
	entries := make([]domain.DiaryEntry, 0, len(items))
	for _, item := range items {
		if item.DiaryEntry != nil && item.DiaryEntry.CompanyID == companyID {
			entries = append(entries, *item.DiaryEntry)
		}
	}
	sortDiaryEntries(entries)
	return entries, nil
}

func (r *DynamoRepository) GetPublicDiaryEntriesByProject(projectID string) ([]domain.DiaryEntry, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(projectPK(projectID))).And(expression.Key("SK").BeginsWith(diarySKPrefix())),
	)
	if err != nil {
		return nil, err
	}
	entries := make([]domain.DiaryEntry, 0, len(items))
	for _, item := range items {
		if item.DiaryEntry == nil {
			continue
		}
		entry := *item.DiaryEntry
		publicItems := make([]domain.DiaryItem, 0, len(entry.Items))
		for _, diaryItem := range entry.Items {
			if diaryItem.Visibility == "public" {
				publicItems = append(publicItems, diaryItem)
			}
		}
		if len(publicItems) == 0 {
			continue
		}
		entry.Items = publicItems
		entries = append(entries, entry)
	}
	sortDiaryEntries(entries)
	return entries, nil
}

func (r *DynamoRepository) GetDiaryEntryByID(id, projectID, companyID string) (*domain.DiaryEntry, error) {
	entries, err := r.GetDiaryEntriesByProject(projectID, companyID)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.ID == id {
			return &entry, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *DynamoRepository) UpdateDiaryEntry(entry *domain.DiaryEntry) error {
	return r.CreateDiaryEntry(entry)
}

func (r *DynamoRepository) DeleteDiaryEntry(id, projectID, companyID string) error {
	entry, err := r.GetDiaryEntryByID(id, projectID, companyID)
	if err != nil {
		return err
	}
	return r.deleteItem(context.Background(), projectPK(projectID), diarySK(entry.EntryDate, entry.ID))
}

func (r *DynamoRepository) enrichProject(project domain.Project) (*domain.Project, error) {
	if project.ClientID != "" {
		client, err := r.GetClientByID(project.ClientID, project.CompanyID)
		if err == nil {
			project.Client = client
		}
	}
	tasks, err := r.GetTasksByProjectID(project.ID)
	if err != nil {
		return nil, err
	}
	project.Tasks = tasks
	return &project, nil
}

func (r *DynamoRepository) subtasksByTask(taskID string) ([]domain.Subtask, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(taskPK(taskID))).And(expression.Key("SK").BeginsWith(subtaskSK(""))),
	)
	if err != nil {
		return nil, err
	}
	subtasks := make([]domain.Subtask, 0, len(items))
	for _, item := range items {
		if item.Subtask != nil {
			subtasks = append(subtasks, *item.Subtask)
		}
	}
	return subtasks, nil
}

// LinkRepository

func (r *DynamoRepository) CreateLink(link *domain.Link) error {
	item := dynamoItem{
		PK:         companyPK(link.CompanyID),
		SK:         linkSK(link.ID),
		GSI1PK:     linkPK(link.ID),
		GSI1SK:     companyPK(link.CompanyID),
		GSI2PK:     userPK(link.UserID),
		GSI2SK:     linkPK(link.ID),
		EntityType: entityLink,
		ID:         link.ID,
		CompanyID:  link.CompanyID,
		UserID:     link.UserID,
		CreatedAt:  timeKey(link.CreatedAt),
		Link:       link,
	}
	return r.putItem(context.Background(), item)
}

func (r *DynamoRepository) GetAllLinks(companyID string) ([]domain.Link, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(companyPK(companyID))).And(expression.Key("SK").BeginsWith(linkSK(""))),
	)
	if err != nil {
		return nil, err
	}
	links := make([]domain.Link, 0, len(items))
	for _, item := range items {
		if item.Link != nil {
			links = append(links, *item.Link)
		}
	}
	return links, nil
}

func (r *DynamoRepository) GetLinkAnalytics(companyID string, startDate, endDate *time.Time) ([]domain.LinkAnalyticsItem, error) {
	links, err := r.GetAllLinks(companyID)
	if err != nil {
		return nil, err
	}
	analytics := make([]domain.LinkAnalyticsItem, 0, len(links))
	for _, link := range links {
		clicks, err := r.countClicks(link.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, domain.LinkAnalyticsItem{
			ID:          link.ID,
			Description: link.Description,
			URL:         link.URL,
			Clicks:      clicks,
		})
	}
	sort.Slice(analytics, func(i, j int) bool {
		if analytics[i].Clicks == analytics[j].Clicks {
			return analytics[i].Description < analytics[j].Description
		}
		return analytics[i].Clicks > analytics[j].Clicks
	})
	return analytics, nil
}

func (r *DynamoRepository) UpdateLink(link *domain.Link) error {
	link.UpdatedAt = time.Now()
	return r.CreateLink(link)
}

func (r *DynamoRepository) DeleteLink(id, companyID string) error {
	return r.deleteItem(context.Background(), companyPK(companyID), linkSK(id))
}

func (r *DynamoRepository) RegisterClick(linkID string) error {
	link, err := r.getLinkByID(linkID)
	if err != nil {
		return err
	}
	click := &domain.LinkClick{
		ID:        uuid.New().String(),
		LinkID:    linkID,
		CreatedAt: time.Now(),
	}
	item := dynamoItem{
		PK:         linkPK(linkID),
		SK:         clickSK(click.CreatedAt, click.ID),
		GSI1PK:     companyPK(link.CompanyID),
		GSI1SK:     clickSK(click.CreatedAt, click.ID),
		EntityType: entityLinkClick,
		ID:         click.ID,
		CompanyID:  link.CompanyID,
		LinkID:     linkID,
		CreatedAt:  timeKey(click.CreatedAt),
		LinkClick:  click,
	}
	if err := r.putItem(context.Background(), item); err != nil {
		return err
	}
	link.Count++
	return r.CreateLink(link)
}

func (r *DynamoRepository) getLinkByID(id string) (*domain.Link, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI1PK").Equal(expression.Value(linkPK(id))),
		withIndex("GSI1"),
		withLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].Link == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return items[0].Link, nil
}

func (r *DynamoRepository) countClicks(linkID string, startDate, endDate *time.Time) (int64, error) {
	items, err := r.query(context.Background(),
		expression.Key("PK").Equal(expression.Value(linkPK(linkID))).And(expression.Key("SK").BeginsWith(clickSKPrefix())),
	)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, item := range items {
		if item.LinkClick == nil {
			continue
		}
		clickTime := item.LinkClick.CreatedAt
		if startDate != nil && clickTime.Before(*startDate) {
			continue
		}
		if endDate != nil && !clickTime.Before(*endDate) {
			continue
		}
		count++
	}
	return count, nil
}

// DashboardRepository and SubscriptionRepository

func (r *DynamoRepository) CountProjectsByCompany(companyID string) (int64, error) {
	projects, err := r.GetAllProjects(companyID)
	return int64(len(projects)), err
}

func (r *DynamoRepository) CountProjectsInProgress(companyID string) (int64, error) {
	projects, err := r.GetAllProjects(companyID)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, project := range projects {
		if project.Status != "Completed" {
			count++
		}
	}
	return count, nil
}

func (r *DynamoRepository) CountCompletedProjects(companyID string) (int64, error) {
	projects, err := r.GetAllProjects(companyID)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, project := range projects {
		if project.Status == "Completed" {
			count++
		}
	}
	return count, nil
}

func (r *DynamoRepository) CountActiveTasks(companyID string) (int64, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI2PK").Equal(expression.Value(companyPK(companyID))).And(expression.Key("GSI2SK").BeginsWith("TASK#")),
		withIndex("GSI2"),
	)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, item := range items {
		if item.Task != nil && item.Task.Status != "Completed" {
			count++
		}
	}
	return count, nil
}

func (r *DynamoRepository) CountLinkClicksByCompany(companyID string) (int64, error) {
	items, err := r.query(context.Background(),
		expression.Key("GSI1PK").Equal(expression.Value(companyPK(companyID))).And(expression.Key("GSI1SK").BeginsWith(clickSKPrefix())),
		withIndex("GSI1"),
	)
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (r *DynamoRepository) CountClientsByCompany(companyID string) (int64, error) {
	clients, err := r.GetAllClients(companyID)
	return int64(len(clients)), err
}

// Item builders and key helpers

func projectItem(project *domain.Project) dynamoItem {
	return dynamoItem{
		PK:         companyPK(project.CompanyID),
		SK:         projectSK(project.ID),
		GSI1PK:     projectPK(project.ID),
		GSI1SK:     companyPK(project.CompanyID),
		GSI3PK:     clientPK(project.ClientID),
		GSI3SK:     projectPK(project.ID),
		EntityType: entityProject,
		ID:         project.ID,
		CompanyID:  project.CompanyID,
		UserID:     project.UserID,
		ClientID:   project.ClientID,
		Status:     project.Status,
		IsPublic:   project.IsPublic,
		Project:    project,
	}
}

func diaryEntryItem(entry *domain.DiaryEntry) dynamoItem {
	return dynamoItem{
		PK:         projectPK(entry.ProjectID),
		SK:         diarySK(entry.EntryDate, entry.ID),
		EntityType: entityDiaryEntry,
		ID:         entry.ID,
		CompanyID:  entry.CompanyID,
		UserID:     entry.UserID,
		ProjectID:  entry.ProjectID,
		EntryDate:  dayKey(entry.EntryDate),
		CreatedAt:  timeKey(entry.CreatedAt),
		DiaryEntry: entry,
	}
}

func key(pk, sk string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}
}

func withIndex(index string) func(*dynamodb.QueryInput) {
	return func(input *dynamodb.QueryInput) {
		input.IndexName = aws.String(index)
	}
}

func withLimit(limit int32) func(*dynamodb.QueryInput) {
	return func(input *dynamodb.QueryInput) {
		input.Limit = aws.Int32(limit)
	}
}

func metadataSK() string             { return "METADATA" }
func companyPK(id string) string     { return "COMPANY#" + id }
func userPK(id string) string        { return "USER#" + id }
func usernamePK(value string) string { return "USERNAME#" + strings.ToLower(value) }
func emailPK(value string) string    { return "EMAIL#" + strings.ToLower(value) }
func clientPK(id string) string      { return "CLIENT#" + id }
func clientSK(id string) string      { return "CLIENT#" + id }
func projectPK(id string) string     { return "PROJECT#" + id }
func projectSK(id string) string     { return "PROJECT#" + id }
func taskPK(id string) string        { return "TASK#" + id }
func taskSK(id string) string        { return "TASK#" + id }
func subtaskPK(id string) string     { return "SUBTASK#" + id }
func subtaskSK(id string) string     { return "SUBTASK#" + id }
func linkPK(id string) string        { return "LINK#" + id }
func linkSK(id string) string        { return "LINK#" + id }
func companySlugPK(slug string) string {
	return "COMPANY_SLUG#" + strings.ToLower(slug)
}

func taskStatusSK(status, id string) string {
	if status == "" {
		status = "unknown"
	}
	return "TASK#" + status + "#" + id
}

func commentSKPrefix() string {
	return "COMMENT#"
}

func commentSK(createdAt time.Time, id string) string {
	return commentSKPrefix() + timeKey(createdAt) + "#" + id
}

func diarySKPrefix() string {
	return "DIARY#"
}

func diarySK(entryDate time.Time, id string) string {
	return diarySKPrefix() + dayKey(entryDate) + "#" + id
}

func clickSKPrefix() string {
	return "CLICK#"
}

func clickSK(createdAt time.Time, id string) string {
	return clickSKPrefix() + timeKey(createdAt) + "#" + id
}

func timeKey(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func dayKey(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}

func sortDiaryEntries(entries []domain.DiaryEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].EntryDate.Equal(entries[j].EntryDate) {
			return entries[i].CreatedAt.After(entries[j].CreatedAt)
		}
		return entries[i].EntryDate.After(entries[j].EntryDate)
	})
}
