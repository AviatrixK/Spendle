import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

/**
 * Editorial Landing Navigation
 * Features the wordmark with the travelling dot start anchor and hides on scroll down.
 */
export default function Nav() {
  const [isVisible, setIsVisible] = useState(true);
  const [hasScrolled, setHasScrolled] = useState(false);

  useEffect(() => {
    let lastScrollY = window.scrollY;

    function handleScroll() {
      const currentScrollY = window.scrollY;
      setHasScrolled(currentScrollY > 20);

      if (currentScrollY > lastScrollY && currentScrollY > 80) {
        // Scrolling down -> hide
        setIsVisible(false);
      } else {
        // Scrolling up -> show
        setIsVisible(true);
      }
      lastScrollY = currentScrollY;
    }

    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <nav
      className={`fixed top-0 left-0 right-0 z-30 transition-transform duration-300 ${
        isVisible ? 'translate-y-0' : '-translate-y-full'
      } ${hasScrolled ? 'bg-[#0E0E11]/80 backdrop-blur-md border-b border-white/5 py-3' : 'py-5'}`}
    >
      <div className="max-w-[1100px] mx-auto px-6 flex items-center justify-between">
        {/* Wordmark with Travelling Dot Anchor */}
        <Link to="/" className="flex items-baseline gap-0.5 group">
          <span className="font-serif text-2xl tracking-tight text-main font-medium">
            Spendle
          </span>
          <span
            data-dot-anchor="nav"
            className="w-1.5 h-1.5 rounded-full bg-accent inline-block ml-0.5"
          />
        </Link>

        {/* Action Link to /login */}
        <div className="flex items-center gap-4">
          <Link
            to="/login"
            className="text-xs font-mono uppercase tracking-wider text-muted hover:text-main transition-colors hidden sm:inline-block"
          >
            Sign in
          </Link>
          <Link
            to="/login"
            data-magnetic
            className="px-4 py-2 rounded-full bg-accent text-[#0E0E11] text-xs font-semibold hover:bg-accent-hover transition-all active:scale-95 shadow-md shadow-accent/20"
          >
            Get started
          </Link>
        </div>
      </div>
    </nav>
  );
}

