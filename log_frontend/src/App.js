import React, { useState, useEffect } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend
} from 'recharts';

const App = () => {
  const [chartData, setChartData] = useState([]);
  const [loading, setLoading] = useState(false);

  // Функция для получения данных графика
  const fetchGraphData = () => {
    fetch('https://log-parser.rulat-bot.duckdns.org/api/v1/graph?type=top-time-spent')
      .then(response => response.json())
      .then(data => {
        const convertedData = data.data.map(item => ({
          ...item,
          time_spent: item.time_spent / 1e9  // перевод из наносекунд в секунды
        }));
        setChartData(convertedData);
      })
      .catch(error => {
        console.error('Ошибка при получении данных графика:', error);
      });
  };

  useEffect(() => {
    fetchGraphData();
  }, []);

  // Функция для обновления данных: вызов parse endpoint с graphTimestamp и затем обновление графика
  const handleRefresh = () => {
    setLoading(true);
    const graphTimestamp = Date.now(); // текущий timestamp (в миллисекундах)
    fetch(`https://log-parser.rulat-bot.duckdns.org/api/v1/parse?t=${graphTimestamp}`)
      .then(response => {
        if (!response.ok) {
          throw new Error('Ошибка при вызове parse endpoint');
        }
        return response.json();
      })
      .then(() => {
        // После успешного вызова parse, обновляем данные графика
        fetchGraphData();
      })
      .catch(error => {
        console.error('Ошибка при обновлении данных:', error);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <div style={{ textAlign: 'center', marginTop: '20px' }}>
      <h1>Столбчатый график времени</h1>
      <button onClick={handleRefresh} disabled={loading}>
        {loading ? 'Обновление...' : 'Обновить'}
      </button>
      <BarChart
        width={600}
        height={300}
        data={chartData}
        margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
        style={{ margin: '0 auto', marginTop: '20px' }}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="nick_name" />
        <YAxis
          label={{
            value: 'Время (сек)',
            angle: -90,
            position: 'insideLeft'
          }}
        />
        <Tooltip />
        <Legend />
        <Bar dataKey="time_spent" fill="#8884d8" />
      </BarChart>
    </div>
  );
};

export default App;
