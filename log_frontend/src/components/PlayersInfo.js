import React from 'react';

const PlayersInfo = ({data, loading}) => {
  if (loading) {
    return <p style={{ color: '#fff' }}>Loading player info...</p>;
  }

  if (!data || data.count === 0) {
    return <p style={{ color: '#fff' }}>No players connected.</p>;
  }

  return (
    <div className="players-info">
      <h4>Players Connected ({data.count})</h4>
      <ul>
        {data.player && data.player.map((p, index) => (
          <li key={index}>
            <span>{p.Name}</span>
            <span>Score: {p.Score}</span>
            <span>Duration: {Number(p.Duration).toFixed(2)}s</span>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default PlayersInfo;
