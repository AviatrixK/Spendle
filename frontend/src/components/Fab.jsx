import React from 'react';

export default function Fab({ onAddExpense, onAddIncome }) {
  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-30 flex items-center gap-3 bg-[#0E0E11]/80 backdrop-blur-md px-4 py-2 rounded-full border border-white/10 shadow-2xl shadow-black/80">
      <button
        type="button"
        onClick={onAddExpense}
        className="flex items-center gap-1.5 px-4 py-2.5 rounded-full bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover transition-all active:scale-95 shadow-lg shadow-accent/20 focus:outline-none focus:ring-2 focus:ring-accent"
      >
        <span className="text-base leading-none font-bold">+</span>
        <span>Expense</span>
      </button>

      <button
        type="button"
        onClick={onAddIncome}
        className="flex items-center gap-1.5 px-4 py-2.5 rounded-full bg-white/10 hover:bg-white/15 text-main font-semibold text-sm transition-all active:scale-95 border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/30"
      >
        <span className="text-base leading-none font-bold">+</span>
        <span>Income</span>
      </button>
    </div>
  );
}

