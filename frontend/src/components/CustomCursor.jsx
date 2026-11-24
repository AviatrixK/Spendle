import React, { useEffect, useRef } from 'react';
import gsap from 'gsap';

/**
 * CustomCursor
 * - Refined dual-layer cursor: a sharp gold central dot and a smooth lagging outer ring.
 * - Always sits at z-[9999] with pointer-events: none (never blocks clicks).
 * - Native cursor remains completely visible and untouched.
 * - Central dot tracks mouse position instantly (with subpixel precision).
 * - Outer ring smoothly interpolates with GSAP quickTo for an organic trailing feel.
 * - Expands dynamically on interactive elements (buttons, links, inputs, cards, [data-cursor-hover]).
 * - Naturally fades out when pointer leaves the browser window.
 * - Automatically disabled on touch devices and when prefers-reduced-motion is active.
 */
export default function CustomCursor() {
  const dotRef = useRef(null);
  const ringRef = useRef(null);
  const isHoveredRef = useRef(false);

  useEffect(() => {
    // 1. Environmental & Capability Checks
    const isFinePointer = window.matchMedia('(hover: hover) and (pointer: fine)').matches;
    const isReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    if (!isFinePointer || isReducedMotion) {
      return;
    }

    const dot = dotRef.current;
    const ring = ringRef.current;
    if (!dot || !ring) return;

    let isVisible = false;

    // 2. GSAP quickTo setters for the lagging ring (avoiding CSS transition conflicts)
    const setRingX = gsap.quickTo(ring, 'x', { duration: 0.22, ease: 'power2.out' });
    const setRingY = gsap.quickTo(ring, 'y', { duration: 0.22, ease: 'power2.out' });

    // GSAP scale setter for ring expansion
    const setRingScale = gsap.quickTo(ring, 'scale', { duration: 0.25, ease: 'power2.out' });

    // 3. Pointer Move Handler
    function handlePointerMove(e) {
      const clientX = e.clientX;
      const clientY = e.clientY;

      if (!isVisible) {
        isVisible = true;
        gsap.to([dot, ring], { opacity: 1, duration: 0.2, overwrite: 'auto' });
      }

      // Dot tracks instantly via direct translate3d (centered at 0, 0 with -50% CSS translate)
      dot.style.transform = `translate3d(${clientX}px, ${clientY}px, 0)`;

      // Ring tracks smoothly with slight lag
      setRingX(clientX);
      setRingY(clientY);
    }

    // 4. Element Hover Tracking (Ring Expansion)
    function handlePointerOver(e) {
      const target = e.target;
      if (!target || !(target instanceof Element)) return;

      const interactive = target.closest(
        'a, button, [role="button"], input, select, textarea, label, [data-cursor-hover], [data-magnetic], .glass-interactive'
      );

      if (interactive && !isHoveredRef.current) {
        isHoveredRef.current = true;
        setRingScale(1.6);
        gsap.to(ring, {
          borderColor: 'rgba(245, 184, 61, 0.9)',
          backgroundColor: 'rgba(245, 184, 61, 0.12)',
          boxShadow: '0 0 16px rgba(245, 184, 61, 0.25)',
          duration: 0.2,
          overwrite: 'auto',
        });
      }
    }

    function handlePointerOut(e) {
      const target = e.target;
      if (!target || !(target instanceof Element)) return;

      const interactive = target.closest(
        'a, button, [role="button"], input, select, textarea, label, [data-cursor-hover], [data-magnetic], .glass-interactive'
      );

      // Check if relatedTarget is still within the same or another interactive element
      const related = e.relatedTarget;
      const isStillInteractive =
        related &&
        related instanceof Element &&
        related.closest(
          'a, button, [role="button"], input, select, textarea, label, [data-cursor-hover], [data-magnetic], .glass-interactive'
        );

      if (interactive && !isStillInteractive && isHoveredRef.current) {
        isHoveredRef.current = false;
        setRingScale(1);
        gsap.to(ring, {
          borderColor: 'rgba(245, 184, 61, 0.4)',
          backgroundColor: 'rgba(245, 184, 61, 0.03)',
          boxShadow: 'none',
          duration: 0.25,
          overwrite: 'auto',
        });
      }
    }

    // 5. Window Leave & Enter Detection
    function handleDocumentLeave(e) {
      if (e.relatedTarget === null || e.clientY <= 0 || e.clientX <= 0 || e.clientX >= window.innerWidth || e.clientY >= window.innerHeight) {
        isVisible = false;
        gsap.to([dot, ring], { opacity: 0, duration: 0.25, overwrite: 'auto' });
      }
    }

    function handleDocumentEnter() {
      isVisible = true;
      gsap.to([dot, ring], { opacity: 1, duration: 0.2, overwrite: 'auto' });
    }

    function handleVisibility() {
      if (document.hidden) {
        isVisible = false;
        gsap.set([dot, ring], { opacity: 0 });
      }
    }

    // Listeners
    window.addEventListener('pointermove', handlePointerMove, { passive: true });
    document.addEventListener('mouseover', handlePointerOver, { passive: true });
    document.addEventListener('mouseout', handlePointerOut, { passive: true });
    document.addEventListener('mouseleave', handleDocumentLeave);
    document.addEventListener('mouseenter', handleDocumentEnter);
    document.addEventListener('visibilitychange', handleVisibility);

    return () => {
      window.removeEventListener('pointermove', handlePointerMove);
      document.removeEventListener('mouseover', handlePointerOver);
      document.removeEventListener('mouseout', handlePointerOut);
      document.removeEventListener('mouseleave', handleDocumentLeave);
      document.removeEventListener('mouseenter', handleDocumentEnter);
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, []);

  return (
    <div
      aria-hidden="true"
      className="pointer-events-none fixed inset-0 z-[9999] overflow-hidden select-none"
    >
      {/* Central Sharp Gold Dot (4px) */}
      <div
        ref={dotRef}
        className="pointer-events-none fixed top-0 left-0 w-1.5 h-1.5 -ml-[3px] -mt-[3px] rounded-full bg-accent shadow-[0_0_8px_rgba(245,184,61,0.8)] opacity-0 will-change-transform"
      />

      {/* Trailing Translucent Gold Ring (32px) */}
      <div
        ref={ringRef}
        className="pointer-events-none fixed top-0 left-0 w-8 h-8 -ml-4 -mt-4 rounded-full border border-accent/40 bg-accent/[0.03] opacity-0 will-change-transform"
      />
    </div>
  );
}
