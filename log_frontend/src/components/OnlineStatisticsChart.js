import React from 'react';
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

const OnlineStatisticsChart = ({ data }) => {
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
        <Line type="monotone" dataKey="concurent_players_count" stroke="#8884d8" activeDot={{ r: 8 }} />
      </LineChart>
    </ResponsiveContainer>
  );
};

export default OnlineStatisticsChart;
