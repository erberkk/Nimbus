package services

import (
	"context"
	"fmt"
	"time"

	"nimbus-backend/database"
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

