import React, { useEffect, useState } from "react";

export default function GameResults() {
  const [results, setResults] = useState([]);

  useEffect(() => {
    const fetchResults = async () => {
      try {
        const res = await fetch("http://localhost:8080/game_results");
        if (!res.ok) throw new Error("Failed to fetch results");
        const data = await res.json();
        setResults(data);
      } catch (err) {
        console.error("Game results fetch error:", err);
      }
    };
    fetchResults();
  }, []);

  return (
    <div style={{ marginTop: 24 }}>
      <h3>Recent Game Results</h3>
      {results.length > 0 ? (
        <table>
          <thead>
            <tr>
              <th>Game ID</th>
              <th>Player 1</th>
              <th>Player 2</th>
              <th>Winner</th>
              <th>Moves</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            {results.map((r, idx) => (
              <tr key={idx}>
                <td>{r.game_id}</td>
                <td>{r.player1}</td>
                <td>{r.player2}</td>
                <td>{r.winner}</td>
                <td>{r.moves}</td>
                <td>{new Date(r.date).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p>No game results yet.</p>
      )}
    </div>
  );
}
