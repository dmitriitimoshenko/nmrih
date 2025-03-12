import React from 'react';
import Chart from './components/Chart';
import Controls from './components/Controls';
import useGraphData from './hooks/useGraphData';

function App() {
  const { chartData, loading, refreshData } = useGraphData();

  return (
    <div className="App">
      <h1>Top time-spent players</h1>
      <Controls onRefresh={refreshData} loading={loading} />
      <div className="graph-container">
        <Chart data={chartData} />
      </div>
    </div>
  );
}

export default App;
