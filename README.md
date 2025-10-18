# 🎮 4 in a Row Game

A real-time multiplayer web application of the classic **4 in a Row (Connect Four)** game, built with **React (frontend)** and **Go (backend)**, using **WebSockets** for real-time interaction and **MongoDB** as the database.

---

## 🚀 Features

- Real-time two-player gameplay using WebSockets  
- Automatic matchmaking (play against another player or bot)  
- Leaderboard with top players and recent match results  
- MongoDB integration for storing game data and statistics  
- Clean and responsive UI built with React

---

## 🛠️ Tech Stack

| Layer       | Technology Used |
|--------------|----------------|
| Frontend     | React + Vite |
| Backend      | Go (Golang) |
| Database     | MongoDB |
| Communication | WebSockets (real-time) |

---

## ⚙️ Setup Instructions

### 1. Clone the Repository
Terminal:
git clone https://github.com/ABHINAY2005-DEVAKI/4-in-a-row-game.git
cd 4-in-a-row-game


2. Backend Setup (Go)
Prerequisites

Go - installed

MongoDB - running locally or accessible remotely

Steps
Terminal:
cd backend
go mod tidy
go run main.go
This will start the backend server (default: http://localhost:8080).

3. Frontend Setup (React)
Prerequisites

Node.js - and npm installed

Steps
cd frontend
npm install
npm run dev


The frontend will start at http://localhost:5173 (default Vite port).

4. Connect Frontend and Backend

Ensure both servers are running:

Frontend: http://localhost:5173

Backend: http://localhost:8080

The frontend automatically connects to the backend WebSocket (ws://localhost:8080/ws) and REST endpoints for leaderboard and game results.

**Project Structure**

4-in-a-row-game/
│
├── backend/          # Go backend (WebSocket server + MongoDB)
│   ├── main.go
│   ├── db.go
│   ├── models/
│   ├── handlers/
│   └── ...
│
├── frontend/         # React frontend (Vite)
│   ├── src/
│   │   ├── App.jsx
│   │   ├── GameBoard.jsx
│   │   ├── Leaderboard.jsx
│   │   └── GameResults.jsx
│   ├── package.json
│   └── vite.config.js
│
└── README.md


🧩 How to Play

Run both frontend and backend locally.

Open the app in your browser.

Enter a username and click Join.

Open another tab (or invite another player).

Players take turns dropping pieces — first to connect four wins!


👨‍💻 Author

Abhinay Devaki
