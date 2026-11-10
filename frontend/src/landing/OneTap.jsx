import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import GlassCard from '../components/GlassCard';
import PhoneMock from './PhoneMock';

/**
 * 01 / One tap. Done.
 * Choreographed interactive phone simulation demonstrating the bottom-sheet logging workflow.
 */
export default function OneTap() {
  const [step, setStep] = useState(0); // 0: idle, 1: ripple, 2: sheet up, 3: typed 340, 4: chip active, 5: saved/sheet drops
  const [spentDisplay, setSpentDisplay] = useState(12400);
  const timerRef = useRef(null);

  function runSimulation() {
    clearTimeout(timerRef.current);
    setStep(0);
    setSpentDisplay(12400);

    // Timeline step execution
    timerRef.current = setTimeout(() => {
      setStep(1); // Tap ripple
      timerRef.current = setTimeout(() => {
        setStep(2); // Sheet slides up
        timerRef.current = setTimeout(() => {
          setStep(3); // Amount types 340
          timerRef.current = setTimeout(() => {
            setStep(4); // UPI chip selected
            timerRef.current = setTimeout(() => {
              setStep(5); // Save pressed & sheet drops
              setSpentDisplay(12740); // Number updates
              // Loop every ~7s
              timerRef.current = setTimeout(runSimulation, 3200);
            }, 700);
          }, 600);
        }, 800);
      }, 350);
    }, 600);
  }

  useEffect(() => {
    runSimulation();
    return () => clearTimeout(timerRef.current);
  }, []);

  return (
    <section
      id="onetap-section"
      className="py-24 px-4 border-b border-white/5 relative overflow-hidden"
    >
      <div className="max-w-[1000px] mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
        {/* Editorial Text Column */}
        <div className="lg:col-span-5 space-y-4">
          <span className="text-[11px] font-mono uppercase tracking-[0.2em] text-accent block">
            01 / Fast Logging
          </span>
          <h2 className="text-3xl sm:text-5xl font-serif text-main tracking-tight">
            One tap. <em className="italic text-accent font-normal">Done</em>.
          </h2>
          <p className="text-muted text-base sm:text-lg leading-relaxed font-sans">
            Amount, chip, save. Logging a chai shouldn't take longer than drinking it.
          </p>
          <div className="pt-2 text-xs font-mono text-muted flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-semantic-profit animate-pulse" />
            <span>Tap phone screen to replay workflow</span>
          </div>
        </div>

        {/* Interactive Simulated Phone */}
        <div className="lg:col-span-7 flex justify-center">
          <div
            onClick={runSimulation}
            className="cursor-pointer group transition-transform active:scale-[0.99]"
            title="Click to replay animation"
          >
            <PhoneMock className="w-[300px] sm:w-[330px]">
              {/* Header */}
              <div className="flex items-center justify-between text-[11px] text-muted pb-3 border-b border-white/5">
                <span className="font-serif text-main">Spendle Dashboard</span>
                <span className="font-mono text-semantic-profit text-[10px]">+₹5,000 net</span>
              </div>

              {/* Background Dashboard Number */}
              <div className="text-center py-6 my-auto">
                <span className="text-[10px] font-mono uppercase text-muted block mb-1">
                  Monthly Expenses
                </span>
                <div className="text-3xl sm:text-4xl font-serif text-main font-medium tabular-nums">
                  ₹{spentDisplay.toLocaleString('en-IN')}
                </div>
                <span className="text-[10px] font-mono text-muted block mt-1">
                  {step === 5 ? '✓ Updated with ₹340 Chai & Snacks' : 'Waiting for log...'}
                </span>
              </div>

              {/* Simulated Floating Action Button */}
              <div className="relative pt-4 flex justify-center">
                <div
                  data-dot-anchor="onetap-fab"
                  className={`relative px-5 py-2.5 rounded-full bg-accent text-[#0E0E11] text-xs font-semibold flex items-center gap-1.5 shadow-lg shadow-accent/25 transition-transform ${
                    step === 1 ? 'scale-90 ring-4 ring-accent/30' : ''
                  }`}
                >
                  <span className="text-sm font-bold">+</span>
                  <span>Expense</span>
                  {step === 1 && (
                    <motion.div
                      initial={{ scale: 0.8, opacity: 0.8 }}
                      animate={{ scale: 1.8, opacity: 0 }}
                      transition={{ duration: 0.4 }}
                      className="absolute inset-0 rounded-full border-2 border-accent"
                    />
                  )}
                </div>
              </div>

              {/* Animated Bottom Quick-Add Sheet Simulation */}
              <AnimatePresence>
                {step >= 2 && step <= 4 && (
                  <motion.div
                    initial={{ y: '100%' }}
                    animate={{ y: 0 }}
                    exit={{ y: '100%' }}
                    transition={{ type: 'spring', damping: 25, stiffness: 280 }}
                    className="absolute inset-x-0 bottom-0 bg-[#16161B] border-t border-white/10 rounded-t-3xl p-4 shadow-2xl z-20 space-y-3"
                  >
                    <div className="w-8 h-1 bg-white/20 rounded-full mx-auto" />

                    <div className="text-center">
                      <span className="text-[10px] font-mono text-muted uppercase">Amount (₹)</span>
                      <div className="text-3xl font-serif text-main">
                        ₹{step >= 3 ? '340' : '|'}
                      </div>
                    </div>

                    {/* Category Chip */}
                    <div className="flex justify-center gap-1.5 text-[11px] font-mono">
                      <span className="px-2.5 py-1 rounded-full bg-accent text-[#0E0E11] font-semibold">
                        Food
                      </span>
                      <span className="px-2.5 py-1 rounded-full bg-white/5 text-muted border border-white/5">
                        Transport
                      </span>
                      <span className="px-2.5 py-1 rounded-full bg-white/5 text-muted border border-white/5">
                        Bills
                      </span>
                    </div>

                    {/* Payment Chip */}
                    <div className="flex justify-center gap-1.5 text-[10px] font-mono">
                      <span
                        className={`px-3 py-1 rounded-lg border transition-colors ${
                          step >= 4
                            ? 'bg-white/15 border-accent text-accent font-semibold'
                            : 'bg-white/5 border-white/10 text-muted'
                        }`}
                      >
                        UPI
                      </span>
                      <span className="px-3 py-1 rounded-lg bg-white/5 border border-white/10 text-muted">
                        Cash
                      </span>
                      <span className="px-3 py-1 rounded-lg bg-white/5 border border-white/10 text-muted">
                        Card
                      </span>
                    </div>

                    {/* Save Button */}
                    <div
                      className={`w-full py-2 rounded-xl text-center text-xs font-semibold transition-all ${
                        step >= 4
                          ? 'bg-accent text-[#0E0E11] shadow-md shadow-accent/30 scale-95'
                          : 'bg-accent/80 text-[#0E0E11]'
                      }`}
                    >
                      Save Expense
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </PhoneMock>
          </div>
        </div>
      </div>
    </section>
  );
}

