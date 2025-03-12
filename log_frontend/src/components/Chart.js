import React from 'react';
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

const Chart = ({ data }) => {
  return (
    <ResponsiveContainer width="100%" height={400}>
      <BarChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#444" />
        <XAxis dataKey="nick_name" stroke="#fff" />
        <YAxis
          stroke="#fff"
          label={{
            value: 'Time (hours)',
            angle: -90,
            position: 'insideLeft',
            fill: "#fff",
          }}
        />
        <Tooltip contentStyle={{ backgroundColor: "#333", border: "none", color: "#fff" }} />
        <Legend wrapperStyle={{ color: "#fff" }} />
        <Bar dataKey="time_spent" fill="#8884d8" />
      </BarChart>
    </ResponsiveContainer>
  );
};

export default Chart;
