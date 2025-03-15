const PlayersInfo = ({ data }) => {
  return (
    <div className="players-info">
      <h4>Players Connected ({data.data.count})</h4>
      <ul>
        {data.data.player && data.data.player.map((p, index) => (
          <li key={index}>
            <span>{p.Name}</span>
            <span>Score: {p.Score}</span>
            <span>Duration: {Number(p.Duration).toFixed(2)}s</span>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default PlayersInfo;
