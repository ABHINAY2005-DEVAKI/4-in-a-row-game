package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(v)
}

func main() {
	db := InitDB()
	if db == nil {
		log.Fatal("❌ MongoDB initialization failed")
	}

	hub := NewHub(db)
	http.HandleFunc("/ws", hub.ServeWS)

	// ---------------- Leaderboard ----------------
	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		type LB struct {
			Username string `json:"username"`
			Wins     int    `json:"wins"`
		}
		var res []LB
		coll := db.Database.Collection("game_results")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"winner": bson.M{"$ne": "draw"}}}},
			{{Key: "$group", Value: bson.M{
				"_id":  "$winner",
				"wins": bson.M{"$sum": 1},
			}}},
			{{Key: "$sort", Value: bson.M{"wins": -1}}},
			{{Key: "$limit", Value: 50}},
		}

		cursor, err := coll.Aggregate(ctx, pipeline)
		if err != nil {
			log.Println("leaderboard query error:", err)
			writeJSON(w, res)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc struct {
				ID   string `bson:"_id"`
				Wins int    `bson:"wins"`
			}
			if err := cursor.Decode(&doc); err == nil {
				res = append(res, LB{Username: doc.ID, Wins: doc.Wins})
			}
		}
		writeJSON(w, res)
	})

	// ---------------- Efficiency Leaderboard ----------------
	http.HandleFunc("/efficiency", func(w http.ResponseWriter, r *http.Request) {
		type Eff struct {
			Username string  `json:"username"`
			Wins     int     `json:"wins"`
			AvgMoves float64 `json:"avg_moves"`
			MinMoves int64   `json:"min_moves"`
			MaxMoves int64   `json:"max_moves"`
		}
		var res []Eff
		coll := db.Database.Collection("game_results")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"winner": bson.M{"$ne": "draw"}, "moves": bson.M{"$ne": nil}}}},
			{{Key: "$group", Value: bson.M{
				"_id":       "$winner",
				"wins":      bson.M{"$sum": 1},
				"avg_moves": bson.M{"$avg": "$moves"},
				"min_moves": bson.M{"$min": "$moves"},
				"max_moves": bson.M{"$max": "$moves"},
			}}},
			{{Key: "$sort", Value: bson.M{"avg_moves": 1}}},
			{{Key: "$limit", Value: 50}},
		}

		cursor, err := coll.Aggregate(ctx, pipeline)
		if err != nil {
			log.Println("efficiency query error:", err)
			writeJSON(w, res)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc struct {
				ID       string  `bson:"_id"`
				Wins     int     `bson:"wins"`
				AvgMoves float64 `bson:"avg_moves"`
				MinMoves int64   `bson:"min_moves"`
				MaxMoves int64   `bson:"max_moves"`
			}
			if err := cursor.Decode(&doc); err == nil {
				res = append(res, Eff{
					Username: doc.ID,
					Wins:     doc.Wins,
					AvgMoves: doc.AvgMoves,
					MinMoves: doc.MinMoves,
					MaxMoves: doc.MaxMoves,
				})
			}
		}
		writeJSON(w, res)
	})

	// ---------------- Stats ----------------
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		type Stats struct {
			TotalPlayers int64 `json:"total_players"`
			TotalGames   int64 `json:"total_games"`
			TotalDraws   int64 `json:"total_draws"`
		}
		var s Stats

		coll := db.Database.Collection("game_results")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := coll.Find(ctx, bson.M{})
		if err == nil {
			playerSet := make(map[string]struct{})
			for cursor.Next(ctx) {
				var doc struct {
					Player1 string `bson:"player1"`
					Player2 string `bson:"player2"`
					Winner  string `bson:"winner"`
				}
				if err := cursor.Decode(&doc); err == nil {
					playerSet[doc.Player1] = struct{}{}
					playerSet[doc.Player2] = struct{}{}
					s.TotalGames++
					if strings.ToLower(doc.Winner) == "draw" {
						s.TotalDraws++
					}
				}
			}
			cursor.Close(ctx)
			s.TotalPlayers = int64(len(playerSet))
		}
		writeJSON(w, s)
	})

	// ---------------- Recent Game Results ----------------
	http.HandleFunc("/game_results", func(w http.ResponseWriter, r *http.Request) {
		type GR struct {
			GameID  string `json:"game_id"`
			Player1 string `json:"player1"`
			Player2 string `json:"player2"`
			Winner  string `json:"winner"`
			Moves   int64  `json:"moves"`
			Date    string `json:"date"`
		}
		var results []GR

		coll := db.Database.Collection("game_results")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		opts := options.Find()
		opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
		opts.SetLimit(50)

		cursor, err := coll.Find(ctx, bson.M{}, opts)
		if err != nil {
			log.Println("game_results query error:", err)
			writeJSON(w, results)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc struct {
				GameID   string    `bson:"game_id"`
				Player1  string    `bson:"player1"`
				Player2  string    `bson:"player2"`
				Winner   string    `bson:"winner"`
				Moves    int64     `bson:"moves"`
				CreatedAt time.Time `bson:"created_at"`
			}
			if err := cursor.Decode(&doc); err == nil {
				results = append(results, GR{
					GameID:  doc.GameID,
					Player1: doc.Player1,
					Player2: doc.Player2,
					Winner:  doc.Winner,
					Moves:   doc.Moves,
					Date:    doc.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}
		writeJSON(w, results)
	})

	fmt.Println("🚀 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
