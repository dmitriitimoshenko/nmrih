import React, { useState } from 'react';

const Controls = ({ onRefresh, loading }) => {
  const [copied, setCopied] = useState(false);

  const copyServerAddress = () => {
    navigator.clipboard.writeText("rulat-bot.duckdns.org:27015")
      .then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 3000);
      })
      .catch((error) => {
        console.error("Failed to copy server address:", error);
      });
  };

  return (
    <div className="controls">
      <button onClick={copyServerAddress} disabled={loading}>
        {copied ? "Copied!" : "Copy server address"}
      </button>
      <button onClick={onRefresh} disabled={loading}>
        {loading ? 'Refreshing...' : 'Refresh Data'}
      </button>
      {loading && <div className="loader"></div>}
    </div>
  );
};

export default Controls;
