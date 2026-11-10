import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import gsap from 'gsap';
import PhoneMock from './PhoneMock';
import GlassCard from '../components/GlassCard';

export default function Hero() {
  const containerRef = useRef(null);
  const headlineRef = useRef(null);
  const chipsRef = useRef(null);
  const [displayedLeft, setDisplayedLeft] = useState(0);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    // Headline slide-up masked reveal
    const ctx = gsap.context(() => {
      if (!prefersReducedMotion) {
        gsap.from('.hero-line-inner', {
          yPercent: 120,
          duration: 0.9,
          ease: 'expo.out',
          stagger: 0.08,
          delay: 0.2,
        });

        gsap.from('.hero-fade-in', {
          opacity: 0,
          y: 20,
          duration: 0.8,
          ease: 'power2.out',
          stagger: 0.1,
          delay: 0.5,
        });

        // Floating chips subtle y-drift
        gsap.to('.hero-chip-1', {
          y: -12,
          duration: 2.8,
          ease: 'power1.inOut',
          repeat: -1,
          yoyo: true,
        });
        gsap.to('.hero-chip-2', {
          y: 10,
          duration: 3.4,
          ease: 'power1.inOut',
          repeat: -1,
          yoyo: true,
          delay: 0.5,
        });
        gsap.to('.hero-chip-3', {
          y: -8,
          duration: 3.1,
          ease: 'power1.inOut',
          repeat: -1,
          yoyo: true,
          delay: 1,
        });
      }

      // Count up to ₹4,250 over 1.6s
      const counterObj = { val: 0 };
      gsap.to(counterObj, {
        val: 4250,
        duration: prefersReducedMotion ? 0 : 1.6,
        ease: 'power2.out',
        delay: 0.4,
        onUpdate: () => {
          setDisplayedLeft(Math.round(counterObj.val));
        },
      });

      // Desktop pointer-based parallax for chips
      const isFinePointer = window.matchMedia('(hover: hover) and (pointer: fine)').matches;
      if (isFinePointer && !prefersReducedMotion && containerRef.current) {
        const handleMouseMove = (e) => {
          const rect = containerRef.current.getBoundingClientRect();
          const x = (e.clientX - rect.left - rect.width / 2) / (rect.width / 2);
          const y = (e.clientY - rect.top - rect.height / 2) / (rect.height / 2);

          gsap.to('.hero-chip-1', { x: x * 15, duration: 0.6, ease: 'power1.out', overwrite: 'auto' });
          gsap.to('.hero-chip-2', { x: -x * 20, duration: 0.6, ease: 'power1.out', overwrite: 'auto' });
          gsap.to('.hero-chip-3', { x: x * 10, duration: 0.6, ease: 'power1.out', overwrite: 'auto' });
        };

        const container = containerRef.current;
        container.addEventListener('mousemove', handleMouseMove);
        return () => container.removeEventListener('mousemove', handleMouseMove);
      }
    }, containerRef);

    return () => ctx.revert();
  }, []);

  function handleScrollDown(e) {
    e.preventDefault();
    const target = document.getElementById('month-section');
    if (target) {
      target.scrollIntoView({ behavior: 'smooth' });
    }
  }

  return (
    <section
      id="hero-section"
      ref={containerRef}
      className="relative min-h-[100svh] pt-28 pb-16 px-4 flex flex-col items-center justify-center overflow-hidden"
    >
      {/* Background warm radial accent bloom */}
      <div
        aria-hidden="true"
        className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[480px] h-[480px] rounded-full bg-accent/5 blur-[120px] pointer-events-none"
      />

      <div className="max-w-[840px] mx-auto text-center z-10 space-y-6 mb-12">
        {/* Eyebrow Label */}
        <div className="overflow-hidden">
          <p className="hero-fade-in text-[11px] font-mono uppercase tracking-[0.2em] text-muted">
            Personal Finance / Zero Spreadsheets
          </p>
        </div>

        {/* Large Editorial Headline with Masked Reveal */}
        <h1
          ref={headlineRef}
          className="text-main font-serif tracking-tight leading-[0.95]"
          style={{ fontSize: 'clamp(3.25rem, 11vw, 7.5rem)' }}
        >
          <span className="block overflow-hidden py-1">
            <span className="hero-line-inner block">
              Know what's <em className="italic text-accent font-normal">left</em>.
            </span>
          </span>
        </h1>

        {/* Subline */}
        <p className="hero-fade-in text-muted text-base sm:text-lg max-w-[540px] mx-auto font-sans leading-relaxed">
          Spendle turns your month into one honest number: what you can still spend, and what you've already saved.
        </p>

        {/* CTA Buttons */}
        <div className="hero-fade-in flex flex-wrap items-center justify-center gap-4 pt-2">
          <Link
            to="/login"
            data-magnetic
            className="px-6 py-3.5 rounded-full bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover transition-all active:scale-95 shadow-xl shadow-accent/25"
          >
            Get started
          </Link>
          <a
            href="#month-section"
            onClick={handleScrollDown}
            className="px-6 py-3.5 rounded-full bg-white/5 hover:bg-white/10 text-main border border-white/10 text-sm font-medium transition-all"
          >
            See how a month works ↓
          </a>
        </div>
      </div>

      {/* Responsive Phone Mock Container with Floating Depth Chips */}
      <div className="relative w-full max-w-[420px] mx-auto mt-4">
        {/* Floating Glass Chips (parallax depths) */}
        <div ref={chipsRef} aria-hidden="true" className="pointer-events-none">
          <div className="hero-chip-1 absolute -top-4 -left-6 sm:-left-12 z-20">
            <GlassCard className="py-1.5 px-3 rounded-full text-[11px] font-mono border-white/15 bg-[#16161B]/80 backdrop-blur-md shadow-lg">
              <span className="text-muted mr-1.5">Zomato</span>
              <span className="text-main font-medium">₹340</span>
            </GlassCard>
          </div>

          <div className="hero-chip-2 absolute top-1/3 -right-6 sm:-right-10 z-20">
            <GlassCard className="py-1.5 px-3 rounded-full text-[11px] font-mono border-white/15 bg-[#16161B]/80 backdrop-blur-md shadow-lg">
              <span className="text-muted mr-1.5">Metro card</span>
              <span className="text-main font-medium">₹60</span>
            </GlassCard>
          </div>

          <div className="hero-chip-3 absolute -bottom-3 -left-4 sm:-left-8 z-20">
            <GlassCard className="py-1.5 px-3 rounded-full text-[11px] font-mono border-white/15 bg-[#16161B]/80 backdrop-blur-md shadow-lg">
              <span className="text-muted mr-1.5">Netflix</span>
              <span className="text-main font-medium">₹199</span>
            </GlassCard>
          </div>
        </div>

        {/* Phone Mock showing Live Counting Number */}
        <PhoneMock>
          {/* Mock Top bar */}
          <div className="flex items-center justify-between text-[11px] text-muted pb-3 border-b border-white/5">
            <span className="font-serif text-main text-sm">October 2026</span>
            <span className="font-mono text-semantic-profit px-2 py-0.5 rounded-full bg-semantic-profit/10 text-[10px]">
              In profit
            </span>
          </div>

          {/* Big Number Left Card */}
          <div className="my-auto py-6 text-center">
            <span className="text-[10px] font-mono uppercase tracking-widest text-muted block mb-1">
              Safe to spend
            </span>
            <div
              data-dot-anchor="hero-number"
              className="text-4xl sm:text-5xl font-serif text-main tracking-tight font-medium"
            >
              ₹{displayedLeft.toLocaleString('en-IN')}{' '}
              <span className="text-accent text-2xl font-serif italic">left</span>
            </div>
            <div className="flex justify-center gap-4 mt-3 text-[11px] font-mono text-muted">
              <span>Income: ₹45,000</span>
              <span>Spent: ₹12,750</span>
            </div>
          </div>

          {/* Mock Mini Budget Row */}
          <div className="bg-white/[0.03] border border-white/5 rounded-2xl p-3 space-y-2">
            <div className="flex justify-between text-[11px] font-mono">
              <span className="text-main">Food & Dining</span>
              <span className="text-muted">₹3,420 of ₹6,000</span>
            </div>
            <div className="w-full bg-white/10 h-1.5 rounded-full overflow-hidden">
              <div className="h-full bg-accent rounded-full w-[57%]" />
            </div>
          </div>
        </PhoneMock>
      </div>
    </section>
  );
}

