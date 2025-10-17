package main

import (
	"errors"
	"time"
)

const (
	Rows = 6
	Cols = 7
)

type Player int

const (
	Empty Player = 0
	P1    Player = 1
	P2    Player = 2
)

// GameLogic handles the in-memory state of a Connect Four game
type GameLogic struct {
	ID           string
	Board        [Rows][Cols]Player
	Turn         Player
	Player1      string
	Player2      string
	StartedAt    time.Time
	Moves        int
	Finished     bool
	WinnerUser   string // username or "draw"
	LastMoveTime time.Time
}

// NewGame initializes a new game
func NewGame(id, p1, p2 string) *GameLogic {
	return &GameLogic{
		ID:           id,
		Turn:         P1,
		Player1:      p1,
		Player2:      p2,
		StartedAt:    time.Now(),
		LastMoveTime: time.Now(),
	}
}

// Drop places a disc in the specified column
func (g *GameLogic) Drop(column int, username string) (row int, err error) {
	if g.Finished {
		return -1, errors.New("game finished")
	}
	if column < 0 || column >= Cols {
		return -1, errors.New("invalid column")
	}

	expected := g.CurrentPlayerName()
	if username != expected && !(username == "BOT" && expected == g.Player2) {
		return -1, errors.New("not your turn")
	}

	for r := Rows - 1; r >= 0; r-- {
		if g.Board[r][column] == Empty {
			var mark Player
			if username == g.Player1 {
				mark = P1
			} else {
				mark = P2
			}
			if username == "BOT" && g.Player2 == "BOT" {
				mark = P2
			}

			g.Board[r][column] = mark
			g.Moves++
			g.LastMoveTime = time.Now()

			if g.checkWin(r, column, mark) {
				g.Finished = true
				g.WinnerUser = username
			} else if g.isFull() {
				g.Finished = true
				g.WinnerUser = "draw"
			} else {
				g.toggleTurn()
			}
			return r, nil
		}
	}

	return -1, errors.New("column full")
}

// CurrentPlayerName returns the username of the player whose turn it is
func (g *GameLogic) CurrentPlayerName() string {
	if g.Turn == P1 {
		return g.Player1
	}
	return g.Player2
}

// toggleTurn switches the current player
func (g *GameLogic) toggleTurn() {
	if g.Turn == P1 {
		g.Turn = P2
	} else {
		g.Turn = P1
	}
}

// isFull checks if the board is full
func (g *GameLogic) isFull() bool {
	for c := 0; c < Cols; c++ {
		if g.Board[0][c] == Empty {
			return false
		}
	}
	return true
}

// checkWin checks if placing mark at (r, c) wins the game
func (g *GameLogic) checkWin(r, c int, mark Player) bool {
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, d := range dirs {
		count := 1
		for step := 1; step < 4; step++ {
			rr := r + d[0]*step
			cc := c + d[1]*step
			if rr < 0 || rr >= Rows || cc < 0 || cc >= Cols || g.Board[rr][cc] != mark {
				break
			}
			count++
		}
		for step := 1; step < 4; step++ {
			rr := r - d[0]*step
			cc := c - d[1]*step
			if rr < 0 || rr >= Rows || cc < 0 || cc >= Cols || g.Board[rr][cc] != mark {
				break
			}
			count++
		}
		if count >= 4 {
			return true
		}
	}

	return false
}

// BotChooseColumn chooses the next column for the bot
func (g *GameLogic) BotChooseColumn(botUsername string) int {
	var botMark, oppMark Player
	if botUsername == g.Player1 {
		botMark = P1
		oppMark = P2
	} else {
		botMark = P2
		oppMark = P1
	}

	canWin := func(col int, mark Player) bool {
		boardCopy := g.Board
		for r := Rows - 1; r >= 0; r-- {
			if boardCopy[r][col] == Empty {
				boardCopy[r][col] = mark
				temp := &GameLogic{Board: boardCopy}
				return temp.checkWin(r, col, mark)
			}
		}
		return false
	}

	// First try to win
	for c := 0; c < Cols; c++ {
		if canWin(c, botMark) {
			return c
		}
	}
	// Block opponent's win
	for c := 0; c < Cols; c++ {
		if canWin(c, oppMark) {
			return c
		}
	}
	// Preferred columns
	pref := []int{3, 2, 4, 1, 5, 0, 6}
	for _, c := range pref {
		if g.Board[0][c] == Empty {
			return c
		}
	}

	return 0
}
