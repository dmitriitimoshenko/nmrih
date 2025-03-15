import React, { useState, useEffect } from 'react';
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend
} from 'recharts';

const OnlineStatisticsChart = () => {
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
        console.error("Error fetching online statistics chart data:", err);
      });
  }, []);

  if (!data || data.length === 0) {
    return <p style={{ color: '#fff' }}>No data available for the online statistics chart.</p>;
  }

  console.log(data);

  return (
    <ResponsiveContainer width="100%" height={400}>
      <LineChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#444" />
        <XAxis dataKey="hour" stroke="#fff" />
        <YAxis
          stroke="#fff"
          label={{
            value: 'Concurrent Players',
            angle: -90,
            position: 'insideLeft',
            fill: "#fff",
          }}
        />
        <Tooltip contentStyle={{ backgroundColor: "#333", border: "none", color: "#fff" }} />
        <Legend wrapperStyle={{ color: "#fff" }} />
        <Line type="monotone" dataKey="concurrent_players_count" stroke="#8884d8" activeDot={{ r: 8 }} />
      </LineChart>
    </ResponsiveContainer>
  );
};

export default OnlineStatisticsChart;
