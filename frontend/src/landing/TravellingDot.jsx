import React, { useEffect, useRef } from 'react';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

gsap.registerPlugin(ScrollTrigger);

/**
 * The Hero Travelling Dot
 * A persistent fixed element representing the Spendle dot that travels through
 * the page, morphing and landing on interactive anchor points during scroll.
 */
export default function TravellingDot() {
  const dotRef = useRef(null);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReducedMotion || !dotRef.current) return;

    const dot = dotRef.current;

    // Helper to get anchor viewport coordinates
    function getAnchorCoords(name) {
      const el = document.querySelector(`[data-dot-anchor="${name}"]`);
      if (!el) return null;
      const rect = el.getBoundingClientRect();
      return {
        x: rect.left + rect.width / 2,
        y: rect.top + rect.height / 2,
        width: rect.width || 8,
        height: rect.height || 8,
      };
    }

    // Initialize at Nav anchor
    const navCoords = getAnchorCoords('nav');
    if (navCoords) {
      gsap.set(dot, {
        x: navCoords.x,
        y: navCoords.y,
        width: 7,
        height: 7,
        opacity: 1,
      });
    }

    // Timeline connected to scroll
    const ctx = gsap.context(() => {
      // 1. Move to Hero number drop point
      ScrollTrigger.create({
        trigger: '#hero-section',
        start: 'top top',
        end: 'bottom center',
        scrub: 0.8,
        onUpdate: (self) => {
          const target = getAnchorCoords(self.progress > 0.5 ? 'hero-number' : 'nav');
          if (target) {
            gsap.to(dot, {
              x: target.x,
              y: target.y,
              width: self.progress > 0.5 ? 12 : 7,
              height: self.progress > 0.5 ? 12 : 7,
              duration: 0.2,
              ease: 'power1.out',
              overwrite: 'auto',
            });
          }
        },
      });

      // 2. Move through OneTap + button target
      ScrollTrigger.create({
        trigger: '#onetap-section',
        start: 'top center',
        end: 'bottom center',
        scrub: 0.8,
        onUpdate: () => {
          const target = getAnchorCoords('onetap-fab');
          if (target) {
            gsap.to(dot, {
              x: target.x,
              y: target.y,
              width: 14,
              height: 14,
              duration: 0.2,
              ease: 'power1.out',
              overwrite: 'auto',
            });
          }
        },
      });

      // 3. Move into Rollover Goal coin target
      ScrollTrigger.create({
        trigger: '#rollover-section',
        start: 'top center',
        end: 'bottom center',
        scrub: 0.8,
        onUpdate: () => {
          const target = getAnchorCoords('rollover-coin');
          if (target) {
            gsap.to(dot, {
              x: target.x,
              y: target.y,
              width: 10,
              height: 10,
              duration: 0.2,
              ease: 'power1.out',
              overwrite: 'auto',
            });
          }
        },
      });

      // 4. Land on CTA button dot
      ScrollTrigger.create({
        trigger: '#cta-section',
        start: 'top 80%',
        end: 'bottom bottom',
        scrub: 0.8,
        onUpdate: () => {
          const target = getAnchorCoords('cta-dot');
          if (target) {
            gsap.to(dot, {
              x: target.x,
              y: target.y,
              width: 8,
              height: 8,
              duration: 0.2,
              ease: 'power1.out',
              overwrite: 'auto',
            });
          }
        },
      });
    });

    return () => ctx.revert();
  }, []);

  return (
    <div
      ref={dotRef}
      aria-hidden="true"
      className="fixed top-0 left-0 -translate-x-1/2 -translate-y-1/2 rounded-full bg-accent z-40 pointer-events-none opacity-0 shadow-lg shadow-accent/40"
      style={{
        boxShadow: '0 0 12px #F5B83D',
      }}
    />
  );
}

