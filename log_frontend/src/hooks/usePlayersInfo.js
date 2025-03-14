import { useState, useEffect } from 'react';

const usePlayersInfo = () => {
  const [playersInfo, setPlayersInfo] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=players-info")
      .then(response => response.json())
      .then(jsonData => {
        setPlayersInfo(jsonData.data);
      })
      .catch(err => {
        console.error("Error fetching player info:", err);
      })
      .finally(() => setLoading(false));
  }, []);

  return { playersInfo, loading };
};

export default usePlayersInfo;
