import React, { useRef, useCallback, useEffect } from 'react';

/**
 * GlassCard
 * - Premium glassmorphic container with 28px border-radius and inner highlight.
 * - Pointer spotlight tracking: calculates mouse coordinates relative to the card's bounding box.
 * - Subtly illuminates the border and card surface with a soft radial gold glow on hover.
 * - Supports variants: 'default', 'elevated', and 'interactive' (smooth hover lift).
 * - Safe for touch devices and prefers-reduced-motion.
 *
 * @param {'default'|'elevated'|'interactive'} variant - Visual elevation and interaction mode
 * @param {boolean} spotlight - Enable relative pointer spotlight (default: true)
 * @param {string} className - Additional CSS classes
 */
export default function GlassCard({
  children,
  variant = 'default',
  spotlight = true,
  className = '',
  style = {},
  ...props
}) {
  const cardRef = useRef(null);
  const rafId = useRef(null);

  const handlePointerMove = useCallback(
    (e) => {
      if (!spotlight || !cardRef.current) return;
      if (!window.matchMedia('(hover: hover) and (pointer: fine)').matches) return;

      if (rafId.current) {
        cancelAnimationFrame(rafId.current);
      }

      const clientX = e.clientX;
      const clientY = e.clientY;

      rafId.current = requestAnimationFrame(() => {
        if (!cardRef.current) return;
        const rect = cardRef.current.getBoundingClientRect();
        const x = clientX - rect.left;
        const y = clientY - rect.top;

        cardRef.current.style.setProperty('--mx', `${x}px`);
        cardRef.current.style.setProperty('--my', `${y}px`);
        cardRef.current.style.setProperty('--spotlight-opacity', '1');
      });
    },
    [spotlight]
  );

  const handlePointerLeave = useCallback(() => {
    if (!spotlight || !cardRef.current) return;
    if (rafId.current) {
      cancelAnimationFrame(rafId.current);
    }
    cardRef.current.style.setProperty('--spotlight-opacity', '0');
  }, [spotlight]);

  useEffect(() => {
    return () => {
      if (rafId.current) {
        cancelAnimationFrame(rafId.current);
      }
    };
  }, []);

  let baseGlassClass = 'glass';
  if (variant === 'elevated') {
    baseGlassClass = 'glass-elevated';
  } else if (variant === 'interactive') {
    baseGlassClass = 'glass-interactive cursor-pointer';
  }

  return (
    <div
      ref={cardRef}
      onPointerMove={handlePointerMove}
      onPointerLeave={handlePointerLeave}
      className={`relative overflow-hidden ${baseGlassClass} p-5 text-main shadow-glass transition-all ${className}`}
      style={style}
      {...props}
    >
      {spotlight && (
        <>
          {/* 1. Subtle Radial Border Glow */}
          <div className="card-spotlight-layer" aria-hidden="true" />
          {/* 2. Soft Gold Radial Surface Illumination */}
          <div className="card-spotlight-surface" aria-hidden="true" />
        </>
      )}
      <div className="relative z-10">{children}</div>
    </div>
  );
}
