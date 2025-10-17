import React from "react";
import "./style.css";

const Rows = 6;
const Cols = 7;

export default function GameBoard({ board, onDrop, winningCells = [] }) {
  const renderCell = (r, c) => {
    const val = board ? board[r][c] : 0;
    const isWinningCell = winningCells.some(([wr, wc]) => wr === r && wc === c);

    let className = "cell";
    if (val === 1) className += " P1";
    if (val === 2) className += " P2";
    if (isWinningCell) className += " winning";

    return <div key={`${r}-${c}`} className={className}></div>;
  };

  return (
    <div>
      <div className="drop-buttons">
        {Array.from({ length: Cols }).map((_, c) => (
          <button key={c} className="drop-button" onClick={() => onDrop(c)}>
            ↓
          </button>
        ))}
      </div>

      <div className="board">
        {Array.from({ length: Rows }).map((_, r) =>
          Array.from({ length: Cols }).map((_, c) => renderCell(r, c))
        )}
      </div>
    </div>
  );
}
