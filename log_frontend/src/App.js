import React, { useState, useEffect } from 'react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend
} from 'recharts';
import './App.css';

const App = () => {
  const [chartData, setChartData] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchGraphData = () => {
    fetch('https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=top-time-spent')
      .then(response => response.json())
      .then(data => {
        const convertedData = data.data.map(item => ({
          ...item,
          time_spent: Number(((item.time_spent / 1e9) / 3600).toFixed(1))
        }));
        setChartData(convertedData);
      })
      .catch(error => {
        console.error('Error fetching graph data:', error);
      });
  };

  useEffect(() => {
    fetchGraphData();
  }, []);

  const handleRefresh = () => {
    setLoading(true);
    const graphTimestamp = Date.now();
    fetch(`https://log-parser.rulat-bot.duckdns.org/api/v1/parse?t=${graphTimestamp}`)
      .then(response => {
        if (!response.ok) {
          throw new Error('Error calling /parse endpoint');
        }
        return response.json();
      })
      .then(() => {
        fetchGraphData();
      })
      .catch(error => {
        console.error('Error while data refresh:', error);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <div className="App">
      <h1>Top time-spent players</h1>
      <div className="controls">
        <button onClick={handleRefresh} disabled={loading}>
          {loading ? 'Refreshing...' : 'Refresh Data'}
        </button>
        {loading && <div className="loader"></div>}
      </div>
      <div className="graph-container">
        <ResponsiveContainer width="100%" height={400}>
          <BarChart
            data={chartData}
            margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="nick_name" />
            <YAxis
              label={{
                value: 'Time (hours)',
                angle: -90,
                position: 'insideLeft'
              }}
            />
            <Tooltip />
            <Legend />
            <Bar dataKey="time_spent" fill="#8884d8" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

export default App;
