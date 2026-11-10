import React, { useEffect, useRef, useState } from 'react';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import GlassCard from '../components/GlassCard';

gsap.registerPlugin(ScrollTrigger);

export default function ProfitLoss() {
  const containerRef = useRef(null);
  const [isProfit, setIsProfit] = useState(true);
  const [amount, setAmount] = useState(8000);

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReducedMotion || !containerRef.current) return;

    const ctx = gsap.context(() => {
      // Create a scroll-triggered toggle with smooth number roll
      ScrollTrigger.create({
        trigger: containerRef.current,
        start: 'top 60%',
        onEnter: () => {
          setIsProfit(true);
          const obj = { val: 0 };
          gsap.to(obj, {
            val: 8000,
            duration: 1,
            ease: 'power2.out',
            onUpdate: () => setAmount(Math.round(obj.val)),
          });
        },
      });

      // Toggle to loss on further scroll down or manual timer loop
      const interval = setInterval(() => {
        setIsProfit((prev) => {
          const next = !prev;
          const target = next ? 8000 : 1200;
          const obj = { val: next ? 1200 : 8000 };
          gsap.to(obj, {
            val: target,
            duration: 0.8,
            ease: 'power2.out',
            onUpdate: () => setAmount(Math.round(obj.val)),
          });
          return next;
        });
      }, 4000);

      return () => clearInterval(interval);
    }, containerRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      id="profitloss-section"
      ref={containerRef}
      className="py-24 px-4 border-b border-white/5"
    >
      <div className="max-w-[800px] mx-auto text-center space-y-6">
        <span className="text-[11px] font-mono uppercase tracking-[0.2em] text-accent block">
          04 / Clarity
        </span>
        <h2 className="text-3xl sm:text-5xl font-serif text-main tracking-tight">
          Profit or loss. <em className="italic text-accent font-normal">No spreadsheet</em>.
        </h2>

        {/* Large Transformational Status Banner */}
        <div className="pt-6">
          <GlassCard className="p-8 sm:p-12 text-center border-white/10 relative overflow-hidden transition-colors duration-500">
            <span
              className={`inline-block px-3 py-1 rounded-full text-xs font-mono uppercase tracking-wider font-semibold mb-4 transition-colors duration-300 ${
                isProfit
                  ? 'bg-semantic-profit/15 text-semantic-profit border border-semantic-profit/30'
                  : 'bg-semantic-loss/15 text-semantic-loss border border-semantic-loss/30'
              }`}
            >
              {isProfit ? 'In Profit' : 'In Loss'}
            </span>

            <div
              className={`text-5xl sm:text-7xl font-serif font-medium tracking-tight mb-4 tabular-nums transition-colors duration-300 ${
                isProfit ? 'text-semantic-profit' : 'text-semantic-loss'
              }`}
            >
              {isProfit ? '+' : '−'}₹{amount.toLocaleString('en-IN')}
            </div>

            <p className="text-xs sm:text-sm text-muted font-sans max-w-md mx-auto">
              {isProfit
                ? 'Your monthly income exceeded all category spending. The surplus is ready for your savings goals.'
                : 'Spending crossed your limits this month. Real-time indicators prevent surprises before rent day.'}
            </p>
          </GlassCard>
        </div>
      </div>
    </section>
  );
}

