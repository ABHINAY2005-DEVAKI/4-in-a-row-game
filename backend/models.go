package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameDB represents an ongoing or finished game record for MongoDB
type GameDB struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	GameID    string              `bson:"game_id"`
	Player1   string              `bson:"player1"`
	Player2   string              `bson:"player2"`
	StartedAt time.Time           `bson:"started_at"`
	Finished  bool                `bson:"finished"`
	Winner    string              `bson:"winner"`
	CreatedAt time.Time           `bson:"created_at"`
	UpdatedAt time.Time           `bson:"updated_at"`
}

// GameResult represents a finished game result
type GameResult struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	GameID    string              `bson:"game_id"`
	Player1   string              `bson:"player1"`
	Player2   string              `bson:"player2"`
	Winner    string              `bson:"winner"`
	Moves     int                 `bson:"moves"`
	Duration  time.Duration       `bson:"duration"`
	CreatedAt time.Time           `bson:"created_at"`
}
