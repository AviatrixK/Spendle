import React, { useEffect, useRef } from 'react';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import GlassCard from '../components/GlassCard';

gsap.registerPlugin(ScrollTrigger);

const BUDGET_ROWS = [
  { name: 'Groceries', limit: 5000, spent: 3400, percent: 68, over: false },
  { name: 'Transport & Fuel', limit: 2500, spent: 1850, percent: 74, over: false },
  { name: 'Bills & Utilities', limit: 4000, spent: 3100, percent: 77, over: false },
  { name: 'Online Food', limit: 1500, spent: 1950, percent: 130, over: true }, // Overshoots & shakes
];

export default function Budgets() {
  const containerRef = useRef(null);
  const overspentRowRef = useRef(null);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReducedMotion || !containerRef.current) return;

    const ctx = gsap.context(() => {
      // Animate progress bars filling as section scrolls into view
      gsap.from('.budget-bar-fill', {
        scaleX: 0,
        transformOrigin: 'left center',
        duration: 1.1,
        ease: 'power2.out',
        stagger: 0.15,
        scrollTrigger: {
          trigger: containerRef.current,
          start: 'top 75%',
          toggleActions: 'play none none none',
        },
      });

      // Online Food row soft shake effect
      if (overspentRowRef.current) {
        gsap.to(overspentRowRef.current, {
          x: 6,
          duration: 0.08,
          repeat: 5,
          yoyo: true,
          ease: 'power1.inOut',
          delay: 0.8,
          scrollTrigger: {
            trigger: containerRef.current,
            start: 'top 75%',
            toggleActions: 'play none none none',
          },
        });
      }
    }, containerRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      id="budgets-section"
      ref={containerRef}
      className="py-24 px-4 border-b border-white/5"
    >
      <div className="max-w-[1000px] mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
        {/* Editorial Text */}
        <div className="lg:col-span-5 space-y-4">
          <span className="text-[11px] font-mono uppercase tracking-[0.2em] text-accent block">
            02 / Category Intelligence
          </span>
          <h2 className="text-3xl sm:text-5xl font-serif text-main tracking-tight">
            Budgets that <em className="italic text-accent font-normal">talk back</em>.
          </h2>
          <p className="text-muted text-base sm:text-lg leading-relaxed font-sans">
            Set a limit per category, even your own, like Online Food. See what's spent, what's left, and when you've gone over.
          </p>
        </div>

        {/* Live Interactive Budget Visualizer Card */}
        <div className="lg:col-span-7">
          <GlassCard className="p-6 space-y-5 border-white/10">
            <div className="flex items-center justify-between text-xs font-mono text-muted pb-3 border-b border-white/5">
              <span>Category Limits</span>
              <span>Spent / Left</span>
            </div>

            <div className="space-y-4">
              {BUDGET_ROWS.map((row) => {
                const isOver = row.over;
                const leftAmount = row.limit - row.spent;

                return (
                  <div
                    key={row.name}
                    ref={isOver ? overspentRowRef : null}
                    className={`p-3.5 rounded-xl border transition-colors ${
                      isOver
                        ? 'bg-semantic-loss/[0.04] border-semantic-loss/30'
                        : 'bg-white/[0.02] border-white/5'
                    }`}
                  >
                    <div className="flex items-baseline justify-between text-xs font-mono mb-2">
                      <span className="font-medium text-main flex items-center gap-2">
                        {row.name}
                        {isOver && (
                          <span className="text-[10px] px-2 py-0.5 rounded-full bg-semantic-loss/20 text-semantic-loss font-semibold">
                            Overspent
                          </span>
                        )}
                      </span>
                      <span className="text-muted text-[11px]">
                        ₹{row.spent.toLocaleString('en-IN')} of ₹{row.limit.toLocaleString('en-IN')}
                      </span>
                    </div>

                    {/* Progress Bar with overshoot */}
                    <div className="w-full bg-white/10 h-2 rounded-full overflow-hidden">
                      <div
                        className={`budget-bar-fill h-full rounded-full ${
                          isOver ? 'bg-semantic-loss' : 'bg-accent'
                        }`}
                        style={{
                          width: `${Math.min(row.percent, 100)}%`,
                        }}
                      />
                    </div>

                    <div className="flex justify-between items-center text-[11px] font-mono mt-1.5">
                      <span className="text-muted">{row.percent}% limit</span>
                      <span className={isOver ? 'text-semantic-loss font-semibold' : 'text-accent'}>
                        {isOver
                          ? `₹${Math.abs(leftAmount).toLocaleString('en-IN')} over`
                          : `₹${leftAmount.toLocaleString('en-IN')} left`}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </GlassCard>
        </div>
      </div>
    </section>
  );
}

