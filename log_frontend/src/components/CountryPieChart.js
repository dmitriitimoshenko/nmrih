import React from 'react';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import useCountryPieChartData from './hooks/useCountryPieChartData';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#A28FD0', '#FF6666', '#66B3FF', '#FFCC99', '#66FF66', '#D0D0D0'];

const CountryPieChart = () => {
  const { data, loading } = useCountryPieChartData();

  if (loading) {
    return <p style={{ color: '#fff' }}>Loading diagram data...</p>;
  }

  if (!data || data.length === 0) {
    return <p style={{ color: '#fff' }}>No data for the diagram.</p>;
  }

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
