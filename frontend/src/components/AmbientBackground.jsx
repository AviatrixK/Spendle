import React, { useEffect, useRef } from 'react';
import gsap from 'gsap';

/**
 * AmbientBackground
 * - Fixed full-viewport background strictly at -z-10 with pointer-events: none.
 * - Layer 1: Warm gold floor glow rising from the bottom viewport edge.
 * - Layer 2: High-performance canvas dot grid (26px spacing, DPR-aware, smooth bell falloff, idle RAF cutoff).
 * - Layer 3: 500px radial gold ambient spotlight following the cursor with GSAP quickTo and natural fade-out.
 * - Completely static on touch devices and when prefers-reduced-motion is active.
 *
 * @param {'full'|'soft'} intensity - 'full' for landing hero, 'soft' for inner app pages
 */
export default function AmbientBackground({ intensity = 'soft' }) {
  const canvasRef = useRef(null);
  const spotlightRef = useRef(null);

  useEffect(() => {
    const isFinePointer = window.matchMedia('(hover: hover) and (pointer: fine)').matches;
    const isReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    // --- 1. Canvas Dot Grid ---
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animId = null;
    let isRunning = false;
    let lastMoveTime = 0;

    let targetX = -2000;
    let targetY = -2000;
    let currentX = -2000;
    let currentY = -2000;

    let width = window.innerWidth;
    let height = window.innerHeight;
    let dpr = 1;

    const SPACING = 26;
    const INFLUENCE_RADIUS = 175;
    const BASE_ALPHA = intensity === 'full' ? 0.11 : 0.07;
    const PEAK_ALPHA = intensity === 'full' ? 0.55 : 0.42;

    function resizeCanvas() {
      if (!canvas) return;
      dpr = Math.min(window.devicePixelRatio || 1, 2);
      width = window.innerWidth;
      height = window.innerHeight;

      canvas.width = Math.round(width * dpr);
      canvas.height = Math.round(height * dpr);
      canvas.style.width = `${width}px`;
      canvas.style.height = `${height}px`;

      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      drawGrid(-2000, -2000);
    }

    function drawGrid(px, py) {
      ctx.clearRect(0, 0, width, height);

      const cols = Math.ceil(width / SPACING) + 1;
      const rows = Math.ceil(height / SPACING) + 1;

      for (let c = 0; c <= cols; c++) {
        for (let r = 0; r <= rows; r++) {
          const x = c * SPACING;
          const y = r * SPACING;

          let radius = 1;
          let rVal = 237;
          let gVal = 234;
          let bVal = 227;
          let alpha = BASE_ALPHA;

          const dx = x - px;
          const dy = y - py;
          const distSq = dx * dx + dy * dy;

          if (distSq < INFLUENCE_RADIUS * INFLUENCE_RADIUS) {
            const dist = Math.sqrt(distSq);
            const factor = 1 - dist / INFLUENCE_RADIUS;
            // Smooth cubic bell curve
            const smooth = factor * factor * (3 - 2 * factor);

            radius = 1 + smooth * 1.35; // Grow up to ~2.35px

            // Interpolate toward gold (#F5B83D: 245, 184, 61)
            rVal = Math.round(237 + (245 - 237) * smooth);
            gVal = Math.round(234 + (184 - 234) * smooth);
            bVal = Math.round(227 + (61 - 227) * smooth);
            alpha = BASE_ALPHA + smooth * (PEAK_ALPHA - BASE_ALPHA);
          }

          ctx.beginPath();
          ctx.arc(x, y, radius, 0, Math.PI * 2);
          ctx.fillStyle = `rgba(${rVal}, ${gVal}, ${bVal}, ${alpha})`;
          ctx.fill();
        }
      }
    }

    function renderLoop() {
      if (!isRunning) return;

      // Smooth lerp toward target
      const dx = targetX - currentX;
      const dy = targetY - currentY;
      currentX += dx * 0.14;
      currentY += dy * 0.14;

      drawGrid(currentX, currentY);

      // Check for idle cutoff: if pointer eased within threshold and has rested
      const distRemaining = Math.hypot(dx, dy);
      const timeSinceMove = performance.now() - lastMoveTime;

      if (distRemaining < 0.25 && timeSinceMove > 300) {
        isRunning = false;
        drawGrid(targetX, targetY);
        return;
      }

      animId = requestAnimationFrame(renderLoop);
    }

    function wakeUpRender() {
      lastMoveTime = performance.now();
      if (!isRunning && isFinePointer && !isReducedMotion) {
        isRunning = true;
        animId = requestAnimationFrame(renderLoop);
      }
    }

    // --- 2. GSAP Spotlight quickTo ---
    const spotlight = spotlightRef.current;
    let setSpotX = null;
    let setSpotY = null;

    if (spotlight && isFinePointer && !isReducedMotion) {
      setSpotX = gsap.quickTo(spotlight, 'x', { duration: 0.45, ease: 'power2.out' });
      setSpotY = gsap.quickTo(spotlight, 'y', { duration: 0.45, ease: 'power2.out' });
    }

    // --- 3. Pointer Listeners ---
    function handlePointerMove(e) {
      if (!isFinePointer || isReducedMotion) return;

      targetX = e.clientX;
      targetY = e.clientY;
      wakeUpRender();

      if (setSpotX && setSpotY) {
        setSpotX(e.clientX);
        setSpotY(e.clientY);
        gsap.to(spotlight, { opacity: 1, duration: 0.2, overwrite: 'auto' });
      }
    }

    function handlePointerLeave(e) {
      if (e.relatedTarget === null || e.clientY <= 0 || e.clientX <= 0 || e.clientX >= window.innerWidth || e.clientY >= window.innerHeight) {
        // Move target offscreen so dots return smoothly to resting state
        targetX = -2000;
        targetY = -2000;
        wakeUpRender();

        if (spotlight) {
          gsap.to(spotlight, { opacity: 0, duration: 0.4, overwrite: 'auto' });
        }
      }
    }

    function handlePointerEnter() {
      if (spotlight && isFinePointer && !isReducedMotion) {
        gsap.to(spotlight, { opacity: 1, duration: 0.25, overwrite: 'auto' });
      }
    }

    function handleVisibility() {
      if (document.hidden) {
        if (animId) cancelAnimationFrame(animId);
        isRunning = false;
      } else {
        drawGrid(targetX, targetY);
      }
    }

    // Initialize
    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);
    window.addEventListener('pointermove', handlePointerMove, { passive: true });
    document.addEventListener('mouseleave', handlePointerLeave);
    document.addEventListener('mouseenter', handlePointerEnter);
    document.addEventListener('visibilitychange', handleVisibility);

    return () => {
      if (animId) cancelAnimationFrame(animId);
      window.removeEventListener('resize', resizeCanvas);
      window.removeEventListener('pointermove', handlePointerMove);
      document.removeEventListener('mouseleave', handlePointerLeave);
      document.removeEventListener('mouseenter', handlePointerEnter);
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, [intensity]);

  const floorGlowAlpha = intensity === 'full' ? '0.14' : '0.08';
  const spotlightAlpha = intensity === 'full' ? '0.13' : '0.08';

  return (
    <div
      aria-hidden="true"
      className="fixed inset-0 pointer-events-none -z-10 overflow-hidden select-none"
    >
      {/* 1. Warm Gold Floor Glow (rising from bottom viewport edge) */}
      <div
        className="absolute inset-x-0 bottom-0 h-[55vh] pointer-events-none"
        style={{
          background: `radial-gradient(ellipse 95% 65% at 50% 100%, rgba(245, 184, 61, ${floorGlowAlpha}) 0%, rgba(245, 184, 61, 0.02) 55%, transparent 80%)`,
        }}
      />

      {/* 2. High-Performance Canvas Dot Grid */}
      <canvas
        ref={canvasRef}
        className="absolute inset-0 block w-full h-full pointer-events-none"
      />

      {/* 3. 500px Radial Gold Ambient Spotlight */}
      <div
        ref={spotlightRef}
        className="fixed top-0 left-0 w-[500px] h-[500px] -ml-[250px] -mt-[250px] rounded-full pointer-events-none opacity-0 will-change-transform"
        style={{
          background: `radial-gradient(circle, rgba(245, 184, 61, ${spotlightAlpha}) 0%, rgba(245, 184, 61, 0.04) 45%, transparent 70%)`,
        }}
      />
    </div>
  );
}
