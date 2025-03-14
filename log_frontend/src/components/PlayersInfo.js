import React from 'react';
import usePlayersInfo from './hooks/usePlayersInfo';

const PlayersInfo = () => {
  const { playersInfo, loading } = usePlayersInfo();

  if (loading) {
    return <p style={{ color: '#fff' }}>Loading player info...</p>;
  }

  if (!playersInfo || playersInfo.count === 0) {
    return <p style={{ color: '#fff' }}>No players connected.</p>;
  }

  return (
    <div className="players-info">
      <h4>Players Connected ({playersInfo.count})</h4>
      <ul>
        {playersInfo.player && playersInfo.player.map((p, index) => (
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
