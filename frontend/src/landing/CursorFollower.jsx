import React, { useEffect, useRef } from 'react';
import gsap from 'gsap';

/**
 * Desktop-only subtle cursor follower and magnetic button effect.
 * Strictly disabled on touch devices and when prefers-reduced-motion is active.
 */
export default function CursorFollower() {
  const followerRef = useRef(null);

  useEffect(() => {
    // Only run on fine-pointer devices (desktop mice, trackpads)
    const isFinePointer = window.matchMedia('(hover: hover) and (pointer: fine)').matches;
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    if (!isFinePointer || prefersReducedMotion || !followerRef.current) return;

    const follower = followerRef.current;
    let mouseX = -100;
    let mouseY = -100;

    // Quick setters for smooth 60fps tracking
    const setX = gsap.quickTo(follower, 'x', { duration: 0.35, ease: 'power3.out' });
    const setY = gsap.quickTo(follower, 'y', { duration: 0.35, ease: 'power3.out' });

    function handleMouseMove(e) {
      mouseX = e.clientX;
      mouseY = e.clientY;
      setX(mouseX);
      setY(mouseY);
    }

    window.addEventListener('mousemove', handleMouseMove);

    // Magnetic buttons handler
    const magneticEls = document.querySelectorAll('[data-magnetic]');
    const cleanups = [];

    magneticEls.forEach((el) => {
      const moveMagnet = (e) => {
        const rect = el.getBoundingClientRect();
        const cx = rect.left + rect.width / 2;
        const cy = rect.top + rect.height / 2;
        const dx = (e.clientX - cx) * 0.25;
        const dy = (e.clientY - cy) * 0.25;

        gsap.to(el, { x: dx, y: dy, duration: 0.3, ease: 'power2.out', overwrite: 'auto' });
      };

      const resetMagnet = () => {
        gsap.to(el, { x: 0, y: 0, duration: 0.45, ease: 'elastic.out(1, 0.4)', overwrite: 'auto' });
      };

      el.addEventListener('mousemove', moveMagnet);
      el.addEventListener('mouseleave', resetMagnet);

      cleanups.push(() => {
        el.removeEventListener('mousemove', moveMagnet);
        el.removeEventListener('mouseleave', resetMagnet);
      });
    });

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      cleanups.forEach((c) => c());
    };
  }, []);

  return (
    <div
      ref={followerRef}
      aria-hidden="true"
      className="hidden md:block fixed top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-8 h-8 rounded-full border border-accent/30 pointer-events-none z-40 transition-opacity duration-300"
      style={{
        boxShadow: '0 0 16px rgba(245, 184, 61, 0.12)',
      }}
    />
  );
}

