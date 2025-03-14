import { useState, useEffect } from 'react';

const usePlayersInfo = () => {
  const [data, setData] = useState(null);

  useEffect(() => {
    fetch("https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=players-info")
      .then(response => response.json())
      .then(jsonData => {
        if (jsonData && jsonData.data) {
          setData(jsonData.data);
        }
      })
      .catch(err => {
        console.error("Error fetching player info:", err);
      })
  }, []);

  return { data };
};

export default usePlayersInfo;
