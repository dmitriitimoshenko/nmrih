import React, { useEffect, useState } from 'react';

const formatDuration = (durationSeconds) => {
  const totalSec = Math.floor(durationSeconds);
  if (totalSec < 60) return `${totalSec}s`;

  const hours = Math.floor(totalSec / 3600);
  const minutes = Math.floor((totalSec % 3600) / 60);
  const seconds = totalSec % 60;

  if (hours > 0) {
    return `${hours}h${minutes}m${seconds > 0 ? `${seconds}s` : ''}`;
  }
  return `${minutes}m${seconds > 0 ? `${seconds}s` : ''}`;
};

const PlayersInfo = () => {
  const [playersInfo, setPlayersInfo] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("https://api.rulat-bot.duckdns.org/api/v1/graph?type=players-info", {
      cache: 'no-cache'
    })
      .then(response => response.json())
      .then(jsonData => {
        setPlayersInfo(jsonData.data);
      })
      .catch(err => {
        console.error("Error fetching player info:", err);
      })
      .finally(() => setLoading(false));
  }, []);

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
            <span>Duration: {formatDuration(p.Duration)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default PlayersInfo;
