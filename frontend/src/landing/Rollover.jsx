import React, { useEffect, useRef, useState } from 'react';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import GlassCard from '../components/GlassCard';

gsap.registerPlugin(ScrollTrigger);

export default function Rollover() {
  const containerRef = useRef(null);
  const [goal1Progress, setGoal1Progress] = useState(65);
  const [goal2Progress, setGoal2Progress] = useState(20);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReducedMotion || !containerRef.current) return;

    const ctx = gsap.context(() => {
      const progressObj = { p1: 65, p2: 20 };

      // On scroll into view, dots spill into goal 1 until 100%, then spill into goal 2
      gsap.to(progressObj, {
        p1: 100,
        p2: 45,
        duration: 1.6,
        ease: 'power2.out',
        scrollTrigger: {
          trigger: containerRef.current,
          start: 'top 70%',
          toggleActions: 'play none none none',
        },
        onUpdate: () => {
          setGoal1Progress(Math.round(progressObj.p1));
          setGoal2Progress(Math.round(progressObj.p2));
        },
      });

      // Subtle cascading entry for goal cards
      gsap.from('.goal-cascade-card', {
        y: 30,
        opacity: 0,
        duration: 0.8,
        stagger: 0.15,
        ease: 'power2.out',
        scrollTrigger: {
          trigger: containerRef.current,
          start: 'top 75%',
          toggleActions: 'play none none none',
        },
      });
    }, containerRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      id="rollover-section"
      ref={containerRef}
      className="py-24 px-4 border-b border-white/5"
    >
      <div className="max-w-[1000px] mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
        {/* Editorial Text */}
        <div className="lg:col-span-5 space-y-4">
          <span className="text-[11px] font-mono uppercase tracking-[0.2em] text-accent block">
            03 / Automated Sweep
          </span>
          <h2 className="text-3xl sm:text-5xl font-serif text-main tracking-tight">
            Your leftover has <em className="italic text-accent font-normal">a job</em>.
          </h2>
          <p className="text-muted text-base sm:text-lg leading-relaxed font-sans">
            At month end, whatever you didn't spend flows into your goals: #1 first, then #2. Automatically.
          </p>
        </div>

        {/* Stacked Priority Goals Cards */}
        <div className="lg:col-span-7 space-y-3 relative">
          {/* Target Coin Anchor */}
          <div
            data-dot-anchor="rollover-coin"
            className="absolute -top-3 left-8 w-2 h-2 opacity-0 pointer-events-none"
          />

          {/* Goal 1: Bike */}
          <div className="goal-cascade-card">
            <GlassCard className="p-4 border-white/10 hover:border-accent/40 transition-colors">
              <div className="flex items-center justify-between text-xs font-mono mb-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 rounded-full bg-accent/20 text-accent font-semibold text-[10px]">
                    Priority #1
                  </span>
                  <span className="font-serif text-base text-main font-medium">Royal Enfield Bike</span>
                </div>
                <span className="text-accent font-semibold">{goal1Progress}%</span>
              </div>
              <div className="w-full bg-white/10 h-2 rounded-full overflow-hidden mb-2">
                <div
                  className="h-full bg-accent rounded-full transition-all duration-300"
                  style={{ width: `${goal1Progress}%` }}
                />
              </div>
              <div className="flex justify-between text-[11px] font-mono text-muted">
                <span>Target: ₹1,50,000</span>
                <span>{goal1Progress === 100 ? '🎉 Fully Funded & Completed!' : 'Remaining: ₹52,500'}</span>
              </div>
            </GlassCard>
          </div>

          {/* Goal 2: MacBook Air */}
          <div className="goal-cascade-card">
            <GlassCard className="p-4 border-white/10 hover:border-white/20 transition-colors">
              <div className="flex items-center justify-between text-xs font-mono mb-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 rounded-full bg-white/10 text-muted font-semibold text-[10px]">
                    Priority #2
                  </span>
                  <span className="font-serif text-base text-main font-medium">MacBook Air M3</span>
                </div>
                <span className="text-accent font-semibold">{goal2Progress}%</span>
              </div>
              <div className="w-full bg-white/10 h-2 rounded-full overflow-hidden mb-2">
                <div
                  className="h-full bg-accent rounded-full transition-all duration-300"
                  style={{ width: `${goal2Progress}%` }}
                />
              </div>
              <div className="flex justify-between text-[11px] font-mono text-muted">
                <span>Target: ₹90,000</span>
                <span>Spillover sweep: +₹2,800</span>
              </div>
            </GlassCard>
          </div>

          {/* Goal 3: Goa Trip */}
          <div className="goal-cascade-card">
            <GlassCard className="p-4 border-white/5 opacity-80">
              <div className="flex items-center justify-between text-xs font-mono mb-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 rounded-full bg-white/5 text-muted text-[10px]">
                    Priority #3
                  </span>
                  <span className="font-serif text-base text-main font-medium">Goa Vacation</span>
                </div>
                <span className="text-muted">0%</span>
              </div>
              <div className="w-full bg-white/5 h-2 rounded-full overflow-hidden mb-2">
                <div className="h-full bg-white/20 rounded-full w-[0%]" />
              </div>
              <div className="flex justify-between text-[11px] font-mono text-muted">
                <span>Target: ₹25,000</span>
                <span>Queued for next rollover</span>
              </div>
            </GlassCard>
          </div>
        </div>
      </div>
    </section>
  );
}

