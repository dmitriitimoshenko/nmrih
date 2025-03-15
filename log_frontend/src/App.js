import React, { useState } from 'react';
import TopTimeChart from './components/TopTimeChart';
import CountryPieChart from './components/CountryPieChart';
import PlayersInfo from './components/PlayersInfo';
import OnlineStatisticsChart from './components/OnlineStatisticsChart';
import useTopTimeChartData from './hooks/useTopTimeChartData';
import useWindowDimensions from './hooks/useWindowDimensions';
import Controls from './components/Controls';
import './App.css';

function App() {
  const { topTimeChartData } = useTopTimeChartData();
  const { width } = useWindowDimensions();
  const [loading, setLoading] = useState(false);

  const handleRefresh = () => {
    setLoading(true);
    fetch("https://api.rulat-bot.duckdns.org/api/v1/parse", {
      cache: 'no-cache'
    })
      .then(response => response.json())
      .then(data => {
        setLoading(false);
        window.location.reload();
      })
      .catch(err => {
        console.error("Error refreshing parse endpoint:", err);
        setLoading(false);
      });
  };

  return (
    <div className="App">
      <h1>Krich Casual NMRiH Server Dashboard</h1>
      <Controls onRefresh={handleRefresh} loading={loading} />
      <table>
        <tbody>
          <tr>
            <td colSpan="2">
              <h3>Top Time-spent Players</h3>
              <div className="graph-container">
                <TopTimeChart data={topTimeChartData} />
              </div>
            </td>
          </tr>
          {width <= 800 ? (
            <div>
              <tr>
                <td>
                  <h3>Top Countries</h3>
                  <div className="pie-chart-container">
                    <CountryPieChart />
                  </div>
                </td>
              </tr>
              <tr>
                <td>
                  <h3>Player Info</h3>
                  <div className="players-info">
                    <PlayersInfo />
                  </div>
                </td>
              </tr>
            </div>
          ) : (
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
          )}
          <tr>
            <td colSpan="2">
              <h3>Online Statistics</h3>
              <div className="graph-container">
                <OnlineStatisticsChart />
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export default App;
