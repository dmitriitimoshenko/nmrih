import React, { useState } from 'react';
import './App.css';

function App() {
  // Состояние для контроля загрузки и для форсирования обновления графика
  const [loading, setLoading] = useState(false);
  const [graphTimestamp, setGraphTimestamp] = useState(Date.now());

  // Функция для вызова эндпоинта парсера
  const refreshData = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://log-parser.rulat-bot.duckdns.org/api/v1/parse');
      if (!response.ok) {
        throw new Error('Ошибка при вызове API');
      }
      // По завершении обновляем временную метку, чтобы перезагрузить график
      setGraphTimestamp(Date.now());
    } catch (error) {
      console.error(error);
      alert('Ошибка при обновлении данных');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="App">
      <h1>График данных</h1>
      <div className="graph-container">
        {/* Добавляем временной параметр, чтобы браузер не кэшировал изображение */}
        <img
          src={`https://log-visualizer.rulat-bot.duckdns.org/graph`}
          alt="График"
          className="graph-image"
        />
      </div>
      <div className="controls">
        <button onClick={refreshData} disabled={loading}>
          {loading ? 'Обновление...' : 'Обновить данные'}
        </button>
        {loading && <div className="loader" />}
      </div>
    </div>
  );
}

export default App;
