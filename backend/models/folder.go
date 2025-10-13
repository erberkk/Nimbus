package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Folder struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     string             `json:"user_id" bson:"user_id"`
	FolderID   *string            `json:"folder_id" bson:"folder_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	Color      string             `json:"color" bson:"color,omitempty"`
	AccessList []AccessEntry      `json:"access_list" bson:"access_list"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

type FolderResponse struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Color      string        `json:"color"`
	ItemCount  int           `json:"item_count"`
	AccessList []AccessEntry `json:"access_list"`
	FolderID   *string       `json:"folder_id"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type CreateFolderRequest struct {
	Name     string `json:"name" validate:"required"`
	Color    string `json:"color"`
	FolderID string `json:"folder_id"`
}

type UpdateFolderRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
