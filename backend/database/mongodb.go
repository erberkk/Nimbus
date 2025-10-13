package database

import (
	"context"
	"log"
	"nimbus-backend/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
var UserCollection *mongo.Collection
var FileCollection *mongo.Collection
var FolderCollection *mongo.Collection
var Client *mongo.Client

func Connect(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return err
	}

	// Ping kontrolü
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	// Global değişkenlere ata
	Client = client
	DB = client.Database(cfg.MongoDB)
	UserCollection = DB.Collection("users")
	FileCollection = DB.Collection("files")
	FolderCollection = DB.Collection("folders")

	log.Println("✅ MongoDB bağlantısı başarılı!")
	return nil
}

func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return Client.Disconnect(ctx)
	}
	return nil
}
