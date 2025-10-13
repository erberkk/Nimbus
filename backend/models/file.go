package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AccessEntry represents a user's access to a file or folder
type AccessEntry struct {
	UserID     string    `json:"user_id" bson:"user_id"`
	AccessType string    `json:"access_type" bson:"access_type"` // "read" or "write"
	GrantedAt  time.Time `json:"granted_at" bson:"granted_at"`
	GrantedBy  string    `json:"granted_by" bson:"granted_by"`
}

type File struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      string             `json:"user_id" bson:"user_id"`
	FolderID    *string            `json:"folder_id" bson:"folder_id,omitempty"`
	Filename    string             `json:"filename" bson:"filename"`
	Size        int64              `json:"size" bson:"size"`
	ContentType string             `json:"content_type" bson:"content_type"`
	MinioPath   string             `json:"minio_path" bson:"minio_path"`
	AccessList  []AccessEntry      `json:"access_list" bson:"access_list"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type FileResponse struct {
	ID          string        `json:"id"`
	Filename    string        `json:"filename"`
	Size        int64         `json:"size"`
	ContentType string        `json:"content_type"`
	AccessList  []AccessEntry `json:"access_list"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CreateFileRequest struct {
	Filename    string  `json:"filename" validate:"required"`
	Size        int64   `json:"size" validate:"required"`
	ContentType string  `json:"content_type" validate:"required"`
	MinioPath   string  `json:"minio_path" validate:"required"`
	FolderID    *string `json:"folder_id"`
}
