import React from 'react';

/**
 * Built for India
 * A quiet, elegant strip highlighting localized monetary design, Indian digit grouping, and native payment methods.
 */
export default function IndiaStrip() {
  return (
    <section className="py-12 px-4 border-b border-white/5 bg-white/[0.01]">
      <div className="max-w-[900px] mx-auto flex flex-col sm:flex-row items-center justify-between gap-4 text-xs font-mono text-muted">
        <div className="flex items-center gap-2">
          <span className="w-1.5 h-1.5 rounded-full bg-accent" />
          <span className="text-main font-medium">Built for India:</span>
          <span>Zero fractions, true paise precision, and Indian digit grouping</span>
          <span className="text-accent font-semibold ml-1">(e.g. ₹1,50,000)</span>
        </div>

        <div className="flex items-center gap-2 text-[11px]">
          <span className="px-2.5 py-1 rounded-full bg-white/5 border border-white/10 text-main">
            UPI
          </span>
          <span className="text-muted/40">•</span>
          <span className="px-2.5 py-1 rounded-full bg-white/5 border border-white/10 text-main">
            Cash
          </span>
          <span className="text-muted/40">•</span>
          <span className="px-2.5 py-1 rounded-full bg-white/5 border border-white/10 text-main">
            Card
          </span>
        </div>
      </div>
    </section>
  );
}

