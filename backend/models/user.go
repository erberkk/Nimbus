package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GoogleID  string             `json:"google_id" bson:"google_id"`
	Email     string             `json:"email" bson:"email"`
	Name      string             `json:"name" bson:"name"`
	Avatar    string             `json:"avatar" bson:"avatar"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// UserResponse kullanıcı bilgileri için response modeli
type UserResponse struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// JWT claims için struct
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	GoogleID string `json:"google_id"`
	jwt.RegisteredClaims
}

// TokenResponse JWT token response modeli
type TokenResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
