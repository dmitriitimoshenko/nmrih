import React, { useState } from 'react';
import './App.css';

function App() {
  const [loading, setLoading] = useState(false);
  const [graphTimestamp, setGraphTimestamp] = useState(Date.now());
  const [imgError, setImgError] = useState(false);

  const refreshData = async () => {
    setLoading(true);
    try {
      const response = await fetch(`https://log-parser.rulat-bot.duckdns.org/api/v1/parse?t=${graphTimestamp}`);
      if (!response.ok) {
        throw new Error('Error on API call');
      }
      setGraphTimestamp(Date.now());
    } catch (error) {
      console.error(error);
      alert('Error on data refresh');
    } finally {
      setLoading(false);
      window.location.reload();
    }
  };

  const handleImgError = () => {
    setImgError(true);
  };

  return (
    <div className="App">
      {/* Graph for Top-Time-Spent Players using the new endpoint */}
      <h1>Graph of Top-Time-Spent Players</h1>
      <div className="graph-container">
        {imgError ? (
          <div style={{ fontSize: '100px' }}>📉</div>
        ) : (
          <img
            src={`https://log-visualizer.rulat-bot.duckdns.org/graph/top-time-spent-players?t=${graphTimestamp}`}
            alt="Top-Time-Spent Players Graph"
            className="graph-image"
            onError={handleImgError}
          />
        )}
      </div>

      {/* Additional graph for Top Countries by sessions (>30 sec) */}
      <h1>Graph of Top Countries by Sessions (&gt;30 sec)</h1>
      <div className="graph-container">
        {imgError ? (
          <div style={{ fontSize: '100px' }}>📉</div>
        ) : (
          <img
            src={`https://log-visualizer.rulat-bot.duckdns.org/graph/top-counties-connected?t=${graphTimestamp}`}
            alt="Top Countries Graph"
            className="graph-image"
            onError={handleImgError}
          />
        )}
      </div>

      <div className="controls">
        <button onClick={refreshData} disabled={loading}>
          {loading ? 'Refreshing...' : 'Refresh Data'}
        </button>
        {loading && <div className="loader" />}
      </div>
    </div>
  );
}

export default App;
