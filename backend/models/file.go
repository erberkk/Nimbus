package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AccessEntry represents a user's access to a file or folder
type AccessEntry struct {
	UserID        string              `json:"user_id" bson:"user_id"`
	AccessType    string              `json:"access_type" bson:"access_type"` // "read" or "write"
	GrantedAt     time.Time           `json:"granted_at" bson:"granted_at"`
	GrantedBy     string              `json:"granted_by" bson:"granted_by"`
	InheritedFrom *primitive.ObjectID `json:"inherited_from,omitempty" bson:"inherited_from,omitempty"`
}

type File struct {
	ID               primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	UserID           string               `json:"user_id" bson:"user_id"`
	FolderID         *string              `json:"folder_id" bson:"folder_id,omitempty"`
	ParentID         *primitive.ObjectID  `json:"parent_id" bson:"parent_id,omitempty"`
	Ancestors        []primitive.ObjectID `json:"ancestors" bson:"ancestors"`
	Filename         string               `json:"filename" bson:"filename"`
	Size             int64                `json:"size" bson:"size"`
	ContentType      string               `json:"content_type" bson:"content_type"`
	MinioPath        string               `json:"minio_path" bson:"minio_path"`
	PublicLink       string               `json:"public_link" bson:"public_link"`
	AccessList       []AccessEntry        `json:"access_list" bson:"access_list"`
	ProcessingStatus string               `json:"processing_status" bson:"processing_status"` // none, pending, processing, completed, failed
	ProcessingError  string               `json:"processing_error,omitempty" bson:"processing_error,omitempty"`
	ProcessedAt      *time.Time           `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	ChunkCount       int                  `json:"chunk_count" bson:"chunk_count"`
	CreatedAt        time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at" bson:"updated_at"`
}

type FileResponse struct {
	ID               string               `json:"id"`
	UserID           string               `json:"user_id"`           // Owner of the file
	Filename         string               `json:"filename"`
	Size             int64                `json:"size"`
	ContentType      string               `json:"content_type"`
	MinioPath        string               `json:"minio_path"`        // Path in MinIO storage
	PublicLink       string               `json:"public_link"`
	AccessList       []AccessEntry        `json:"access_list"`
	ParentID         *primitive.ObjectID  `json:"parent_id,omitempty"`
	Ancestors        []primitive.ObjectID `json:"ancestors"`
	ProcessingStatus string               `json:"processing_status"`
	ProcessingError  string               `json:"processing_error,omitempty"`
	ProcessedAt      *time.Time           `json:"processed_at,omitempty"`
	ChunkCount       int                  `json:"chunk_count"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type CreateFileRequest struct {
	Filename    string  `json:"filename" validate:"required"`
	Size        int64   `json:"size" validate:"required"`
	ContentType string  `json:"content_type" validate:"required"`
	MinioPath   string  `json:"minio_path" validate:"required"`
	FolderID    *string `json:"folder_id"`
}
