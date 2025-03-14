import React from 'react';
import TopTimeChart from './components/TopTimeChart';
import CountryPieChart from './components/CountryPieChart';
import PlayersInfo from './components/PlayersInfo';
import Controls from './components/Controls';
import useGraphData from './hooks/useGraphData';

function App() {
  const { chartData, loading, refreshData } = useGraphData();

  return (
    <div className="App">
      <h1>Krich Casual NMRiH Server Dashboard</h1>
      <Controls onRefresh={refreshData} loading={loading} />
      
      <h3>Top Time-spent Players</h3>
      <div className="graph-container">
        <TopTimeChart data={chartData} />
      </div>

      <h3>Top Countries & Player Info</h3>
      <div className="chart-and-info-container">
        <div className="pie-chart-container">
          <CountryPieChart />
        </div>
        <div className="players-info">
          <PlayersInfo />
        </div>
      </div>
    </div>
  );
}

export default App;
