import React from 'react';

export default function SpendleLoader({ label = 'Getting things ready' }) {
  return <div className="spendle-loader" role="status" aria-live="polite">
    <span className="loader-mark" aria-hidden="true"><span className="loader-coin">₹</span><i /><i /><i /></span>
    <span className="loader-wordmark">Spendle<span>.</span></span>
    <span className="loader-caption">{label}<span className="loader-dots" aria-hidden="true">...</span></span>
  </div>;
}
