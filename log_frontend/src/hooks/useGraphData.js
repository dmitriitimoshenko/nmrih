import { useState, useEffect } from 'react';

const useGraphData = () => {
  const [chartData, setChartData] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchGraphData = async () => {
    try {
      const response = await fetch('https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=top-time-spent');
      const data = await response.json();
      const convertedData = data.data.map(item => ({
        ...item,
        time_spent: Number(((item.time_spent / 1e9) / 3600).toFixed(1))
      }));
      setChartData(convertedData);
    } catch (error) {
      console.error('Error fetching graph data:', error);
    }
  };

  const refreshData = async () => {
    setLoading(true);
    try {
      const graphTimestamp = Date.now();
      const response = await fetch(`https://log-parser.rulat-bot.duckdns.org/api/v1/parse?t=${graphTimestamp}`);
      if (!response.ok) {
        throw new Error('Error calling /parse endpoint');
      }
      await response.json();
      await fetchGraphData();
    } catch (error) {
      console.error('Error refreshing data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGraphData();
  }, []);

  return { chartData, loading, refreshData };
};

export default useGraphData;
