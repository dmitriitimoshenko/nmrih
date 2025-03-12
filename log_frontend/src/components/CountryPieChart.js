import React, { useState, useEffect } from 'react';
import { PieChart, Pie, Tooltip, Legend, ResponsiveContainer, Cell } from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#A28FD0', '#FF6666', '#66B3FF', '#FFCC99', '#66FF66', '#D0D0D0'];

const CountryPieChart = () => {
  const [data, setData] = useState([]);

  useEffect(() => {
    fetch("https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=top-country")
      .then(response => response.json())
      .then(json => {
        if (json && json.data) {
          setData(json.data);
        }
      })
      .catch(error => {
        console.error("Error fetching pie chart data", error);
      });
  }, []);

  return (
    <div className="pie-chart-container">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="percentage"
            nameKey="country"
            cx="50%"
            cy="50%"
            outerRadius={180}  // Adjusts the diameter; 2*180 = 360px, within 400px height
            fill="#8884d8"
            label
          >
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip contentStyle={{ backgroundColor: "#333", border: "none", color: "#fff" }} />
          <Legend wrapperStyle={{ color: "#fff" }} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
};

export default CountryPieChart;
