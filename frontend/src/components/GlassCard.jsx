import React from 'react';

export default function GlassCard({ children, className = '', ...props }) {
  return (
    <div
      className={`glass-panel rounded-2xl p-5 shadow-xl shadow-black/20 text-main ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}

