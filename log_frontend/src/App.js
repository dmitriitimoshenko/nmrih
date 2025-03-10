import React, { useState } from 'react';
import './App.css';

function App() {
  const [loading, setLoading] = useState(false);
  const [graphTimestamp, setGraphTimestamp] = useState(Date.now());

  const refreshData = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://log-parser.rulat-bot.duckdns.org/api/v1/parse');
      if (!response.ok) {
        throw new Error('Error on API call');
      }
      setGraphTimestamp(Date.now());
    } catch (error) {
      console.error(error);
      alert('Error on data refresh');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="App">
      <h1>График данных</h1>
      <div className="graph-container">
        {/*adding a temp var*/}
        <img
          src={`https://log-visualizer.rulat-bot.duckdns.org/graph`}
          alt="Graph"
          className="graph-image"
        />
      </div>
      <div className="controls">
        <button onClick={refreshData} disabled={loading}>
          {loading ? 'Refreshing...' : 'Refrash data'}
        </button>
        {loading && <div className="loader" />}
      </div>
    </div>
  );
}

export default App;
