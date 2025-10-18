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

		// Check if resource exists and user has access (owner or in access_list)
		var file models.File
		err = database.FileCollection.FindOne(context.Background(), bson.M{
			"_id": resourceOID,
			"$or": []bson.M{
				{"user_id": userID},             // Owner
				{"access_list.user_id": userID}, // Has access
			},
		}).Decode(&file)

		if err != nil {
			var folder models.Folder
			err = database.FolderCollection.FindOne(context.Background(), bson.M{
				"_id": resourceOID,
				"$or": []bson.M{
					{"user_id": userID},             // Owner
					{"access_list.user_id": userID}, // Has access
				},
			}).Decode(&folder)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Resource not found or access denied",
				})
			}

			// For folders, return access list info
			sharedUsers := []models.UserResponse{}
			for _, access := range folder.AccessList {
				var user models.User
				userOID, err := primitive.ObjectIDFromHex(access.UserID)
				if err != nil {
					continue
				}
				userErr := database.UserCollection.FindOne(context.Background(), bson.M{
					"_id": userOID,
				}).Decode(&user)
				if userErr == nil {
					sharedUsers = append(sharedUsers, models.UserResponse{
						ID:    user.ID.Hex(),
						Email: user.Email,
						Name:  user.Name,
					})
				}
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

		// For files, return access list info
		sharedUsers := []models.UserResponse{}
		for _, access := range file.AccessList {
			var user models.User
			userOID, err := primitive.ObjectIDFromHex(access.UserID)
			if err != nil {
				continue
			}
			userErr := database.UserCollection.FindOne(context.Background(), bson.M{
				"_id": userOID,
			}).Decode(&user)
			if userErr == nil {
				sharedUsers = append(sharedUsers, models.UserResponse{
					ID:    user.ID.Hex(),
					Email: user.Email,
					Name:  user.Name,
				})
			}
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

func GetSharedWithMe() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var sharedItems []fiber.Map

		// Get files shared with this user (from access_list)
		fileCursor, err := database.FileCollection.Find(context.Background(), bson.M{
			"access_list.user_id": userID,
		})
		if err == nil {
			defer fileCursor.Close(context.Background())
			for fileCursor.Next(context.Background()) {
				var file models.File
				if err := fileCursor.Decode(&file); err == nil {
					// Get owner info
					var owner models.User
					database.UserCollection.FindOne(context.Background(), bson.M{
						"_id": file.UserID,
					}).Decode(&owner)

					sharedItems = append(sharedItems, fiber.Map{
						"resource": models.FileResponse{
							ID:          file.ID.Hex(),
							Filename:    file.Filename,
							Size:        file.Size,
							ContentType: file.ContentType,
							CreatedAt:   file.CreatedAt,
							UpdatedAt:   file.UpdatedAt,
						},
						"access_type":   getAccessTypeFromList(file.AccessList, userID),
						"resource_type": "file",
						"owner": models.UserResponse{
							ID:     owner.ID.Hex(),
							Email:  owner.Email,
							Name:   owner.Name,
							Avatar: owner.Avatar,
						},
					})
				}
			}
		}

		// Get folders shared with this user (from access_list)
		folderCursor, err := database.FolderCollection.Find(context.Background(), bson.M{
			"access_list.user_id": userID,
		})
		if err == nil {
			defer folderCursor.Close(context.Background())
			for folderCursor.Next(context.Background()) {
				var folder models.Folder
				if err := folderCursor.Decode(&folder); err == nil {
					// Get owner info
					var owner models.User
					database.UserCollection.FindOne(context.Background(), bson.M{
						"_id": folder.UserID,
					}).Decode(&owner)

					// Calculate item count for the shared folder (only accessible items)
					// Since we're already filtering accessible items, we can calculate based on that
					subFolders, _ := services.FolderServiceInstance.GetSubFolders(folder.ID.Hex())
					files, _ := services.FolderServiceInstance.GetFolderFiles(folder.ID.Hex())

					// Count only accessible items
					accessibleSubCount := 0
					for _, subFolder := range subFolders {
						if canAccess, _ := helpers.CanUserAccess(userID, "folder", subFolder.ID.Hex(), helpers.AccessLevelRead); canAccess {
							accessibleSubCount++
						}
					}

					accessibleFileCount := 0
					for _, file := range files {
						if canAccess, _ := helpers.CanUserAccess(userID, "file", file.ID.Hex(), helpers.AccessLevelRead); canAccess {
							accessibleFileCount++
						}
					}

					var count int64 = int64(accessibleSubCount + accessibleFileCount)

					sharedItems = append(sharedItems, fiber.Map{
						"resource": models.FolderResponse{
							ID:        folder.ID.Hex(),
							Name:      folder.Name,
							Color:     folder.Color,
							ItemCount: int(count),
							CreatedAt: folder.CreatedAt,
							UpdatedAt: folder.UpdatedAt,
						},
						"access_type":   getAccessTypeFromList(folder.AccessList, userID),
						"resource_type": "folder",
						"owner": models.UserResponse{
							ID:     owner.ID.Hex(),
							Email:  owner.Email,
							Name:   owner.Name,
							Avatar: owner.Avatar,
						},
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

		// Klasör bilgisini getir ve erişim kontrolü yap
		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		// Kullanıcının bu klasöre erişimi var mı kontrol et (owner veya access list)
		canAccess, err := helpers.CanUserAccess(userID, "folder", folderID, helpers.AccessLevelRead)
		if err != nil || !canAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu klasöre erişim yetkiniz yok",
			})
		}

		// Klasördeki alt klasörleri getir
		subFolders, err := services.FolderServiceInstance.GetSubFolders(folderID)
		if err != nil {
			log.Printf("Alt klasörleri alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Alt klasörler listelenemedi",
			})
		}

		// Sadece erişimi olan alt klasörleri filtrele ve erişim bilgilerini ekle
		accessibleSubFolders := make([]fiber.Map, 0)
		for _, subFolder := range subFolders {
			canAccessSub, err := helpers.CanUserAccess(userID, "folder", subFolder.ID.Hex(), helpers.AccessLevelRead)
			if err == nil && canAccessSub {
				// Alt klasör için erişim bilgilerini al
				accessType := getAccessTypeFromList(subFolder.AccessList, userID)

				// Sahibi bilgilerini al
				var owner models.User
				database.UserCollection.FindOne(context.Background(), bson.M{
					"_id": subFolder.UserID,
				}).Decode(&owner)

				// Calculate real count for this accessible subfolder
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
					},
					"access_type": accessType,
					"owner": models.UserResponse{
						ID:     owner.ID.Hex(),
						Email:  owner.Email,
						Name:   owner.Name,
						Avatar: owner.Avatar,
					},
					"is_shared": true,
				})
			}
		}

		// Klasördeki dosyaları getir
		files, err := services.FolderServiceInstance.GetFolderFiles(folderID)
		if err != nil {
			log.Printf("Klasör dosyaları alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosyalar listelenemedi",
			})
		}

		// Sadece erişimi olan dosyaları filtrele ve erişim bilgilerini ekle
		accessibleFiles := make([]fiber.Map, 0)
		for _, file := range files {
			canAccessFile, err := helpers.CanUserAccess(userID, "file", file.ID.Hex(), helpers.AccessLevelRead)
			if err == nil && canAccessFile {
				// Dosya için erişim bilgilerini al
				accessType := getAccessTypeFromList(file.AccessList, userID)

				// Sahibi bilgilerini al
				var owner models.User
				database.UserCollection.FindOne(context.Background(), bson.M{
					"_id": file.UserID,
				}).Decode(&owner)

				accessibleFiles = append(accessibleFiles, fiber.Map{
					"file": models.FileResponse{
						ID:          file.ID.Hex(),
						Filename:    file.Filename,
						Size:        file.Size,
						ContentType: file.ContentType,
						CreatedAt:   file.CreatedAt,
						UpdatedAt:   file.UpdatedAt,
					},
					"access_type": accessType,
					"owner": models.UserResponse{
						ID:     owner.ID.Hex(),
						Email:  owner.Email,
						Name:   owner.Name,
						Avatar: owner.Avatar,
					},
					"is_shared": true,
				})
			}
		}

		// Klasör response'ları formatla
		folderList := make([]fiber.Map, 0, len(accessibleSubFolders))
		for _, subFolderData := range accessibleSubFolders {
			subFolder := subFolderData["folder"].(models.FolderResponse)
			// Calculate real count for nested shared folders too
			count, err := services.FolderServiceInstance.GetFolderItemCount(subFolder.ID)
			if err != nil {
				log.Printf("Nested shared folder item count hesaplama hatası: %v", err)
				count = 0
			}
			subFolder.ItemCount = int(count)
			folderList = append(folderList, fiber.Map{
				"resource":      subFolder,
				"access_type":   subFolderData["access_type"],
				"resource_type": "folder",
				"owner":         subFolderData["owner"],
			})
		}

		// Dosya response'ları formatla
		fileList := make([]fiber.Map, 0, len(accessibleFiles))
		for _, fileData := range accessibleFiles {
			fileList = append(fileList, fiber.Map{
				"resource":      fileData["file"],
				"access_type":   fileData["access_type"],
				"resource_type": "file",
				"owner":         fileData["owner"],
			})
		}

		return c.JSON(fiber.Map{
			"folder": models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: len(accessibleSubFolders) + len(accessibleFiles),
				FolderID:  folder.FolderID,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
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

		// Check if user can share this resource (requires write access or owner)
		canShare, err := helpers.CanUserShare(userID, "file", resourceID)
		if err != nil || !canShare {
			// Try folder
			canShare, err = helpers.CanUserShare(userID, "folder", resourceID)
			if err != nil || !canShare {
				fmt.Printf("DEBUG: User %s cannot share this resource\n", userID)
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You don't have permission to modify access for this resource",
				})
			}
		}

		// Initialize variables
		var updateResult *mongo.UpdateResult
		var accessEntry models.AccessEntry

		// First try to update existing access entry for this user
		fmt.Printf("DEBUG: Attempting to update existing access entry for user %s\n", req.UserID)

		// Try to update existing entry first
		fileUpdateResult, err := database.FileCollection.UpdateOne(
			context.Background(),
			bson.M{
				"_id":                 resourceOID,
				"user_id":             userID,
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
			fmt.Printf("DEBUG: File update error: %v\n", err)
		} else {
			fmt.Printf("DEBUG: File update result - Matched: %d, Modified: %d\n", fileUpdateResult.MatchedCount, fileUpdateResult.ModifiedCount)
		}

		// If no existing entry was updated, add new one
		if fileUpdateResult == nil || fileUpdateResult.MatchedCount == 0 {
			accessEntry = models.AccessEntry{
				UserID:     req.UserID,
				AccessType: req.Permission,
				GrantedAt:  time.Now(),
				GrantedBy:  userID,
			}

			// First, ensure access_list field exists (if null, set to empty array)
			database.FileCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":         resourceOID,
					"user_id":     userID,
					"access_list": nil,
				},
				bson.M{
					"$set": bson.M{"access_list": []models.AccessEntry{}},
				},
			)

			fmt.Printf("DEBUG: Adding new access entry for user %s\n", req.UserID)
			fileUpdateResult, err = database.FileCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":     resourceOID,
					"user_id": userID,
				},
				bson.M{
					"$push": bson.M{"access_list": accessEntry},
					"$set":  bson.M{"updated_at": time.Now()},
				},
			)

			if err != nil {
				fmt.Printf("DEBUG: File push error: %v\n", err)
			} else {
				fmt.Printf("DEBUG: File push result - Matched: %d, Modified: %d\n", fileUpdateResult.MatchedCount, fileUpdateResult.ModifiedCount)
			}
		}

		// If file update didn't work, try folder
		if fileUpdateResult == nil || fileUpdateResult.MatchedCount == 0 {
			fmt.Printf("DEBUG: File update didn't work, trying folder\n")

			// Try to update existing folder entry first
			folderUpdateResult, err := database.FolderCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":                 resourceOID,
					"user_id":             userID,
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
				fmt.Printf("DEBUG: Folder update error: %v\n", err)
			} else {
				fmt.Printf("DEBUG: Folder update result - Matched: %d, Modified: %d\n", folderUpdateResult.MatchedCount, folderUpdateResult.ModifiedCount)
			}

			// If no existing folder entry was updated, add new one
			if folderUpdateResult == nil || folderUpdateResult.MatchedCount == 0 {
				// Ensure access_list field exists for folder too
				database.FolderCollection.UpdateOne(
					context.Background(),
					bson.M{
						"_id":         resourceOID,
						"user_id":     userID,
						"access_list": nil,
					},
					bson.M{
						"$set": bson.M{"access_list": []models.AccessEntry{}},
					},
				)

				fmt.Printf("DEBUG: Adding new folder access entry for user %s\n", req.UserID)
				updateResult, err = database.FolderCollection.UpdateOne(
					context.Background(),
					bson.M{
						"_id":     resourceOID,
						"user_id": userID,
					},
					bson.M{
						"$push": bson.M{"access_list": accessEntry},
						"$set":  bson.M{"updated_at": time.Now()},
					},
				)

				if err != nil {
					fmt.Printf("DEBUG: Folder push error: %v\n", err)
				} else {
					fmt.Printf("DEBUG: Folder push result - Matched: %d, Modified: %d\n", updateResult.MatchedCount, updateResult.ModifiedCount)
				}
			} else {
				updateResult = folderUpdateResult
			}
		}

		// Set updateResult to fileUpdateResult if folder update didn't work
		if updateResult == nil {
			updateResult = fileUpdateResult
		}

		if err != nil || updateResult == nil || updateResult.MatchedCount == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource or access entry not found",
			})
		}

		// Hiyerarşik propagation - tüm alt öğeleri güncelle
		if req.Permission != "none" {
			// Yeni erişim ekleme veya güncelleme - tüm alt öğeleri de güncelle
			err = helpers.PropagateAccessToChildren(resourceOID, req.UserID, req.Permission, userID)
			if err != nil {
				fmt.Printf("DEBUG: Propagation error: %v\n", err)
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
		fmt.Printf("DEBUG: RemoveUserAccess called\n")

		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			fmt.Printf("DEBUG: Failed to get current user ID: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		fmt.Printf("DEBUG: Current user ID: %s\n", userID)

		// Get userId from URL parameter instead of request body
		userIDToRemove := c.Params("userId")
		fmt.Printf("DEBUG: Removing user ID: %s\n", userIDToRemove)

		resourceID := c.Params("resourceId")
		resourceOID, err := primitive.ObjectIDFromHex(resourceID)
		if err != nil {
			fmt.Printf("DEBUG: Failed to parse resource ID: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid resource ID",
			})
		}
		fmt.Printf("DEBUG: Resource ID: %s\n", resourceID)

		// Check if user can share this resource (requires write access or owner)
		canShare, err := helpers.CanUserShare(userID, "file", resourceID)
		if err != nil || !canShare {
			// Try folder
			canShare, err = helpers.CanUserShare(userID, "folder", resourceID)
			if err != nil || !canShare {
				fmt.Printf("DEBUG: User %s cannot share this resource\n", userID)
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You don't have permission to modify access for this resource",
				})
			}
		}

		// Remove user from file's access_list
		fmt.Printf("DEBUG: Attempting to remove user from file access_list\n")
		updateResult, err := database.FileCollection.UpdateOne(
			context.Background(),
			bson.M{
				"_id":                 resourceOID,
				"user_id":             userID,
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
			fmt.Printf("DEBUG: File removal error: %v\n", err)
		} else {
			fmt.Printf("DEBUG: File removal result - Matched: %d, Modified: %d\n", updateResult.MatchedCount, updateResult.ModifiedCount)
		}

		if updateResult.MatchedCount == 0 {
			// Try removing from folder's access_list
			fmt.Printf("DEBUG: File removal didn't work, trying folder\n")
			updateResult, err = database.FolderCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id":                 resourceOID,
					"user_id":             userID,
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
				fmt.Printf("DEBUG: Folder removal error: %v\n", err)
			} else {
				fmt.Printf("DEBUG: Folder removal result - Matched: %d, Modified: %d\n", updateResult.MatchedCount, updateResult.ModifiedCount)
			}
		}

		if err != nil || updateResult == nil || updateResult.MatchedCount == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource or access entry not found",
			})
		}

		// Hiyerarşik propagation - tüm alt öğelerden de kullanıcıyı çıkar
		err = helpers.RemoveAccessFromChildren(resourceOID, userIDToRemove)
		if err != nil {
			fmt.Printf("DEBUG: Remove propagation error: %v\n", err)
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

		// Önce file'larda ara
		var file models.File
		err = database.FileCollection.FindOne(ctx, bson.M{"public_link": publicLink}).Decode(&file)
		if err == nil {
			// File bulundu, kullanıcıyı access list'e ekle
			accessEntry := models.AccessEntry{
				UserID:     userID,
				AccessType: "read",
				GrantedAt:  time.Now(),
				GrantedBy:  file.UserID,
			}

			// Kullanıcı zaten access list'te var mı kontrol et
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

		// File bulunamadı, folder'larda ara
		var folder models.Folder
		err = database.FolderCollection.FindOne(ctx, bson.M{"public_link": publicLink}).Decode(&folder)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Geçersiz public link veya resource bulunamadı",
			})
		}

		// Folder bulundu, kullanıcıyı access list'e ekle
		accessEntry := models.AccessEntry{
			UserID:     userID,
			AccessType: "read",
			GrantedAt:  time.Now(),
			GrantedBy:  folder.UserID,
		}

		// Kullanıcı zaten access list'te var mı kontrol et
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
				ItemCount:  0, // Bu daha sonra hesaplanacak
				AccessList: folder.AccessList,
				FolderID:   folder.FolderID,
				CreatedAt:  folder.CreatedAt,
				UpdatedAt:  folder.UpdatedAt,
			},
			"resource_type": "folder",
		})
	}
}
