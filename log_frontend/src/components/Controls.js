import React from 'react';

const Controls = ({ onRefresh, loading }) => {
  return (
    <div className="controls">
      <button onClick={onRefresh} disabled={loading}>
        {loading ? 'Refreshing...' : 'Refresh Data'}
      </button>
      {loading && <div className="loader"></div>}
    </div>
  );
};

export default Controls;
