import React, { useState, useRef, useEffect } from "react";
import GameBoard from "./GameBoard";
import Leaderboard from "./Leaderboard";
import GameResults from "./GameResults";

export default function App() {
  const [wsConnected, setWsConnected] = useState(false);
  const [username, setUsername] = useState("");
  const [gameId, setGameId] = useState(null);
  const [board, setBoard] = useState(null);
  const [status, setStatus] = useState("not_started");
  const [statusMessage, setStatusMessage] = useState("");
  const [winningCells, setWinningCells] = useState([]);
  const [leaderboardTrigger, setLeaderboardTrigger] = useState(0);
  const [showRecentGames, setShowRecentGames] = useState(false);
  const wsRef = useRef(null);

  useEffect(() => {
    return () => {
      if (wsRef.current) wsRef.current.close();
    };
  }, []);

  const connectWS = (username) => {
    if (!username) return alert("Enter username");

    const socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
      socket.send(JSON.stringify({ type: "join", username }));
      setWsConnected(true);
      setUsername(username);
    };

    socket.onmessage = (ev) => {
  const msg = JSON.parse(ev.data);

  switch (msg.type) {
    case "start": {
      setGameId(msg.gameId);
      setStatus("playing");
      setBoard(Array(6).fill(0).map(() => Array(7).fill(0)));
      setWinningCells([]);
      setStatusMessage("Game started! Make your move.");
      break;
    }

    case "move": {
      const boardUpdate = msg.payload?.board;
      const currentPlayer = msg.payload?.currentPlayer;

      if (boardUpdate) setBoard(boardUpdate);
      if (currentPlayer) setStatusMessage(`${currentPlayer}'s turn`);
      break;
    }

    case "end": {
      const winner = msg.payload.winner;
      const winningCells = msg.payload.winningCells || [];

      setWinningCells(winningCells);
      setStatus("ended");
      setStatusMessage(
        winner === "draw" ? "Game ended in a draw!" : `Winner: ${winner}`
      );
      setLeaderboardTrigger((prev) => prev + 1);
      break;
    }

    case "error": {
      const errorMsg = msg.payload;
      alert("Error: " + errorMsg);
      break;
    }

    default: {
      console.warn("Unknown message type:", msg.type);
    }
  }
};


    socket.onclose = () => setWsConnected(false);
    wsRef.current = socket;
  };

  const dropColumn = (col) => {
    if (wsRef.current && status === "playing") {
      wsRef.current.send(JSON.stringify({ type: "drop", column: col, username }));
    }
  };

  const goHome = () => {
    if (wsRef.current) wsRef.current.close();
    setUsername("");
    setGameId(null);
    setBoard(null);
    setWinningCells([]);
    setStatus("not_started");
    setStatusMessage("");
  };

  return (
    <div style={{ padding: 20, fontFamily: "Inter, system-ui, sans-serif" }}>
      <h1>4 in a Row</h1>

      {status !== "not_started" && <div className="status">{statusMessage}</div>}

      {(status === "not_started" || status === "ended") && (
        <>
          <Leaderboard refreshTrigger={leaderboardTrigger} />
          <div style={{ marginTop: 12 }}>
            <button
              onClick={() => setShowRecentGames((prev) => !prev)}
              style={{ marginBottom: 8 }}
            >
              {showRecentGames ? "Hide Recently Played" : "Show Recently Played"}
            </button>
            {showRecentGames && <GameResults />}
          </div>
        </>
      )}

      {status === "not_started" && (
        <div style={{ marginTop: 12 }}>
          <input
            placeholder="Enter username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <button onClick={() => connectWS(username)} style={{ marginLeft: 8 }}>
            Join
          </button>
          <p style={{ color: "#fff", marginTop: 8 }}>
            Open another tab with a different username to play vs a human, or wait 10s
            to play vs the bot.
          </p>
        </div>
      )}

      {status === "playing" && (
        <div style={{ marginTop: 12 }}>
          <div style={{ marginBottom: 8 }}>
            <strong>Game ID:</strong> {gameId}
            <span style={{ marginLeft: 16 }}>Username: {username}</span>
          </div>
          <GameBoard board={board} onDrop={dropColumn} winningCells={winningCells} />
        </div>
      )}

      {status === "ended" && (
        <div style={{ marginTop: 12 }}>
          <button
            onClick={() => window.location.reload()}
            style={{ marginRight: 8 }}
          >
            Play again
          </button>
          <button onClick={goHome} style={{ background: "#eee", padding: "6px 12px" }}>
            Home
          </button>
        </div>
      )}
    </div>
  );
}
