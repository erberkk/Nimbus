package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Folder struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	UserID     string               `json:"user_id" bson:"user_id"`
	FolderID   *string              `json:"folder_id" bson:"folder_id,omitempty"`
	ParentID   *primitive.ObjectID  `json:"parent_id" bson:"parent_id,omitempty"`
	Ancestors  []primitive.ObjectID `json:"ancestors" bson:"ancestors"`
	Name       string               `json:"name" bson:"name"`
	Color      string               `json:"color" bson:"color,omitempty"`
	PublicLink string               `json:"public_link" bson:"public_link"`
	AccessList []AccessEntry        `json:"access_list" bson:"access_list"`
	IsStarred  bool                 `json:"is_starred" bson:"is_starred"`
	DeletedAt  *time.Time           `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	CreatedAt  time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at" bson:"updated_at"`
}

type FolderResponse struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Color      string               `json:"color"`
	PublicLink string               `json:"public_link"`
	ItemCount  int                  `json:"item_count"`
	Size       int64                `json:"size"` // Total size of all files in folder (recursive)
	AccessList []AccessEntry        `json:"access_list"`
	ParentID   *primitive.ObjectID  `json:"parent_id,omitempty"`
	Ancestors  []primitive.ObjectID `json:"ancestors"`
	FolderID   *string              `json:"folder_id"`
	IsStarred  bool                 `json:"is_starred"`
	DeletedAt  *time.Time           `json:"deleted_at,omitempty"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
	Owner      *UserResponse         `json:"owner,omitempty"` // Owner information
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
