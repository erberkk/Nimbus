package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"nimbus-backend/database"
	"nimbus-backend/helpers"
	"nimbus-backend/models"
	"nimbus-backend/services"
)

func GetResourceShares() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		resourceID := c.Params("resourceId")

		resourceOID, err := primitive.ObjectIDFromHex(resourceID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid resource ID",
			})
		}

		var file models.File
		err = database.FileCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
			"$or": []bson.M{
				{"user_id": userID},
				{"access_list.user_id": userID},
			},
		}).Decode(&file)

		if err != nil {
			var folder models.Folder
			err = database.FolderCollection.FindOne(context.Background(), bson.M{
				"_id": resourceOID,
				"$or": []bson.M{
					{"user_id": userID},
					{"access_list.user_id": userID},
				},
			}).Decode(&folder)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Resource not found or access denied",
				})
			}

			users, err := services.AccessControlServiceInstance.GetAccessibleUsers("folder", resourceID)
			if err != nil {
				log.Printf("Error getting accessible users: %v", err)
				users = []models.User{}
			}

			sharedUsers := make([]models.UserResponse, 0, len(users))
			for _, user := range users {
				sharedUsers = append(sharedUsers, models.UserResponse{
					ID:    user.ID.Hex(),
					Email: user.Email,
					Name:  user.Name,
				})
			}

			return c.JSON(fiber.Map{
				"resource_id":   resourceOID.Hex(),
				"resource_type": "folder",
				"user_id":       folder.UserID,
				"public_link":   folder.PublicLink,
				"access_list":   folder.AccessList,
				"shared_with":   sharedUsers,
			})
		}

		users, err := services.AccessControlServiceInstance.GetAccessibleUsers("file", resourceID)
		if err != nil {
			log.Printf("Error getting accessible users: %v", err)
			users = []models.User{}
		}

		sharedUsers := make([]models.UserResponse, 0, len(users))
		for _, user := range users {
			sharedUsers = append(sharedUsers, models.UserResponse{
				ID:    user.ID.Hex(),
				Email: user.Email,
				Name:  user.Name,
			})
		}

		return c.JSON(fiber.Map{
			"resource_id":   resourceOID.Hex(),
			"resource_type": "file",
			"user_id":       file.UserID,
			"public_link":   file.PublicLink,
			"access_list":   file.AccessList,
			"shared_with":   sharedUsers,
		})
	}
}

// getRecursiveItemCount - Recursively count all accessible items in a folder
func getRecursiveItemCount(userID string, folderID string) (int64, error) {
	subFolders, err := services.FolderServiceInstance.GetSubFolders(folderID)
	if err != nil {
		return 0, err
	}

	files, err := services.FolderServiceInstance.GetFolderFiles(folderID)
	if err != nil {
		return 0, err
	}

	count := int64(0)

	for _, file := range files {
		if canAccess, _ := helpers.CanUserAccess(userID, "file", file.ID.Hex(), helpers.AccessLevelRead); canAccess {
			count++
		}
	}

	for _, subFolder := range subFolders {
		if canAccess, _ := helpers.CanUserAccess(userID, "folder", subFolder.ID.Hex(), helpers.AccessLevelRead); canAccess {
			count++
			subCount, err := getRecursiveItemCount(userID, subFolder.ID.Hex())
			if err == nil {
				count += subCount
			}
		}
	}

	return count, nil
}

func GetSharedWithMe() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var sharedItems []fiber.Map

		fileCursor, err := database.FileCollection.Find(context.Background(), bson.M{
			"access_list.user_id": userID,
		})
		if err == nil {
			defer fileCursor.Close(context.Background())
			for fileCursor.Next(context.Background()) {
				var file models.File
				if err := fileCursor.Decode(&file); err == nil {
					sharedItems = append(sharedItems, fiber.Map{
						"resource": models.FileResponse{
							ID:               file.ID.Hex(),
							UserID:           file.UserID,
							Filename:         file.Filename,
							Size:             file.Size,
							ContentType:      file.ContentType,
							MinioPath:        file.MinioPath,
							PublicLink:       file.PublicLink,
							AccessList:       file.AccessList,
							ParentID:         file.ParentID,
							Ancestors:        file.Ancestors,
							IsStarred:        file.IsStarred,
							ProcessingStatus: file.ProcessingStatus,
							ProcessingError:  file.ProcessingError,
							ProcessedAt:      file.ProcessedAt,
							ChunkCount:       file.ChunkCount,
							DeletedAt:        file.DeletedAt,
							CreatedAt:        file.CreatedAt,
							UpdatedAt:        file.UpdatedAt,
							Owner:            services.UserServiceInstance.GetUserResponse(file.UserID),
						},
						"access_type":   getAccessTypeFromList(file.AccessList, userID),
						"resource_type": "file",
						"owner":         services.UserServiceInstance.GetUserResponse(file.UserID),
					})
				}
			}
		}

		folderCursor, err := database.FolderCollection.Find(context.Background(), bson.M{
			"access_list.user_id": userID,
		})
		if err == nil {
			defer folderCursor.Close(context.Background())
			for folderCursor.Next(context.Background()) {
				var folder models.Folder
				if err := folderCursor.Decode(&folder); err == nil {
					count, err := getRecursiveItemCount(userID, folder.ID.Hex())
					if err != nil {
						log.Printf("Recursive item count hesaplama hatası: %v", err)
						count = 0
					}

					size, _ := services.FolderServiceInstance.GetFolderSize(folder.ID.Hex())
					sharedItems = append(sharedItems, fiber.Map{
						"resource": models.FolderResponse{
							ID:        folder.ID.Hex(),
							Name:      folder.Name,
							Color:     folder.Color,
							ItemCount: int(count),
							Size:      size,
							CreatedAt: folder.CreatedAt,
							UpdatedAt: folder.UpdatedAt,
							Owner:     services.UserServiceInstance.GetUserResponse(folder.UserID),
						},
						"access_type":   services.AccessControlServiceInstance.GetAccessTypeFromList(folder.AccessList, userID),
						"resource_type": "folder",
						"owner":         services.UserServiceInstance.GetUserResponse(folder.UserID),
					})
				}
			}
		}

		return c.JSON(sharedItems)
	}
}

// GetSharedFolderContents - Paylaşılan klasörün içeriğini getir (alt klasörler ve dosyalar)
func GetSharedFolderContents() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("folderId")
		if folderID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "folder ID parametresi gerekli",
			})
		}

		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		canAccess, err := helpers.CanUserAccess(userID, "folder", folderID, helpers.AccessLevelRead)
		if err != nil || !canAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu klasöre erişim yetkiniz yok",
			})
		}

		subFolders, err := services.FolderServiceInstance.GetSubFolders(folderID)
		if err != nil {
			log.Printf("Alt klasörleri alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Alt klasörler listelenemedi",
			})
		}

		accessibleSubFolders := make([]fiber.Map, 0)
		for _, subFolder := range subFolders {
			canAccessSub, err := helpers.CanUserAccess(userID, "folder", subFolder.ID.Hex(), helpers.AccessLevelRead)
			if err == nil && canAccessSub {
				accessType := getAccessTypeFromList(subFolder.AccessList, userID)

				count, err := services.FolderServiceInstance.GetFolderItemCount(subFolder.ID.Hex())
				if err != nil {
					log.Printf("Subfolder item count hesaplama hatası: %v", err)
					count = 0
				}

				accessibleSubFolders = append(accessibleSubFolders, fiber.Map{
					"folder": models.FolderResponse{
						ID:        subFolder.ID.Hex(),
						Name:      subFolder.Name,
						Color:     subFolder.Color,
						ItemCount: int(count),
						FolderID:  subFolder.FolderID,
						CreatedAt: subFolder.CreatedAt,
						UpdatedAt: subFolder.UpdatedAt,
						Owner:     services.UserServiceInstance.GetUserResponse(subFolder.UserID),
					},
					"access_type": accessType,
					"is_shared":   true,
				})
			}
		}

		files, err := services.FolderServiceInstance.GetFolderFiles(folderID)
		if err != nil {
			log.Printf("Klasör dosyaları alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosyalar listelenemedi",
			})
		}

		accessibleFiles := make([]fiber.Map, 0)
		for _, file := range files {
			canAccessFile, err := helpers.CanUserAccess(userID, "file", file.ID.Hex(), helpers.AccessLevelRead)
			if err == nil && canAccessFile {
				accessType := getAccessTypeFromList(file.AccessList, userID)

				accessibleFiles = append(accessibleFiles, fiber.Map{
					"file": models.FileResponse{
						ID:               file.ID.Hex(),
						UserID:           file.UserID,
						Filename:         file.Filename,
						Size:             file.Size,
						ContentType:      file.ContentType,
						MinioPath:        file.MinioPath,
						PublicLink:       file.PublicLink,
						AccessList:       file.AccessList,
						ParentID:         file.ParentID,
						Ancestors:        file.Ancestors,
						IsStarred:        file.IsStarred,
						ProcessingStatus: file.ProcessingStatus,
						ProcessingError:  file.ProcessingError,
						ProcessedAt:      file.ProcessedAt,
						ChunkCount:       file.ChunkCount,
						DeletedAt:        file.DeletedAt,
						CreatedAt:        file.CreatedAt,
						UpdatedAt:        file.UpdatedAt,
						Owner:            services.UserServiceInstance.GetUserResponse(file.UserID),
					},
					"access_type": accessType,
					"owner":       services.UserServiceInstance.GetUserResponse(file.UserID),
					"is_shared":   true,
				})
			}
		}

		folderList := make([]fiber.Map, 0, len(accessibleSubFolders))
		for _, subFolderData := range accessibleSubFolders {
			subFolder := subFolderData["folder"].(models.FolderResponse)
			count, err := services.FolderServiceInstance.GetFolderItemCount(subFolder.ID)
			if err != nil {
				log.Printf("Nested shared folder item count hesaplama hatası: %v", err)
				count = 0
			}
			subFolder.ItemCount = int(count)
			size, _ := services.FolderServiceInstance.GetFolderSize(subFolder.ID)
			subFolder.Size = size
			folderList = append(folderList, fiber.Map{
				"resource":      subFolder,
				"access_type":   subFolderData["access_type"],
				"resource_type": "folder",
				"owner":         subFolderData["owner"],
			})
		}

		fileList := make([]fiber.Map, 0, len(accessibleFiles))
		for _, fileData := range accessibleFiles {
			fileList = append(fileList, fiber.Map{
				"resource":      fileData["file"],
				"access_type":   fileData["access_type"],
				"resource_type": "file",
				"owner":         fileData["owner"],
			})
		}

		size, _ := services.FolderServiceInstance.GetFolderSize(folderID)
		return c.JSON(fiber.Map{
			"folder": models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: len(accessibleSubFolders) + len(accessibleFiles),
				Size:      size,
				FolderID:  folder.FolderID,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
				Owner:     services.UserServiceInstance.GetUserResponse(folder.UserID),
			},
			"folders": folderList,
			"files":   fileList,
			"count":   len(folderList) + len(fileList),
		})
	}
}

func UpdateAccessPermission() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var req struct {
			UserID     string `json:"user_id" validate:"required"`
			Permission string `json:"permission" validate:"required"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		resourceID := c.Params("resourceId")
		resourceOID, err := primitive.ObjectIDFromHex(resourceID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid resource ID",
			})
		}

		canShare, err := helpers.CanUserShare(userID, "file", resourceID)
		if err != nil || !canShare {
			canShare, err = helpers.CanUserShare(userID, "folder", resourceID)
			if err != nil || !canShare {
				fmt.Printf("DEBUG: User %s cannot share this resource\n", userID)
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You don't have permission to modify access for this resource",
				})
			}
		}

		var updateResult *mongo.UpdateResult
		var accessEntry models.AccessEntry

		fileUpdateResult, err := database.FileCollection.UpdateOne(
			context.Background(),
			bson.M{
				"_id":                 resourceOID,
				"access_list.user_id": req.UserID,
			},
			bson.M{
				"$set": bson.M{
					"access_list.$.access_type": req.Permission,
					"access_list.$.granted_at":  time.Now(),
					"access_list.$.granted_by":  userID,
					"updated_at":                time.Now(),
				},
			},
		)

		if err != nil {
			// File update error - continue to folder update
		}

		if fileUpdateResult == nil || fileUpdateResult.MatchedCount == 0 {
			accessEntry = models.AccessEntry{
				UserID:     req.UserID,
				AccessType: req.Permission,
				GrantedAt:  time.Now(),
				GrantedBy:  userID,
			}

			database.FileCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":         resourceOID,
					"access_list": nil,
				},
				bson.M{
					"$set": bson.M{"access_list": []models.AccessEntry{}},
				},
			)

			fileUpdateResult, err = database.FileCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id": resourceOID,
				},
				bson.M{
					"$push": bson.M{"access_list": accessEntry},
					"$set":  bson.M{"updated_at": time.Now()},
				},
			)

			if err != nil {
				// File push error - continue to folder update
			}
		}

		if fileUpdateResult == nil || fileUpdateResult.MatchedCount == 0 {

			folderUpdateResult, err := database.FolderCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":                 resourceOID,
					"access_list.user_id": req.UserID,
				},
				bson.M{
					"$set": bson.M{
						"access_list.$.access_type": req.Permission,
						"access_list.$.granted_at":  time.Now(),
						"access_list.$.granted_by":  userID,
						"updated_at":                time.Now(),
					},
				},
			)

			if err != nil {
				// Folder update error - continue
			}

			if folderUpdateResult == nil || folderUpdateResult.MatchedCount == 0 {
				database.FolderCollection.UpdateOne(
					context.Background(),
					bson.M{
						"_id":         resourceOID,
						"access_list": nil,
					},
					bson.M{
						"$set": bson.M{"access_list": []models.AccessEntry{}},
					},
				)

				updateResult, err = database.FolderCollection.UpdateOne(
					context.Background(),
					bson.M{
						"_id": resourceOID,
					},
					bson.M{
						"$push": bson.M{"access_list": accessEntry},
						"$set":  bson.M{"updated_at": time.Now()},
					},
				)

				if err != nil {
					// Folder push error - continue
				}
			} else {
				updateResult = folderUpdateResult
			}
		}

		if updateResult == nil {
			updateResult = fileUpdateResult
		}

		if err != nil || updateResult == nil || updateResult.MatchedCount == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource or access entry not found",
			})
		}

		if req.Permission != "none" {
			err = helpers.PropagateAccessToChildren(resourceOID, req.UserID, req.Permission, userID)
			if err != nil {
				// Propagation başarısız olsa da ana işlem başarılı, devam et
			}
		}

		return c.JSON(fiber.Map{
			"message": "Access permission updated successfully",
		})
	}
}

func RemoveUserAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		userIDToRemove := c.Params("userId")

		resourceID := c.Params("resourceId")
		resourceOID, err := primitive.ObjectIDFromHex(resourceID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid resource ID",
			})
		}

		canShare, err := helpers.CanUserShare(userID, "file", resourceID)
		if err != nil || !canShare {
			canShare, err = helpers.CanUserShare(userID, "folder", resourceID)
			if err != nil || !canShare {
				fmt.Printf("DEBUG: User %s cannot share this resource\n", userID)
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You don't have permission to modify access for this resource",
				})
			}
		}

		updateResult, err := database.FileCollection.UpdateOne(
			context.Background(),
			bson.M{
				"_id":                 resourceOID,
				"access_list.user_id": userIDToRemove,
			},
			bson.M{
				"$pull": bson.M{
					"access_list": bson.M{"user_id": userIDToRemove},
				},
				"$set": bson.M{"updated_at": time.Now()},
			},
		)

		if err != nil {
			// File removal error - continue to folder removal
		}

		if updateResult.MatchedCount == 0 {
			updateResult, err = database.FolderCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":                 resourceOID,
					"access_list.user_id": userIDToRemove,
				},
				bson.M{
					"$pull": bson.M{
						"access_list": bson.M{"user_id": userIDToRemove},
					},
					"$set": bson.M{"updated_at": time.Now()},
				},
			)

			if err != nil {
				// Folder removal error - continue
			}
		}

		if err != nil || updateResult == nil || updateResult.MatchedCount == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource or access entry not found",
			})
		}

		err = helpers.RemoveAccessFromChildren(resourceOID, userIDToRemove)
		if err != nil {
			// Propagation başarısız olsa da ana işlem başarılı, devam et
		}

		return c.JSON(fiber.Map{
			"message": "User access removed successfully",
		})
	}
}

func SearchUsers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		query := c.Query("q")
		if len(query) < 2 {
			return c.JSON([]models.UserResponse{})
		}

		userOID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid user ID format",
			})
		}

		cursor, err := database.UserCollection.Find(context.Background(), bson.M{
			"email": bson.M{"$regex": query, "$options": "i"},
			"_id":   bson.M{"$ne": userOID},
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search users",
			})
		}
		defer cursor.Close(context.Background())

		var users []models.UserResponse
		for cursor.Next(context.Background()) {
			var user models.User
			if err := cursor.Decode(&user); err != nil {
				continue
			}

			users = append(users, models.UserResponse{
				ID:    user.ID.Hex(),
				Email: user.Email,
				Name:  user.Name,
			})
		}

		return c.JSON(users)
	}
}

// Helper function to get access type from access list
func getAccessTypeFromList(accessList []models.AccessEntry, userID string) string {
	for _, access := range accessList {
		if access.UserID == userID {
			return access.AccessType
		}
	}
	return "read" // default
}

// GetResourceByPublicLink - Public link ile resource'a erişim sağla ve kullanıcıyı otomatik ekle
func GetResourceByPublicLink() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Giriş yapmanız gerekiyor",
			})
		}

		publicLink := c.Params("publicLink")
		if publicLink == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Geçersiz public link",
			})
		}

		ctx := context.Background()

		var file models.File
		err = database.FileCollection.FindOne(ctx, bson.M{"public_link": publicLink}).Decode(&file)
		if err == nil {
			accessEntry := models.AccessEntry{
				UserID:     userID,
				AccessType: "read",
				GrantedAt:  time.Now(),
				GrantedBy:  file.UserID,
			}

			existingAccess := false
			for _, access := range file.AccessList {
				if access.UserID == userID {
					existingAccess = true
					break
				}
			}

			if !existingAccess {
				update := bson.M{
					"$push": bson.M{"access_list": accessEntry},
					"$set":  bson.M{"updated_at": time.Now()},
				}
				_, err = database.FileCollection.UpdateOne(ctx, bson.M{"_id": file.ID}, update)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Kullanıcı erişim listesine eklenemedi",
					})
				}
			}

			return c.JSON(fiber.Map{
				"resource": models.FileResponse{
					ID:          file.ID.Hex(),
					Filename:    file.Filename,
					Size:        file.Size,
					ContentType: file.ContentType,
					PublicLink:  file.PublicLink,
					AccessList:  file.AccessList,
					CreatedAt:   file.CreatedAt,
					UpdatedAt:   file.UpdatedAt,
				},
				"resource_type": "file",
			})
		}

		var folder models.Folder
		err = database.FolderCollection.FindOne(ctx, bson.M{"public_link": publicLink}).Decode(&folder)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Geçersiz public link veya resource bulunamadı",
			})
		}

		accessEntry := models.AccessEntry{
			UserID:     userID,
			AccessType: "read",
			GrantedAt:  time.Now(),
			GrantedBy:  folder.UserID,
		}

		existingAccess := false
		for _, access := range folder.AccessList {
			if access.UserID == userID {
				existingAccess = true
				break
			}
		}

		if !existingAccess {
			update := bson.M{
				"$push": bson.M{"access_list": accessEntry},
				"$set":  bson.M{"updated_at": time.Now()},
			}
			_, err = database.FolderCollection.UpdateOne(ctx, bson.M{"_id": folder.ID}, update)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Kullanıcı erişim listesine eklenemedi",
				})
			}
		}

		return c.JSON(fiber.Map{
			"resource": models.FolderResponse{
				ID:         folder.ID.Hex(),
				Name:       folder.Name,
				Color:      folder.Color,
				PublicLink: folder.PublicLink,
				ItemCount:  0,
				AccessList: folder.AccessList,
				FolderID:   folder.FolderID,
				CreatedAt:  folder.CreatedAt,
				UpdatedAt:  folder.UpdatedAt,
			},
			"resource_type": "folder",
		})
	}
}
