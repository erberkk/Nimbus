package services

import (
	"context"
	"fmt"
	"time"

	"nimbus-backend/database"
	"nimbus-backend/helpers"
	"nimbus-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationService struct{}

var ConversationServiceInstance = &ConversationService{}

// GetOrCreateConversation gets existing conversation for a file or creates a new one
func (s *ConversationService) GetOrCreateConversation(userID, fileID string) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to find existing conversation
	var conversation models.Conversation
	err := database.ConversationCollection.FindOne(ctx, bson.M{
		"user_id": userID,
		"file_id": fileID,
	}).Decode(&conversation)

	if err == nil {
		// Found existing conversation
		return &conversation, nil
	}

	// Create new conversation
	conversation = models.Conversation{
		ID:        primitive.NewObjectID(),
		FileID:    fileID,
		UserID:    userID,
		Messages:  []models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = database.ConversationCollection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &conversation, nil
}

// AddMessage adds a new message to the conversation
func (s *ConversationService) AddMessage(userID, fileID string, message models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure message has timestamp
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Update or insert conversation
	filter := bson.M{
		"user_id": userID,
		"file_id": fileID,
	}

	update := bson.M{
		"$push": bson.M{"messages": message},
		"$set":  bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := database.ConversationCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}

	return nil
}

// GetConversation retrieves the conversation for a specific file
func (s *ConversationService) GetConversation(userID, fileID string) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var conversation models.Conversation
	err := database.ConversationCollection.FindOne(ctx, bson.M{
		"user_id": userID,
		"file_id": fileID,
	}).Decode(&conversation)

	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	return &conversation, nil
}

// DeleteConversationsByFileID deletes all conversations for a specific file
func (s *ConversationService) DeleteConversationsByFileID(fileID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.ConversationCollection.DeleteMany(ctx, bson.M{
		"file_id": fileID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete conversations: %w", err)
	}

	return nil
}

// ClearConversation removes all messages from a conversation (but keeps the conversation record)
func (s *ConversationService) ClearConversation(userID, fileID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"messages":   []models.Message{},
			"updated_at": time.Now(),
		},
	}

	_, err := database.ConversationCollection.UpdateOne(ctx, bson.M{
		"user_id": userID,
		"file_id": fileID,
	}, update)

	if err != nil {
		return fmt.Errorf("failed to clear conversation: %w", err)
	}

	return nil
}

// GetUserConversations retrieves all conversations for a user with file information
func (s *ConversationService) GetUserConversations(userID string) ([]models.ConversationWithFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all conversations for the user, ordered by updated_at descending
	cursor, err := database.ConversationCollection.Find(ctx, bson.M{
		"user_id": userID,
	}, options.Find().SetSort(bson.M{"updated_at": -1}))

	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer cursor.Close(ctx)

	var conversations []models.Conversation
	if err = cursor.All(ctx, &conversations); err != nil {
		return nil, fmt.Errorf("failed to decode conversations: %w", err)
	}

	// Get file information for each conversation
	conversationsWithFile := make([]models.ConversationWithFile, 0, len(conversations))
	for _, conv := range conversations {
		// Get file information
		file, err := FileServiceInstance.GetFileByID(conv.FileID)
		if err != nil {
			// Skip if file not found (might be deleted)
			continue
		}

		// Check if user still has access to the file
		hasAccess, err := helpers.CanUserAccess(userID, "file", conv.FileID, helpers.AccessLevelRead)
		if err != nil || !hasAccess {
			// Skip if user no longer has access
			continue
		}

		conversationsWithFile = append(conversationsWithFile, models.ConversationWithFile{
			ID:        conv.ID.Hex(),
			FileID:    conv.FileID,
			Messages:  conv.Messages,
			CreatedAt: conv.CreatedAt,
			UpdatedAt: conv.UpdatedAt,
			File: models.FileInfo{
				ID:          file.ID.Hex(),
				Filename:    file.Filename,
				ContentType: file.ContentType,
				Size:        file.Size,
			},
		})
	}

	return conversationsWithFile, nil
}
