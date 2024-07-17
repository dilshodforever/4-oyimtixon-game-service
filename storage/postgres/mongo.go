package postgres

import (
	"context"
	"fmt"
	"log"

	u "github.com/dilshodforever/4-oyimtixon-game-service/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	Db    *mongo.Database
	Games u.GameStorage
}

func NewMongoConnecti0n() (u.InitRoot, error) {
	clientOptions := options.Client().ApplyURI("mongodb+srv://dilshod:2514@cluster0.klxv3df.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error: Couldn't connect to the database.", err)
	}

	fmt.Println("Connected to MongoDB!")

	db := client.Database("game")

	return &MongoStorage{Db: db}, err
}

func (s *MongoStorage) Game() u.GameStorage {
	if s.Games == nil {
		s.Games = &GameStorage{s.Db}
	}
	return s.Games
}
