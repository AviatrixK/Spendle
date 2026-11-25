import React from 'react';
import { Link, useLocation } from 'react-router-dom';

export default function Fab({ onHome, onAddExpense, onAddIncome }) {
  const location = useLocation();
  const isHome = location.pathname === '/';
  return (
    <nav className="floating-dock" aria-label="Main navigation">
      <button type="button" className={`dock-item ${isHome ? 'active' : ''}`} onClick={onHome} aria-current={isHome ? 'page' : undefined}>
        <span className="dock-icon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m3.5 10 8.5-7 8.5 7v10a1 1 0 0 1-1 1h-5.2v-6.5H9.7V21H4.5a1 1 0 0 1-1-1z" /></svg></span><span>Home</span>
      </button>
      <button type="button" className="dock-item dock-expense" onClick={onAddExpense}>
        <span className="dock-icon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg></span><span>Expense</span>
      </button>
      <button type="button" className="dock-item dock-income" onClick={onAddIncome}>
        <span className="dock-icon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 19V5m0 0L6.5 10.5M12 5l5.5 5.5" /></svg></span><span>Income</span>
      </button>
      <Link to="/goals" className={`dock-item ${location.pathname === '/goals' ? 'active' : ''}`} aria-current={location.pathname === '/goals' ? 'page' : undefined}>
        <span className="dock-icon"><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8.5" /><circle cx="12" cy="12" r="4.5" /><circle cx="12" cy="12" r="1" /></svg></span><span>Goals</span>
      </Link>
    </nav>
  );
}
