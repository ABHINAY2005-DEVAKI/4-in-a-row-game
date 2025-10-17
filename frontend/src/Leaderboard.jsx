import React, { useEffect, useState } from "react";

export default function Leaderboard({ refreshTrigger = 0 }) {
  const [list, setList] = useState([]);
  const [showFull, setShowFull] = useState(false);

  useEffect(() => {
    fetch("http://localhost:8080/leaderboard")
      .then((res) => res.json())
      .then((data) => setList(Array.isArray(data) ? data : [])) // ✅ ensure it's an array
      .catch((err) => console.error("Leaderboard fetch error:", err));
  }, [refreshTrigger]);

  const safeList = Array.isArray(list) ? list : [];
  const displayedList = showFull ? safeList : safeList.slice(0, 3);

  return (
    <div style={{ marginTop: 24 }}>
      <h3>Leaderboard</h3>
      {safeList.length > 0 ? (
        <>
          <ol>
            {displayedList.map((l, idx) => (
              <li key={idx}>
                {l.username} — {l.wins} wins
              </li>
            ))}
          </ol>
          {safeList.length > 3 && (
            <button onClick={() => setShowFull((prev) => !prev)}>
              {showFull ? "Show Top 3" : "Show Full Leaderboard"}
            </button>
          )}
        </>
      ) : (
        <p>No results yet</p>
      )}
    </div>
  );
}
