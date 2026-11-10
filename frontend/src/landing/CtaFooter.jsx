import React from 'react';
import { Link } from 'react-router-dom';

/**
 * Final Call-to-Action & Capstone Footer
 * Features the signature "Kitna bacha?" serif headline and the final landing spot for the travelling gold dot.
 */
export default function CtaFooter() {
  return (
    <footer id="cta-section" className="pt-24 pb-16 px-4 text-center relative overflow-hidden">
      {/* Subtle ambient gold radial background */}
      <div
        aria-hidden="true"
        className="absolute bottom-0 left-1/2 -translate-x-1/2 w-[540px] h-[340px] rounded-full bg-accent/5 blur-[120px] pointer-events-none"
      />

      <div className="max-w-[720px] mx-auto space-y-6 relative z-10 mb-16">
        <span className="text-[11px] font-mono uppercase tracking-[0.2em] text-accent block">
          One Honest Question
        </span>

        {/* Giant Serif Headline */}
        <h2
          className="text-main font-serif tracking-tight leading-[0.9]"
          style={{ fontSize: 'clamp(3.5rem, 12vw, 8rem)' }}
        >
          Kitna bacha<span className="text-accent italic font-normal">?</span>
        </h2>

        <p className="text-muted text-base sm:text-lg max-w-md mx-auto font-sans">
          Stop guessing your bank balance. Turn your month into clear, actionable financial control.
        </p>

        {/* Big Swelling CTA Button with Anchor for Travelling Dot */}
        <div className="pt-4 flex items-center justify-center">
          <Link
            to="/login"
            data-magnetic
            className="group relative inline-flex items-center gap-2.5 px-8 py-4 rounded-full bg-accent text-[#0E0E11] text-base font-semibold hover:bg-accent-hover transition-all active:scale-95 shadow-xl shadow-accent/25 hover:shadow-accent/40"
          >
            <span>Get started with Spendle</span>
            <span
              data-dot-anchor="cta-dot"
              className="w-2.5 h-2.5 rounded-full bg-[#0E0E11] inline-block transition-transform group-hover:scale-125"
            />
          </Link>
        </div>
      </div>

      {/* Capstone University Attribution Footer */}
      <div className="max-w-[800px] mx-auto pt-8 border-t border-white/5 text-xs font-mono text-muted/70">
        <p>
          A capstone project by <strong className="text-muted font-medium">Mayank Rawat &amp; Kiruthika</strong> ·{' '}
          <strong className="text-muted font-medium">K.R. Mangalam University</strong> · Built with Go, PostgreSQL and React.
        </p>
      </div>
    </footer>
  );
}

