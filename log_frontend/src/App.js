import React from 'react';
import TopTimeChart from './components/TopTimeChart';
import CountryPieChart from './components/CountryPieChart';
import PlayersInfo from './components/PlayersInfo';
import Controls from './components/Controls';
import useGraphData from './hooks/useGraphData';
import usePlayersInfo from './hooks/usePlayersInfo';
import useCountryPieChartData from './hooks/useCountryPieChartData';
import useWindowDimensions from './hooks/useWindowDimensions';
import './App.css'; 

function App() {
  const { topTimeChartData, loadingTopTimeChartData, refreshTopTimeChartData } = useGraphData();
  const { playerInfoData, loadingPlayerInfoData, refreshPlayerInfoData } = usePlayersInfo();
  const { countryPieChartData, loadingCountryPieChartData, refreshCountryPieChartData} = useCountryPieChartData();
  const { windowDimensions } = useWindowDimensions();

  console.log(windowDimensions)

  return (
    <div className="App">
      <h1>
        {windowDimensions.width} x {windowDimensions.height}
      </h1>

      <h1>Krich Casual NMRiH Server Dashboard</h1>
      <Controls onRefresh={refreshTopTimeChartData} loading={loadingTopTimeChartData} />
      <Controls onRefresh={refreshPlayerInfoData} loading={loadingPlayerInfoData} />
      <Controls onRefresh={refreshCountryPieChartData} loading={loadingCountryPieChartData} />

      <table>
        <tbody>
          <tr>
            <td colspan="2">
              <h3>Top Time-spent Players</h3>
              <div className="graph-container">
                <TopTimeChart data={topTimeChartData} />
              </div>
            </td>
          </tr>
          <tr>
            <td>
              <h3>Top Countries</h3>
              <div className="pie-chart-container">
                <CountryPieChart data={countryPieChartData} />
              </div>
            </td>
            <td>
              <h3>Player Info</h3>
              <div className="players-info">
                <PlayersInfo data={playerInfoData} />
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export default App;
