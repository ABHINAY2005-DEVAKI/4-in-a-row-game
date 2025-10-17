package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

type WSClient struct {
	Conn     *websocket.Conn
	Username string
	Send     chan []byte
	GameID   string
}

type Hub struct {
	mu      sync.Mutex
	waiting []*WSClient
	games   map[string]*GameInstance
	db      *MongoDB
}

type GameInstance struct {
	Game      *GameLogic
	P1        *WSClient
	P2        *WSClient
	CreatedAt time.Time
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewHub(db *MongoDB) *Hub {
	return &Hub{
		waiting: []*WSClient{},
		games:   make(map[string]*GameInstance),
		db:      db,
	}
}

type WSMessage struct {
	Type     string      `json:"type"`
	Username string      `json:"username,omitempty"`
	Column   int         `json:"column,omitempty"`
	GameID   string      `json:"gameId,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
}

// ServeWS handles new WebSocket connections
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open ws", http.StatusBadRequest)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var m WSMessage
	json.Unmarshal(msg, &m)
	if m.Type != "join" || m.Username == "" {
		conn.Close()
		return
	}

	client := &WSClient{Conn: conn, Username: m.Username, Send: make(chan []byte, 256)}
	go h.writer(client)
	go h.reader(client)
	h.addToQueue(client)
}

func (h *Hub) addToQueue(c *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.waiting) > 0 {
		other := h.waiting[0]
		h.waiting = h.waiting[1:]

		gameID := uuid.NewString()
		g := NewGame(gameID, other.Username, c.Username)
		inst := &GameInstance{Game: g, P1: other, P2: c, CreatedAt: time.Now()}
		other.GameID = gameID
		c.GameID = gameID
		h.games[gameID] = inst

		if h.db != nil {
			coll := h.db.Database.Collection("games")
			_, err := coll.InsertOne(context.TODO(), GameDB{
				GameID:    gameID,
				Player1:   other.Username,
				Player2:   c.Username,
				StartedAt: time.Now(),
				Finished:  false,
				CreatedAt: time.Now(),
			})
			if err != nil {
				log.Println("Mongo insert error:", err)
			}
		}

		startMsg := WSMessage{
			Type:   "start",
			GameID: gameID,
			Payload: map[string]interface{}{
				"player1": other.Username,
				"player2": c.Username,
			},
		}
		h.sendJSON(other, startMsg)
		h.sendJSON(c, startMsg)
		return
	}

	h.waiting = append(h.waiting, c)

	// Start bot game if no opponent joins in 10s
	go func(username string, client *WSClient) {
		time.Sleep(10 * time.Second)
		h.mu.Lock()
		defer h.mu.Unlock()

		for i, w := range h.waiting {
			if w.Username == username {
				h.waiting = append(h.waiting[:i], h.waiting[i+1:]...)
				gameID := uuid.NewString()
				botName := "BOT"
				g := NewGame(gameID, username, botName)
				inst := &GameInstance{Game: g, P1: client, P2: nil, CreatedAt: time.Now()}
				client.GameID = gameID
				h.games[gameID] = inst

				if h.db != nil {
					coll := h.db.Database.Collection("games")
					_, err := coll.InsertOne(context.TODO(), GameDB{
						GameID:    gameID,
						Player1:   username,
						Player2:   botName,
						StartedAt: time.Now(),
						Finished:  false,
						CreatedAt: time.Now(),
					})
					if err != nil {
						log.Println("Mongo insert error:", err)
					}
				}

				startMsg := WSMessage{
					Type:   "start",
					GameID: gameID,
					Payload: map[string]interface{}{
						"player1": username,
						"player2": botName,
					},
				}
				h.sendJSON(client, startMsg)
				go h.botLoop(inst)
				break
			}
		}
	}(c.Username, c)
}

func (h *Hub) sendJSON(client *WSClient, m WSMessage) {
	b, _ := json.Marshal(m)
	select {
	case client.Send <- b:
	default:
	}
}

func (h *Hub) writer(client *WSClient) {
	for msg := range client.Send {
		client.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (h *Hub) reader(client *WSClient) {
	defer client.Conn.Close()
	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			h.handleDisconnect(client)
			break
		}
		var m WSMessage
		json.Unmarshal(data, &m)
		if m.Type == "drop" {
			h.handleDrop(client, m.Column)
		}
	}
}

func (h *Hub) handleDrop(client *WSClient, col int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	inst, ok := h.games[client.GameID]
	if !ok {
		return
	}

	_, err := inst.Game.Drop(col, client.Username)
	if err != nil {
		h.sendJSON(client, WSMessage{Type: "error", Payload: err.Error()})
		return
	}

	moveMsg := WSMessage{Type: "move", GameID: inst.Game.ID, Payload: map[string]interface{}{
		"player": client.Username,
		"column": col,
		"board":  inst.Game.Board,
	}}
	if inst.P1 != nil {
		h.sendJSON(inst.P1, moveMsg)
	}
	if inst.P2 != nil {
		h.sendJSON(inst.P2, moveMsg)
	}

	if inst.Game.Finished {
		h.finishGame(inst)
	} else if inst.P2 == nil {
		go h.botLoop(inst)
	}
}

func (h *Hub) finishGame(inst *GameInstance) {
	resMsg := WSMessage{Type: "end", GameID: inst.Game.ID, Payload: map[string]interface{}{
		"winner": inst.Game.WinnerUser,
	}}
	if inst.P1 != nil {
		h.sendJSON(inst.P1, resMsg)
	}
	if inst.P2 != nil {
		h.sendJSON(inst.P2, resMsg)
	}

	if h.db != nil {
		resColl := h.db.Database.Collection("game_results")
		gameColl := h.db.Database.Collection("games")

		_, err := resColl.InsertOne(context.TODO(), GameResult{
			GameID:    inst.Game.ID,
			Player1:   inst.Game.Player1,
			Player2:   inst.Game.Player2,
			Winner:    inst.Game.WinnerUser,
			Moves:     inst.Game.Moves,
			Duration:  time.Since(inst.Game.StartedAt),
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Println("Error inserting game result:", err)
		}

		_, err = gameColl.UpdateOne(
			context.TODO(),
			bson.M{"game_id": inst.Game.ID},
			bson.M{"$set": bson.M{
				"finished":   true,
				"winner":     inst.Game.WinnerUser,
				"updated_at": time.Now(),
			}},
		)
		if err != nil {
			log.Println("Error updating game:", err)
		}
	}

	delete(h.games, inst.Game.ID)
}

func (h *Hub) botLoop(inst *GameInstance) {
	time.Sleep(350 * time.Millisecond)
	if inst.Game.Finished {
		return
	}
	botName := "BOT"
	col := inst.Game.BotChooseColumn(botName)
	inst.Game.Drop(col, botName)

	moveMsg := WSMessage{Type: "move", GameID: inst.Game.ID, Payload: map[string]interface{}{
		"player": botName,
		"column": col,
		"board":  inst.Game.Board,
	}}
	if inst.P1 != nil {
		h.sendJSON(inst.P1, moveMsg)
	}

	if inst.Game.Finished {
		h.finishGame(inst)
	}
}

func (h *Hub) handleDisconnect(c *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, w := range h.waiting {
		if w == c {
			h.waiting = append(h.waiting[:i], h.waiting[i+1:]...)
			break
		}
	}

	gameID := c.GameID
	if gameID == "" {
		return
	}
	inst, ok := h.games[gameID]
	if !ok || inst.Game.Finished {
		return
	}

	var other string
	if inst.P1 != nil && inst.P1.Username == c.Username {
		if inst.P2 != nil {
			other = inst.P2.Username
		} else {
			other = "BOT"
		}
	} else {
		other = inst.P1.Username
	}

	inst.Game.Finished = true
	inst.Game.WinnerUser = other

	endMsg := WSMessage{Type: "end", GameID: gameID, Payload: map[string]interface{}{
		"winner":  other,
		"forfeit": true,
	}}
	if inst.P1 != nil {
		h.sendJSON(inst.P1, endMsg)
	}
	if inst.P2 != nil {
		h.sendJSON(inst.P2, endMsg)
	}

	if h.db != nil {
		resColl := h.db.Database.Collection("game_results")
		gameColl := h.db.Database.Collection("games")

		_, err := resColl.InsertOne(context.TODO(), GameResult{
			GameID:    inst.Game.ID,
			Player1:   inst.Game.Player1,
			Player2:   inst.Game.Player2,
			Winner:    inst.Game.WinnerUser,
			Moves:     inst.Game.Moves,
			Duration:  time.Since(inst.Game.StartedAt),
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Println("Error inserting forfeit result:", err)
		}

		_, err = gameColl.UpdateOne(
			context.TODO(),
			bson.M{"game_id": inst.Game.ID},
			bson.M{"$set": bson.M{
				"finished":   true,
				"winner":     inst.Game.WinnerUser,
				"updated_at": time.Now(),
			}},
		)
		if err != nil {
			log.Println("Error updating forfeit game:", err)
		}
	}

	delete(h.games, gameID)
}
