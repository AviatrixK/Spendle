import React, { useEffect, useRef, useState } from 'react';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import PhoneMock from './PhoneMock';
import GlassCard from '../components/GlassCard';

gsap.registerPlugin(ScrollTrigger);

const EXPENSES = [
  { id: 'exp-1', day: 3, label: 'House Rent', amount: 9000, cat: 'Rent', mode: 'Netbanking' },
  { id: 'exp-2', day: 7, label: 'BigBasket Groceries', amount: 2400, cat: 'Groceries', mode: 'Card' },
  { id: 'exp-3', day: 12, label: 'Swiggy Dinner', amount: 550, cat: 'Food', mode: 'UPI' },
  { id: 'exp-4', day: 18, label: 'Shell Petrol', amount: 1200, cat: 'Transport', mode: 'Card' },
  { id: 'exp-5', day: 23, label: 'Jio 5G Recharge', amount: 399, cat: 'Bills', mode: 'UPI' },
  { id: 'exp-6', day: 28, label: 'Netflix Premium', amount: 649, cat: 'Entertainment', mode: 'Card' },
];

export default function MonthScroll() {
  const sectionRef = useRef(null);
  const pinContainerRef = useRef(null);
  const [day, setDay] = useState(1);
  const [leftAmount, setLeftAmount] = useState(21000);
  const [bikeRingPercent, setBikeRingPercent] = useState(10);
  const [isRolloverActive, setIsRolloverActive] = useState(false);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReducedMotion || !sectionRef.current || !pinContainerRef.current) return;

    const ctx = gsap.context(() => {
      const tl = gsap.timeline({
        scrollTrigger: {
          trigger: sectionRef.current,
          start: 'top top',
          end: '+=350%',
          pin: pinContainerRef.current,
          scrub: 0.8,
          anticipatePin: 1,
        },
      });

      // Advance day counter 1 -> 30
      const dayObj = { d: 1, left: 21000, bike: 10 };
      tl.to(dayObj, {
        d: 30,
        left: 6802,
        duration: 3,
        ease: 'none',
        onUpdate: () => {
          setDay(Math.round(dayObj.d));
          setLeftAmount(Math.round(dayObj.left));
        },
      });

      // Animate expense chips docking into phone sequentially
      EXPENSES.forEach((exp, i) => {
        tl.fromTo(
          `#dock-chip-${i}`,
          { y: -60, opacity: 0, scale: 0.8 },
          { y: 0, opacity: 1, scale: 1, duration: 0.4, ease: 'power2.out' },
          i * 0.45
        );
      });

      // Day 30 Rollover sweep: leftover lifts off and fills Bike goal ring
      tl.to(
        '#rollover-stream',
        {
          opacity: 1,
          scale: 1.1,
          duration: 0.4,
          onStart: () => setIsRolloverActive(true),
          onReverseComplete: () => setIsRolloverActive(false),
        },
        2.5
      );

      tl.to(dayObj, {
        bike: 85,
        duration: 0.8,
        ease: 'power2.out',
        onUpdate: () => {
          setBikeRingPercent(Math.round(dayObj.bike));
        },
      }, 2.6);
    }, sectionRef);

    return () => ctx.revert();
  }, []);

  // SVG circle circumference for Bike goal ring (radius 18)
  const circleRadius = 18;
  const circumference = 2 * Math.PI * circleRadius;
  const strokeDashoffset = circumference - (bikeRingPercent / 100) * circumference;

  return (
    <section
      id="month-section"
      ref={sectionRef}
      className="relative bg-[#0E0E11] border-t border-b border-white/5"
    >
      <div
        ref={pinContainerRef}
        className="min-h-[100svh] w-full flex flex-col justify-center py-12 px-4 max-w-[1100px] mx-auto"
      >
        {/* Header / Section Tag */}
        <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 mb-8">
          <div>
            <span className="text-[11px] font-mono uppercase tracking-[0.2em] text-accent block mb-1">
              The Signature Sequence
            </span>
            <h2 className="text-3xl sm:text-5xl font-serif text-main tracking-tight">
              A month, in <em className="italic text-accent font-normal">one scroll</em>.
            </h2>
          </div>

          {/* Large Day Counter Indicator */}
          <div className="flex items-baseline gap-2 bg-white/5 border border-white/10 px-4 py-2 rounded-2xl w-fit">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted">Calendar Day</span>
            <span className="text-3xl sm:text-4xl font-serif font-bold text-main tabular-nums">
              {day}
            </span>
            <span className="text-muted text-xs font-mono">/ 30</span>
          </div>
        </div>

        {/* Content Grid: Chips Fall In -> Dock into Phone -> Sweep into Goal */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
          {/* Left Column: Stream of Live Expenses */}
          <div className="lg:col-span-5 space-y-2.5">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted block mb-2">
              Incoming Expenses
            </span>
            <div className="space-y-2">
              {EXPENSES.map((exp, i) => (
                <div
                  key={exp.id}
                  id={`dock-chip-${i}`}
                  className="transition-all"
                >
                  <GlassCard className="py-2.5 px-3.5 rounded-xl flex items-center justify-between text-xs font-mono border-white/10">
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-white/5 text-muted border border-white/5">
                        Day {exp.day}
                      </span>
                      <span className="text-main font-medium">{exp.label}</span>
                    </div>
                    <span className="text-semantic-loss font-semibold">
                      -₹{exp.amount.toLocaleString('en-IN')}
                    </span>
                  </GlassCard>
                </div>
              ))}
            </div>
          </div>

          {/* Right Column: Phone Mock receiving docked items & sweeping to Goal */}
          <div className="lg:col-span-7 flex justify-center relative">
            <PhoneMock className="w-[300px] sm:w-[330px]">
              {/* Month Header */}
              <div className="flex items-center justify-between text-[11px] text-muted pb-3 border-b border-white/5">
                <span className="font-serif text-main">Budget Overview</span>
                <span className="font-mono text-accent text-[10px]">Day {day}</span>
              </div>

              {/* Dynamic Safe to Spend Counter */}
              <div className="text-center py-4 my-auto">
                <span className="text-[10px] font-mono uppercase text-muted block mb-1">
                  Remaining Safe to Spend
                </span>
                <div className="text-3xl sm:text-4xl font-serif text-main font-medium">
                  ₹{leftAmount.toLocaleString('en-IN')}
                </div>
                <div className="text-[10px] font-mono text-muted mt-1">
                  Initial Budget: ₹21,000
                </div>
              </div>

              {/* Day 30 Sweeping Dot Stream Indicator */}
              <div
                id="rollover-stream"
                className={`py-2 px-3 rounded-xl border text-center transition-all ${
                  isRolloverActive
                    ? 'bg-accent/20 border-accent text-accent shadow-lg shadow-accent/20'
                    : 'bg-white/5 border-white/5 text-muted opacity-30'
                }`}
              >
                <span className="text-[11px] font-mono block font-semibold">
                  {day >= 30 ? '✨ Month-End Sweep Triggered!' : 'Sweeps automatically at Day 30'}
                </span>
                <span className="text-[10px] text-muted">
                  Leftover ₹6,802 moving to Priority #1 Goal
                </span>
              </div>

              {/* Target Savings Goal (#1 Bike) with Animated SVG Progress Ring */}
              <div className="bg-white/[0.04] border border-white/10 rounded-2xl p-3.5 mt-3 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {/* SVG Circular Ring */}
                  <div className="relative w-11 h-11 flex items-center justify-center">
                    <svg className="w-11 h-11 -rotate-90">
                      <circle
                        cx="22"
                        cy="22"
                        r={circleRadius}
                        className="stroke-white/10 fill-none"
                        strokeWidth="3.5"
                      />
                      <circle
                        cx="22"
                        cy="22"
                        r={circleRadius}
                        className="stroke-accent fill-none transition-all duration-300"
                        strokeWidth="3.5"
                        strokeDasharray={circumference}
                        strokeDashoffset={strokeDashoffset}
                        strokeLinecap="round"
                      />
                    </svg>
                    <span className="absolute font-mono text-[10px] text-main font-semibold">
                      {bikeRingPercent}%
                    </span>
                  </div>

                  <div>
                    <div className="flex items-center gap-1.5">
                      <span className="text-[10px] font-mono px-1.5 rounded bg-accent/20 text-accent font-semibold">
                        #1
                      </span>
                      <span className="font-serif text-sm text-main font-medium">Bike Target</span>
                    </div>
                    <span className="text-[10px] font-mono text-muted block">
                      Target ₹1,50,000
                    </span>
                  </div>
                </div>

                <span
                  className={`text-xs font-mono font-semibold ${
                    isRolloverActive ? 'text-semantic-profit' : 'text-muted'
                  }`}
                >
                  {isRolloverActive ? '+₹6,802' : '₹0'}
                </span>
              </div>
            </PhoneMock>
          </div>
        </div>
      </div>
    </section>
  );
}

