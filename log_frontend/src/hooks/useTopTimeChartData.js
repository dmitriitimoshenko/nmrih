import { useState, useEffect } from 'react';

const useTopTimeChartData = () => {
  const [topTimeChartData, setTopTimeChartData] = useState([]);

  const fetchGraphData = async () => {
    try {
      const response = await fetch('https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=top-time-spent');
      const data = await response.json();
      const convertedData = data.data.map(item => ({
        ...item,
        // Convert nanoseconds to hours with one decimal place
        time_spent: Number(((item.time_spent / 1e9) / 3600).toFixed(1))
      }));
      setTopTimeChartData(convertedData);
    } catch (error) {
      console.error('Error fetching graph data:', error);
    }
  };

  useEffect(() => {
    fetchGraphData();
  }, []);

  return { topTimeChartData, topTimeChartLoading };
};

export default useTopTimeChartData;
