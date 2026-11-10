import React, { useEffect, useRef } from 'react';
import Lenis from 'lenis';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

import Nav from './Nav';
import Hero from './Hero';
import MonthScroll from './MonthScroll';
import OneTap from './OneTap';
import Budgets from './Budgets';
import Rollover from './Rollover';
import ProfitLoss from './ProfitLoss';
import IndiaStrip from './IndiaStrip';
import CtaFooter from './CtaFooter';
import FilmGrain from './FilmGrain';
import CursorFollower from './CursorFollower';
import TravellingDot from './TravellingDot';

gsap.registerPlugin(ScrollTrigger);

export default function Landing() {
  const lenisRef = useRef(null);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    // Initialize smooth scrolling with Lenis only if reduced-motion is not requested
    if (!prefersReducedMotion) {
      const lenis = new Lenis({
        duration: 1.1,
        easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
        smoothWheel: true,
      });
      lenisRef.current = lenis;

      // Synchronize Lenis with GSAP ScrollTrigger
      lenis.on('scroll', ScrollTrigger.update);

      const tickerCallback = (time) => {
        lenis.raf(time * 1000);
      };

      gsap.ticker.add(tickerCallback);
      gsap.ticker.lagSmoothing(0);

      // Refresh ScrollTrigger after web fonts and layout stabilize
      const timer = setTimeout(() => {
        ScrollTrigger.refresh();
      }, 250);

      return () => {
        clearTimeout(timer);
        gsap.ticker.remove(tickerCallback);
        lenis.destroy();
        ScrollTrigger.getAll().forEach((st) => st.kill());
      };
    } else {
      // For reduced motion, ensure layout coordinates are fresh without smooth scroll
      ScrollTrigger.refresh();
      return () => {
        ScrollTrigger.getAll().forEach((st) => st.kill());
      };
    }
  }, []);

  return (
    <div className="relative min-h-screen bg-[#0E0E11] text-[#EDEAE3] font-sans antialiased overflow-x-hidden selection:bg-accent selection:text-[#0E0E11]">
      {/* Subtle tactile film grain overlay */}
      <FilmGrain />

      {/* Persistent hero travelling gold dot */}
      <TravellingDot />

      {/* Desktop-only lagging cursor follower & magnetic button interactions */}
      <CursorFollower />

      {/* Navigation */}
      <Nav />

      {/* Main Page Flow Sections */}
      <main>
        <Hero />
        <MonthScroll />
        <OneTap />
        <Budgets />
        <Rollover />
        <ProfitLoss />
        <IndiaStrip />
      </main>

      {/* Footer and final CTA */}
      <CtaFooter />
    </div>
  );
}

