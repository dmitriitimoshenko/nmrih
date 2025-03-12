import React, { useEffect, useState } from 'react';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#A28FD0', '#FF6666', '#66B3FF', '#FFCC99', '#66FF66', '#D0D0D0'];

const CountryPieChart = () => {
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
      });
  }, []);

  if (!data || data.length === 0) {
    return <p style={{ color: '#fff' }}>No data available for the pie chart.</p>;
  }

  return (
    <div className="pie-chart-container" style={{ margin: '0 auto', maxWidth: '400px', height: '400px' }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="percentage"
            nameKey="country"
            cx="50%"
            cy="50%"
            outerRadius={150}
            fill="#8884d8"
            label={false}
            labelLine={false}
          >
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{ backgroundColor: "#333", border: "none", color: "#fff" }}
            formatter={(value, name) => [`${Number(value).toFixed(2)}%`, name]}
          />
          <Legend wrapperStyle={{ color: "#fff" }} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
};

export default CountryPieChart;
