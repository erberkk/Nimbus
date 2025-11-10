package helpers

import (
	"context"
	"errors"

	"nimbus-backend/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AccessLevel represents the level of access a user has
type AccessLevel string

const (
	AccessLevelNone  AccessLevel = "none"
	AccessLevelRead  AccessLevel = "read"
	AccessLevelWrite AccessLevel = "write"
	AccessLevelOwner AccessLevel = "owner"
)

// CanUserAccess checks if a user can access a resource with a specific access level
func CanUserAccess(userID string, resourceType string, resourceID string, requiredLevel AccessLevel) (bool, error) {
	resourceOID, err := primitive.ObjectIDFromHex(resourceID)
	if err != nil {
		return false, errors.New("invalid resource ID")
	}

	if resourceType == "file" {
		var file struct {
			UserID     string `bson:"user_id"`
			AccessList []struct {
				UserID     string `bson:"user_id"`
				AccessType string `bson:"access_type"`
			} `bson:"access_list"`
		}

		err = database.FileCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
		}).Decode(&file)

		if err != nil {
			return false, err
		}

		// Check if user is the owner
		if file.UserID == userID {
			return requiredLevel == AccessLevelOwner || requiredLevel == AccessLevelWrite || requiredLevel == AccessLevelRead, nil
		}

		// Check access list - doğrudan erişim kontrolü
		for _, access := range file.AccessList {
			if access.UserID == userID {
				if requiredLevel == AccessLevelRead && (access.AccessType == "read" || access.AccessType == "write") {
					return true, nil
				}
				if requiredLevel == AccessLevelWrite && access.AccessType == "write" {
					return true, nil
				}
			}
		}

		// Eğer doğrudan erişim yoksa, hiyerarşik erişim kontrolü yap
		// Dosyanın ancestors'ını al ve üst klasörlerde erişim ara
		if len(file.AccessList) == 0 {
			// Dosyanın ancestors'ını al
			var fullFile struct {
				Ancestors []primitive.ObjectID `bson:"ancestors"`
			}
			err = database.FileCollection.FindOne(context.Background(), bson.M{
				"_id": resourceOID,
			}).Decode(&fullFile)

			if err == nil && len(fullFile.Ancestors) > 0 {
				// Üst klasörlerde erişim ara
				folderCursor, err := database.FolderCollection.Find(context.Background(), bson.M{
					"_id":                 bson.M{"$in": fullFile.Ancestors},
					"access_list.user_id": userID,
				})
				if err == nil {
					defer folderCursor.Close(context.Background())
					for folderCursor.Next(context.Background()) {
						var folder struct {
							AccessList []struct {
								UserID     string `bson:"user_id"`
								AccessType string `bson:"access_type"`
							} `bson:"access_list"`
						}
						if err := folderCursor.Decode(&folder); err != nil {
							continue
						}
						for _, access := range folder.AccessList {
							if access.UserID == userID {
								if requiredLevel == AccessLevelRead && (access.AccessType == "read" || access.AccessType == "write") {
									return true, nil
								}
								if requiredLevel == AccessLevelWrite && access.AccessType == "write" {
									return true, nil
								}
							}
						}
					}
				}
			}
		}

		return false, nil

	} else if resourceType == "folder" {
		var folder struct {
			UserID     string `bson:"user_id"`
			AccessList []struct {
				UserID     string `bson:"user_id"`
				AccessType string `bson:"access_type"`
			} `bson:"access_list"`
		}

		err = database.FolderCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
		}).Decode(&folder)

		if err != nil {
			return false, err
		}

		// Check if user is the owner
		if folder.UserID == userID {
			return requiredLevel == AccessLevelOwner || requiredLevel == AccessLevelWrite || requiredLevel == AccessLevelRead, nil
		}

		// Check access list - doğrudan erişim kontrolü
		for _, access := range folder.AccessList {
			if access.UserID == userID {
				if requiredLevel == AccessLevelRead && (access.AccessType == "read" || access.AccessType == "write") {
					return true, nil
				}
				if requiredLevel == AccessLevelWrite && access.AccessType == "write" {
					return true, nil
				}
			}
		}

		// Eğer doğrudan erişim yoksa, hiyerarşik erişim kontrolü yap
		// Klasörün ancestors'ını al ve üst klasörlerde erişim ara
		if len(folder.AccessList) == 0 {
			// Klasörün ancestors'ını al
			var fullFolder struct {
				Ancestors []primitive.ObjectID `bson:"ancestors"`
			}
			err = database.FolderCollection.FindOne(context.Background(), bson.M{
				"_id": resourceOID,
			}).Decode(&fullFolder)

			if err == nil && len(fullFolder.Ancestors) > 0 {
				// Üst klasörlerde erişim ara
				folderCursor, err := database.FolderCollection.Find(context.Background(), bson.M{
					"_id":                 bson.M{"$in": fullFolder.Ancestors},
					"access_list.user_id": userID,
				})
				if err == nil {
					defer folderCursor.Close(context.Background())
					for folderCursor.Next(context.Background()) {
						var parentFolder struct {
							AccessList []struct {
								UserID     string `bson:"user_id"`
								AccessType string `bson:"access_type"`
							} `bson:"access_list"`
						}
						if err := folderCursor.Decode(&parentFolder); err != nil {
							continue
						}
						for _, access := range parentFolder.AccessList {
							if access.UserID == userID {
								if requiredLevel == AccessLevelRead && (access.AccessType == "read" || access.AccessType == "write") {
									return true, nil
								}
								if requiredLevel == AccessLevelWrite && access.AccessType == "write" {
									return true, nil
								}
							}
						}
					}
				}
			}
		}

		return false, nil
	}

	return false, errors.New("invalid resource type")
}

// CanUserShare checks if a user can share a resource (requires write access)
func CanUserShare(userID string, resourceType string, resourceID string) (bool, error) {
	return CanUserAccess(userID, resourceType, resourceID, AccessLevelWrite)
}

// CheckFileAccessWithOwnerFallback checks file access with owner fallback
// Returns true if user has required access level OR is the file owner
func CheckFileAccessWithOwnerFallback(userID, fileID, fileOwnerID string, requiredLevel AccessLevel) (bool, error) {
	hasAccess, err := CanUserAccess(userID, "file", fileID, requiredLevel)
	if err != nil {
		return false, err
	}
	
	// If user has access, return true
	if hasAccess {
		return true, nil
	}
	
	// If user doesn't have access, check if they are the owner
	// Owner always has full access
	return fileOwnerID == userID, nil
}

// GetUserAccessLevel returns the access level of a user for a resource
func GetUserAccessLevel(userID string, resourceType string, resourceID string) (AccessLevel, error) {
	resourceOID, err := primitive.ObjectIDFromHex(resourceID)
	if err != nil {
		return AccessLevelNone, errors.New("invalid resource ID")
	}

	if resourceType == "file" {
		var file struct {
			UserID     string `bson:"user_id"`
			AccessList []struct {
				UserID     string `bson:"user_id"`
				AccessType string `bson:"access_type"`
			} `bson:"access_list"`
		}

		err = database.FileCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
		}).Decode(&file)

		if err != nil {
			return AccessLevelNone, err
		}

		// Check if user is the owner
		if file.UserID == userID {
			return AccessLevelOwner, nil
		}

		// Check access list
		for _, access := range file.AccessList {
			if access.UserID == userID {
				if access.AccessType == "write" {
					return AccessLevelWrite, nil
				} else if access.AccessType == "read" {
					return AccessLevelRead, nil
				}
			}
		}

		return AccessLevelNone, nil

	} else if resourceType == "folder" {
		var folder struct {
			UserID     string `bson:"user_id"`
			AccessList []struct {
				UserID     string `bson:"user_id"`
				AccessType string `bson:"access_type"`
			} `bson:"access_list"`
		}

		err = database.FolderCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
		}).Decode(&folder)

		if err != nil {
			return AccessLevelNone, err
		}

		// Check if user is the owner
		if folder.UserID == userID {
			return AccessLevelOwner, nil
		}

		// Check access list
		for _, access := range folder.AccessList {
			if access.UserID == userID {
				if access.AccessType == "write" {
					return AccessLevelWrite, nil
				} else if access.AccessType == "read" {
					return AccessLevelRead, nil
				}
			}
		}

		return AccessLevelNone, nil
	}

	return AccessLevelNone, errors.New("invalid resource type")
}
