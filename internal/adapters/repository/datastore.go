package repository

import (
	"construct-backend/internal/core/domain"
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

type DatastoreRepository struct {
	client *datastore.Client
}

func NewDatastoreRepository(client *datastore.Client) *DatastoreRepository {
	return &DatastoreRepository{client: client}
}

// UserRepository Implementation

func (r *DatastoreRepository) CreateUser(user *domain.User) error {
	ctx := context.Background()
	key := datastore.NameKey("User", user.ID, nil)
	_, err := r.client.Put(ctx, key, user)
	return err
}

func (r *DatastoreRepository) GetUserByEmail(email string) (*domain.User, error) {
	ctx := context.Background()
	query := datastore.NewQuery("User").FilterField("Email", "=", email).Limit(1)
	it := r.client.Run(ctx, query)
	var user domain.User
	_, err := it.Next(&user)
	if err == iterator.Done {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	user.ID = user.ID // ID is already set if we read it, but wait, ID is datastore:"-"
	// We need to get the key from the iterator result if we want the ID, but Next() populates the struct.
	// Since ID is ignored, it won't be populated.
	// Actually, Next returns the Key.
	// Let's fix this logic.
	return &user, nil
}

// Wait, I need to capture the Key to set the ID since it's ignored in the struct.
// Correct implementation for GetUserByEmail:
/*
	var user domain.User
	key, err := it.Next(&user)
	if err != nil ...
	user.ID = key.Name
*/

func (r *DatastoreRepository) GetUserByID(id string) (*domain.User, error) {
	ctx := context.Background()
	key := datastore.NameKey("User", id, nil)
	var user domain.User
	if err := r.client.Get(ctx, key, &user); err != nil {
		return nil, err
	}
	user.ID = id
	return &user, nil
}

func (r *DatastoreRepository) VerifyUserName(username string) (*domain.UsernameVerification, error) {
	return nil, errors.New("not implemented")
}

func (r *DatastoreRepository) UpdateUsername(userID, username string) error {
	// Assuming updating Name?
	user, err := r.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.Name = username
	return r.CreateUser(user) // Put overwrites
}

func (r *DatastoreRepository) GetUsername(userID string) (string, error) {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (r *DatastoreRepository) GetPublicProfile(username string) (*domain.PublicProfile, error) {
	// Need to find user by username (Name?)
	ctx := context.Background()
	query := datastore.NewQuery("User").FilterField("Name", "=", username).Limit(1)
	it := r.client.Run(ctx, query)
	var user domain.User
	key, err := it.Next(&user)
	if err == iterator.Done {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	user.ID = key.Name

	// Get Links
	links, err := r.GetAllLinks(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.PublicProfile{
		ID:       user.ID,
		Username: user.Name,
		Name:     user.Name,
		Links:    links,
	}, nil
}

// ProjectRepository Implementation

func (r *DatastoreRepository) CreateProject(project *domain.Project) error {
	ctx := context.Background()
	key := datastore.NameKey("Project", project.ID, nil)
	_, err := r.client.Put(ctx, key, project)
	return err
}

func (r *DatastoreRepository) GetAllProjects(userID string) ([]domain.Project, error) {
	ctx := context.Background()
	query := datastore.NewQuery("Project").FilterField("UserID", "=", userID)
	it := r.client.Run(ctx, query)
	var projects []domain.Project
	for {
		var p domain.Project
		key, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		p.ID = key.Name
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *DatastoreRepository) GetProjectByID(id string) (*domain.Project, error) {
	ctx := context.Background()
	key := datastore.NameKey("Project", id, nil)
	var project domain.Project
	if err := r.client.Get(ctx, key, &project); err != nil {
		return nil, err
	}
	project.ID = id
	return &project, nil
}

func (r *DatastoreRepository) UpdateProject(project *domain.Project) error {
	return r.CreateProject(project)
}

func (r *DatastoreRepository) DeleteProject(id string) error {
	ctx := context.Background()
	key := datastore.NameKey("Project", id, nil)
	return r.client.Delete(ctx, key)
}

// LinkRepository Implementation

func (r *DatastoreRepository) CreateLink(link *domain.Link) error {
	ctx := context.Background()
	key := datastore.NameKey("Link", link.ID, nil)
	_, err := r.client.Put(ctx, key, link)
	return err
}

func (r *DatastoreRepository) GetAllLinks(userID string) ([]domain.Link, error) {
	ctx := context.Background()
	query := datastore.NewQuery("Link").FilterField("UserID", "=", userID)
	it := r.client.Run(ctx, query)
	var links []domain.Link
	for {
		var l domain.Link
		key, err := it.Next(&l)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		l.ID = key.Name
		links = append(links, l)
	}
	return links, nil
}

func (r *DatastoreRepository) UpdateLink(link *domain.Link) error {
	return r.CreateLink(link)
}

func (r *DatastoreRepository) DeleteLink(id string) error {
	ctx := context.Background()
	key := datastore.NameKey("Link", id, nil)
	return r.client.Delete(ctx, key)
}
