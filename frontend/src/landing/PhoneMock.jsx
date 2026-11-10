import React from 'react';

/**
 * Editorial CSS Phone Mockup
 * Fully responsive, sharp vector-rendered phone shell with subtle glass reflections.
 */
export default function PhoneMock({ children, className = '', notch = true }) {
  return (
    <div
      className={`relative mx-auto w-[290px] sm:w-[320px] rounded-[44px] p-3.5 bg-[#141419] border-[3px] border-white/10 shadow-2xl shadow-black/80 ${className}`}
      style={{
        boxShadow: '0 25px 60px -15px rgba(0, 0, 0, 0.9), inset 0 0 2px 1px rgba(255, 255, 255, 0.1)',
      }}
    >
      {/* Dynamic island / speaker pill */}
      {notch && (
        <div className="absolute top-5 left-1/2 -translate-x-1/2 w-20 h-4 bg-black/90 rounded-full z-30 flex items-center justify-end pr-2 border border-white/5">
          <div className="w-2 h-2 rounded-full bg-white/10" />
        </div>
      )}

      {/* Screen container */}
      <div className="relative rounded-[32px] overflow-hidden bg-[#0E0E11] min-h-[480px] sm:min-h-[530px] border border-white/5 flex flex-col justify-between p-4 pt-8">
        {children}
      </div>
    </div>
  );
}

