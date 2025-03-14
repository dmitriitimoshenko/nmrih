import React from 'react';
import TopTimeChart from './components/TopTimeChart';
import CountryPieChart from './components/CountryPieChart';
import PlayersInfo from './components/PlayersInfo';
import Controls from './components/Controls';
import useGraphData from './hooks/useGraphData';
import useWindowDimensions from './hooks/useWindowDimensions';
import './App.css'; 

function App() {
  const { chartData, loading, refreshData } = useGraphData();

  const { windowDimensions } = useWindowDimensions();
  console.log(windowDimensions)

  return (
    <div className="App">

      <h1>Krich Casual NMRiH Server Dashboard</h1>
      <Controls onRefresh={refreshData} loading={loading} />
      
      <table>
        <tbody>
          <tr>
            <td colspan="2">
              <h3>Top Time-spent Players</h3>
              <div className="graph-container">
                <TopTimeChart data={chartData} />
              </div>
            </td>
          </tr>
          <tr>
            <td>
              <h3>Top Countries</h3>
              <div className="pie-chart-container">
                <CountryPieChart />
              </div>
            </td>
            <td>
              <h3>Player Info</h3>
              <div className="players-info">
                <PlayersInfo />
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export default App;
