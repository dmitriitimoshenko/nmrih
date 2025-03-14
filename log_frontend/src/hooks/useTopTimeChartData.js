import { useState, useEffect } from 'react';

const useTopTimeChartData = () => {
  const [topTimeChartData, setTopTimeChartData] = useState([]);
  const [topTimeChartLoading, setTopTimeChartLoading] = useState(false);

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

  const topTimeChartRefresh = async () => {
    setTopTimeChartLoading(true);
    try {
      const graphTimestamp = Date.now();
      const response = await fetch(`https://log-parser.rulat-bot.duckdns.org/api/v1/parse?t=${graphTimestamp}`);
      if (!response.ok) {
        throw new Error('Error calling /parse endpoint');
      }
      // Wait for parse endpoint to complete and then fetch updated graph data
      await response.json();
      await fetchGraphData();
    } catch (error) {
      console.error('Error refreshing data:', error);
    } finally {
      setTopTimeChartLoading(false);
    }
  };

  useEffect(() => {
    fetchGraphData();
  }, []);

  return { topTimeChartData, topTimeChartLoading, topTimeChartRefresh };
};

export default useTopTimeChartData;
