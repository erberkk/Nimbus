package services

import (
	"context"
	"nimbus-backend/database"
	"nimbus-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct{}

var UserServiceInstance = &UserService{}

// GetUserByID - Get user by ID
func (us *UserService) GetUserByID(userID string) (*models.User, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = database.UserCollection.FindOne(context.Background(), bson.M{
		"_id": userOID,
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUsersByIDs - Get multiple users by IDs
func (us *UserService) GetUsersByIDs(userIDs []string) ([]models.User, error) {
	if len(userIDs) == 0 {
		return []models.User{}, nil
	}

	// Convert string IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, 0, len(userIDs))
	for _, userID := range userIDs {
		userOID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			continue // Skip invalid IDs
		}
		objectIDs = append(objectIDs, userOID)
	}

	if len(objectIDs) == 0 {
		return []models.User{}, nil
	}

	cursor, err := database.UserCollection.Find(context.Background(), bson.M{
		"_id": bson.M{"$in": objectIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []models.User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserByGoogleID - Get user by Google ID
func (us *UserService) GetUserByGoogleID(googleID string) (*models.User, error) {
	var user models.User
	err := database.UserCollection.FindOne(context.Background(), bson.M{
		"google_id": googleID,
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser - Create new user
func (us *UserService) CreateUser(user *models.User) error {
	_, err := database.UserCollection.InsertOne(context.Background(), user)
	return err
}

// UpdateUser - Update user
func (us *UserService) UpdateUser(userID string, updateData bson.M) error {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = database.UserCollection.UpdateOne(context.Background(), bson.M{
		"_id": userOID,
	}, bson.M{
		"$set": updateData,
	})

	return err
}

// GetUserResponse - UserID'den UserResponse oluşturur (owner bilgisi için kullanılır)
func (us *UserService) GetUserResponse(userID string) *models.UserResponse {
	if userID == "" {
		return nil
	}

	owner, err := us.GetUserByID(userID)
	if err != nil || owner == nil {
		return nil
	}

	return &models.UserResponse{
		ID:     owner.ID.Hex(),
		Email:  owner.Email,
		Name:   owner.Name,
		Avatar: owner.Avatar,
	}
}
