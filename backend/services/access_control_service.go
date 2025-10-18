package services

import (
	"context"
	"nimbus-backend/database"
	"nimbus-backend/helpers"
	"nimbus-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccessControlService struct{}

var AccessControlServiceInstance = &AccessControlService{}

// CheckAccess - Check if user has access to resource
func (acs *AccessControlService) CheckAccess(userID, resourceType, resourceID string, level helpers.AccessLevel) (bool, error) {
	return helpers.CanUserAccess(userID, resourceType, resourceID, level)
}

// GetAccessibleUsers - Get users who have access to a resource
func (acs *AccessControlService) GetAccessibleUsers(resourceType, resourceID string) ([]models.User, error) {
	resourceOID, err := primitive.ObjectIDFromHex(resourceID)
	if err != nil {
		return nil, err
	}

	var accessList []struct {
		UserID     string `bson:"user_id"`
		AccessType string `bson:"access_type"`
	}

	if resourceType == "file" {
		var file struct {
			AccessList []struct {
				UserID     string `bson:"user_id"`
				AccessType string `bson:"access_type"`
			} `bson:"access_list"`
		}

		err = database.FileCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
		}).Decode(&file)

		if err != nil {
			return nil, err
		}

		accessList = file.AccessList
	} else if resourceType == "folder" {
		var folder struct {
			AccessList []struct {
				UserID     string `bson:"user_id"`
				AccessType string `bson:"access_type"`
			} `bson:"access_list"`
		}

		err = database.FolderCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
		}).Decode(&folder)

		if err != nil {
			return nil, err
		}

		accessList = folder.AccessList
	}

	// Extract user IDs
	userIDs := make([]string, 0, len(accessList))
	for _, access := range accessList {
		userIDs = append(userIDs, access.UserID)
	}

	// Get users
	return UserServiceInstance.GetUsersByIDs(userIDs)
}

// GetAccessTypeFromList - Get access type for user from access list
func (acs *AccessControlService) GetAccessTypeFromList(accessList []models.AccessEntry, userID string) string {
	for _, access := range accessList {
		if access.UserID == userID {
			return access.AccessType
		}
	}
	return "none"
}

// AddUserAccess - Add user access to resource
func (acs *AccessControlService) AddUserAccess(resourceType, resourceID, userID, accessType string) error {
	resourceOID, err := primitive.ObjectIDFromHex(resourceID)
	if err != nil {
		return err
	}

	accessEntry := models.AccessEntry{
		UserID:     userID,
		AccessType: accessType,
	}

	if resourceType == "file" {
		_, err = database.FileCollection.UpdateOne(context.Background(), bson.M{
			"_id": resourceOID,
		}, bson.M{
			"$addToSet": bson.M{
				"access_list": accessEntry,
			},
		})
	} else if resourceType == "folder" {
		_, err = database.FolderCollection.UpdateOne(context.Background(), bson.M{
			"_id": resourceOID,
		}, bson.M{
			"$addToSet": bson.M{
				"access_list": accessEntry,
			},
		})
	}

	return err
}

// RemoveUserAccess - Remove user access from resource
func (acs *AccessControlService) RemoveUserAccess(resourceType, resourceID, userID string) error {
	resourceOID, err := primitive.ObjectIDFromHex(resourceID)
	if err != nil {
		return err
	}

	if resourceType == "file" {
		_, err = database.FileCollection.UpdateOne(context.Background(), bson.M{
			"_id": resourceOID,
		}, bson.M{
			"$pull": bson.M{
				"access_list": bson.M{
					"user_id": userID,
				},
			},
		})
	} else if resourceType == "folder" {
		_, err = database.FolderCollection.UpdateOne(context.Background(), bson.M{
			"_id": resourceOID,
		}, bson.M{
			"$pull": bson.M{
				"access_list": bson.M{
					"user_id": userID,
				},
			},
		})
	}

	return err
}
