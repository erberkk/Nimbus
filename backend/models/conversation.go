package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Role      string    `json:"role" bson:"role"`                           // "user" or "assistant"
	Content   string    `json:"content" bson:"content"`                     // Message text
	Sources   []string  `json:"sources,omitempty" bson:"sources,omitempty"` // For assistant messages
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`                 // Message time
}

type Conversation struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FileID    string             `json:"file_id" bson:"file_id"`       // Associated file
	UserID    string             `json:"user_id" bson:"user_id"`       // Owner
	Messages  []Message          `json:"messages" bson:"messages"`     // Chat history
	CreatedAt time.Time          `json:"created_at" bson:"created_at"` // First message time
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"` // Last message time
}

type ConversationResponse struct {
	ID        string    `json:"id"`
	FileID    string    `json:"file_id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddMessageRequest struct {
	FileID  string   `json:"file_id" validate:"required"`
	Role    string   `json:"role" validate:"required,oneof=user assistant"`
	Content string   `json:"content" validate:"required"`
	Sources []string `json:"sources,omitempty"`
}
