package repository

import (
	"construct-backend/internal/core/domain"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoDBRepository struct {
	db *mongo.Database
}

func NewMongoDBRepository(db *mongo.Database) *MongoDBRepository {
	return &MongoDBRepository{db: db}
}

// User Repository Implementation

func (r *MongoDBRepository) CreateUser(user *domain.User) error {
	_, err := r.db.Collection("users").InsertOne(context.Background(), user)
	return err
}

func (r *MongoDBRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *MongoDBRepository) GetUserByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Project Repository Implementation

func (r *MongoDBRepository) CreateProject(project *domain.Project) error {
	_, err := r.db.Collection("projects").InsertOne(context.Background(), project)
	return err
}

func (r *MongoDBRepository) GetAllProjects(userID string) ([]domain.Project, error) {
	var projects []domain.Project
	cursor, err := r.db.Collection("projects").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *MongoDBRepository) GetProjectByID(id string) (*domain.Project, error) {
	var project domain.Project
	err := r.db.Collection("projects").FindOne(context.Background(), bson.M{"_id": id}).Decode(&project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *MongoDBRepository) UpdateProject(project *domain.Project) error {
	_, err := r.db.Collection("projects").ReplaceOne(context.Background(), bson.M{"_id": project.ID}, project)
	return err
}

func (r *MongoDBRepository) DeleteProject(id string) error {
	_, err := r.db.Collection("projects").DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

// Link Repository Implementation

func (r *MongoDBRepository) CreateLink(link *domain.Link) error {
	_, err := r.db.Collection("links").InsertOne(context.Background(), link)
	return err
}

func (r *MongoDBRepository) GetAllLinks(userID string) ([]domain.Link, error) {
	var links []domain.Link
	cursor, err := r.db.Collection("links").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &links); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *MongoDBRepository) DeleteLink(id string) error {
	_, err := r.db.Collection("links").DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

func (r *MongoDBRepository) UpdateLink(link *domain.Link) error {
	_, err := r.db.Collection("links").ReplaceOne(context.Background(), bson.M{"_id": link.ID}, link)
	return err
}
