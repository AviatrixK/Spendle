import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { rupeesToPaise } from '../lib/format';
import { transactionApi } from '../api';

const PAYMENT_MODES = [
  { id: 'upi', label: 'UPI' },
  { id: 'cash', label: 'Cash' },
  { id: 'card', label: 'Card' },
];

export default function QuickAddSheet({
  isOpen,
  onClose,
  initialType = 'expense',
  categories = [],
  onSuccess,
}) {
  const [type, setType] = useState(initialType);
  const [amount, setAmount] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [paymentMode, setPaymentMode] = useState('upi');
  const [note, setNote] = useState('');
  const [occurredOn, setOccurredOn] = useState(() => new Date().toISOString().slice(0, 10));
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const inputRef = useRef(null);

  // Sync type and reset form when sheet opens
  useEffect(() => {
    if (isOpen) {
      setType(initialType);
      setAmount('');
      setNote('');
      setError('');
      setOccurredOn(new Date().toISOString().slice(0, 10));
      setPaymentMode('upi');
      // Pre-select first category if available for expenses
      if (categories.length > 0) {
        setCategoryId(categories[0].id);
      } else {
        setCategoryId('');
      }

      // Auto-focus amount field after animation
      const timer = setTimeout(() => {
        if (inputRef.current) inputRef.current.focus();
      }, 150);
      return () => clearTimeout(timer);
    }
  }, [isOpen, initialType, categories]);

  if (!isOpen) return null;

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');

    const paise = rupeesToPaise(amount);
    if (paise <= 0) {
      setError('Please enter an amount greater than ₹0');
      return;
    }

    if (type === 'expense' && !categoryId) {
      setError('Please select or create a category for this expense');
      return;
    }

    try {
      setIsSubmitting(true);
      await transactionApi.create({
        type,
        amount_paise: paise,
        category_id: type === 'expense' ? categoryId : null,
        payment_mode: paymentMode,
        note: note.trim(),
        occurred_on: occurredOn,
      });

      if (onSuccess) onSuccess();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to save transaction');
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-end justify-center">
        {/* Backdrop */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm"
        />

        {/* Bottom Sheet */}
        <motion.div
          initial={{ y: '100%' }}
          animate={{ y: 0 }}
          exit={{ y: '100%' }}
          transition={{ type: 'spring', damping: 25, stiffness: 280 }}
          className="relative w-full max-w-[540px] bg-[#16161B] border-t border-white/10 rounded-t-3xl p-6 shadow-2xl z-10 max-h-[90vh] overflow-y-auto"
        >
          {/* Grab handle */}
          <div className="w-10 h-1 bg-white/20 rounded-full mx-auto mb-5" />

          {/* Header & Type Toggle */}
          <div className="flex items-center justify-between mb-5">
            <div className="flex bg-white/5 p-1 rounded-full border border-white/10">
              <button
                type="button"
                onClick={() => setType('expense')}
                className={`px-4 py-1.5 rounded-full text-xs font-medium transition-all ${
                  type === 'expense'
                    ? 'bg-accent text-[#0E0E11] font-semibold'
                    : 'text-muted hover:text-main'
                }`}
              >
                Expense
              </button>
              <button
                type="button"
                onClick={() => setType('income')}
                className={`px-4 py-1.5 rounded-full text-xs font-medium transition-all ${
                  type === 'income'
                    ? 'bg-semantic-profit text-[#0E0E11] font-semibold'
                    : 'text-muted hover:text-main'
                }`}
              >
                Income
              </button>
            </div>

            <button
              type="button"
              onClick={onClose}
              className="text-muted hover:text-main p-1.5 rounded-full hover:bg-white/5"
              aria-label="Close"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Big Amount Field */}
            <div className="bg-white/[0.03] border border-white/10 rounded-2xl p-4 text-center focus-within:border-accent/60 transition-colors">
              <label htmlFor="amount-input" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                Amount (₹)
              </label>
              <div className="flex items-center justify-center">
                <span className="text-3xl font-serif text-accent mr-1">₹</span>
                <input
                  id="amount-input"
                  ref={inputRef}
                  type="number"
                  step="any"
                  inputMode="decimal"
                  placeholder="0"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="bg-transparent text-4xl font-serif font-medium text-main text-center outline-none w-48 placeholder:text-white/20"
                  required
                />
              </div>
            </div>

            {/* Category Chips (Expenses only) */}
            {type === 'expense' && (
              <div>
                <label className="block text-xs uppercase tracking-wider text-muted mb-2 font-mono">
                  Category
                </label>
                {categories.length === 0 ? (
                  <p className="text-xs text-muted">No categories found. Create one first from Budgets.</p>
                ) : (
                  <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto pr-1">
                    {categories.map((c) => {
                      const isSelected = categoryId === c.id;
                      return (
                        <button
                          key={c.id}
                          type="button"
                          onClick={() => setCategoryId(c.id)}
                          className={`px-3 py-1.5 rounded-full text-xs font-medium border transition-all ${
                            isSelected
                              ? 'bg-accent text-[#0E0E11] border-accent font-semibold shadow-md shadow-accent/20'
                              : 'bg-white/5 text-muted border-white/10 hover:border-white/20 hover:text-main'
                          }`}
                        >
                          {c.name}
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            )}

            {/* Payment Mode Chips */}
            <div>
              <label className="block text-xs uppercase tracking-wider text-muted mb-2 font-mono">
                Payment Mode
              </label>
              <div className="flex gap-2">
                {PAYMENT_MODES.map((mode) => {
                  const isSelected = paymentMode === mode.id;
                  return (
                    <button
                      key={mode.id}
                      type="button"
                      onClick={() => setPaymentMode(mode.id)}
                      className={`flex-1 py-2 rounded-xl text-xs font-medium border transition-all ${
                        isSelected
                          ? 'bg-white/15 text-main border-accent/60 font-semibold shadow-sm'
                          : 'bg-white/5 text-muted border-white/10 hover:border-white/20'
                      }`}
                    >
                      {mode.label}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Note and Date */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label htmlFor="tx-note" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                  Note (optional)
                </label>
                <input
                  id="tx-note"
                  type="text"
                  maxLength={100}
                  placeholder="e.g. Lunch with team"
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-xs text-main placeholder:text-muted/60 outline-none focus:border-accent/60"
                />
              </div>

              <div>
                <label htmlFor="tx-date" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                  Date
                </label>
                <input
                  id="tx-date"
                  type="date"
                  value={occurredOn}
                  onChange={(e) => setOccurredOn(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-xs text-main outline-none focus:border-accent/60 font-mono"
                  required
                />
              </div>
            </div>

            {/* Error Message */}
            {error && (
              <p className="text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-2.5 text-center">
                {error}
              </p>
            )}

            {/* Save Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 rounded-2xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.99] transition-all shadow-lg shadow-accent/25 disabled:opacity-50"
            >
              {isSubmitting ? 'Saving...' : `Save ${type === 'expense' ? 'Expense' : 'Income'}`}
            </button>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}

