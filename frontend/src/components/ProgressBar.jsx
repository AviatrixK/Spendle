import React from 'react';
import { motion } from 'framer-motion';

export default function ProgressBar({
  percent = 0,
  isOverspent = false,
  className = '',
  height = 'h-2',
}) {
  // Cap visual width at 100%
  const clampedPercent = Math.min(Math.max(percent, 0), 100);

  return (
    <div className={`w-full bg-white/[0.06] rounded-full overflow-hidden ${height} ${className}`}>
      <motion.div
        initial={{ width: 0 }}
        animate={{ width: `${clampedPercent}%` }}
        transition={{ duration: 0.3, ease: 'easeOut' }}
        className={`h-full rounded-full ${
          isOverspent ? 'bg-semantic-loss' : 'bg-accent'
        }`}
      />
    </div>
  );
}

