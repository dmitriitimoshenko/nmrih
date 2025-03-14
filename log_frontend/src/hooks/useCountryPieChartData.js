import { useState, useEffect } from 'react';

const useCountryPieChartData = () => {
  const [data, setData] = useState([]);

  useEffect(() => {
    fetch("https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=top-country")
      .then(response => response.json())
      .then(jsonData => {
        if (jsonData && jsonData.data) {
          setData(jsonData.data);
        }
      })
      .catch(err => {
        console.error("Error fetching pie chart data:", err);
      })
  }, []);

  return { data };
};

export default useCountryPieChartData;
