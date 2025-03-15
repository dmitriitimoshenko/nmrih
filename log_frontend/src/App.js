import React from 'react';
import TopTimeChart from './components/TopTimeChart';
import CountryPieChart from './components/CountryPieChart';
import OnlineStatisticsChart from './components/OnlineStatisticsChart';
import PlayersInfo from './components/PlayersInfo';
import useTopTimeChartData from './hooks/useTopTimeChartData';
import useCountryPieChartData from './hooks/useCountryPieChartData';
import useOnlineStatisticsChartData from './hooks/useOnlineStatisticsChartData';
import usePlayersInfo from './hooks/usePlayersInfo';
import useWindowDimensions from './hooks/useWindowDimensions';
import './App.css'; 

function App() {
  const { topTimeChartData } = useTopTimeChartData();
  const { countryPieChartData } = useCountryPieChartData();
  const { playersInfoData } = usePlayersInfo();
  const { onlineStatisticsChartData } = useOnlineStatisticsChartData();
  const { width } = useWindowDimensions();

  // console.log(onlineStatisticsChartData);

  const dashBoardUpperPart = (
    <div className="App">
      <h1>Krich Casual NMRiH Server Dashboard</h1>
      <table>
        <tbody>
          <tr>
            <td colSpan="2">
              <h3>Top Time-spent Players</h3>
              <div className="graph-container">
                <TopTimeChart data={ topTimeChartData } />
              </div>
            </td>
          </tr>
          {width <= 800 ? (
            <div>
              <tr>
                <td>
                  <h3>Top Countries</h3>
                  <div className="pie-chart-container">
                    <CountryPieChart data={ countryPieChartData }/>
                  </div>
                </td>
              </tr>
              <tr>
                <td>
                  <h3>Player Info</h3>
                  <div className="players-info">
                    <PlayersInfo data={ playersInfoData }/>
                  </div>
                </td>
              </tr>
            </div>
          ) : (
            <tr>
              <td>
                <h3>Top Countries</h3>
                <div className="pie-chart-container">
                  <CountryPieChart data={ countryPieChartData }/>
                </div>
              </td>
              <td>
                <h3>Player Info</h3>
                <div className="players-info">
                  <PlayersInfo data={ playersInfoData }/>
                </div>
              </td>
            </tr>
          ) }
          <tr>
            <td colSpan="2">
            <h3>Online Statistics</h3>
              <div className="graph-container">
                <OnlineStatisticsChart data={ onlineStatisticsChartData } />
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );

  return dashBoardUpperPart;
}

export default App;
