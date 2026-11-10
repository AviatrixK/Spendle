import React from 'react';

/**
 * Editorial film grain texture overlay using an inline SVG fractal noise filter.
 * Kept at extremely low opacity (0.025) and pointer-events-none so it doesn't interfere
 * with clicks or GPU paint performance.
 */
export default function FilmGrain() {
  return (
    <div
      aria-hidden="true"
      className="fixed inset-0 pointer-events-none z-50 opacity-[0.025] mix-blend-screen"
    >
      <svg className="w-full h-full" xmlns="http://www.w3.org/2000/svg">
        <filter id="noiseFilter">
          <feTurbulence
            type="fractalNoise"
            baseFrequency="0.8"
            numOctaves="3"
            stitchTiles="stitch"
          />
          <feColorMatrix
            type="matrix"
            values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 1 0"
          />
        </filter>
        <rect width="100%" height="100%" filter="url(#noiseFilter)" />
      </svg>
    </div>
  );
}

