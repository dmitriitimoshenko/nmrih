import React from 'react';
import TopTimeChart from './components/TopTimeChart';
import CountryPieChart from './components/CountryPieChart';
import PlayersInfo from './components/PlayersInfo';
import useTopTimeChartData from './hooks/useTopTimeChartData';
import useWindowDimensions from './hooks/useWindowDimensions';
import './App.css'; 

function App() {
  const { topTimeChartData, topTimeChartLoading } = useTopTimeChartData();
  const { width } = useWindowDimensions();
  const dashBoardUpperPart = (
    <div className="App">
      <h1>Krich Casual NMRiH Server Dashboard</h1>
      
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
            ) }
        </tbody>
      </table>
    </div>
  );

  return dashBoardUpperPart;
}

export default App;
