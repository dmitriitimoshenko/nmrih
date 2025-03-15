import { useState, useEffect } from 'react';

const useOnlineStatisticsChartData = () => {
  const [data, setData] = useState([]);

  useEffect(() => {
    fetch("https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=online-statistics")
      .then(response => response.json())
      .then(jsonData => {
        if (jsonData && jsonData.data) {
          setData(jsonData.data);
        }
      })
      .catch(err => {
        console.error("Error fetching online statistics data:", err);
      })
  }, []);

  return { data };
};

export default useOnlineStatisticsChartData;
