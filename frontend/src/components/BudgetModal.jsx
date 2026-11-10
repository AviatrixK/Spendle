import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { rupeesToPaise, formatINR } from '../lib/format';
import { budgetApi } from '../api';

export default function BudgetModal({
  isOpen,
  onClose,
  month,
  categories = [],
  onSuccess,
}) {
  const [isCreatingCategory, setIsCreatingCategory] = useState(false);
  const [categoryId, setCategoryId] = useState('');
  const [newCategoryName, setNewCategoryName] = useState('');
  const [limitRupees, setLimitRupees] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!isOpen) return null;

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');

    const limitPaise = rupeesToPaise(limitRupees);
    if (limitPaise < 0) {
      setError('Budget limit cannot be negative');
      return;
    }

    try {
      setIsSubmitting(true);
      let targetCategoryId = categoryId;

      // If creating a brand new category first
      if (isCreatingCategory) {
        const trimmedName = newCategoryName.trim();
        if (!trimmedName) {
          setError('Please enter a category name');
          setIsSubmitting(false);
          return;
        }
        const createdCat = await budgetApi.createCategory(trimmedName);
        targetCategoryId = createdCat.id;
      } else if (!targetCategoryId) {
        setError('Please select a category');
        setIsSubmitting(false);
        return;
      }

      // Save or update budget for the active month
      await budgetApi.setBudget(targetCategoryId, month, limitPaise);

      if (onSuccess) onSuccess();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to save budget');
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        {/* Backdrop */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm"
        />

        {/* Modal */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{ duration: 0.2 }}
          className="relative w-full max-w-md bg-[#16161B] border border-white/10 rounded-2xl p-6 shadow-2xl z-10"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-serif text-main">Set Monthly Budget</h2>
            <button
              type="button"
              onClick={onClose}
              className="text-muted hover:text-main p-1 rounded-lg"
              aria-label="Close"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <p className="text-xs text-muted mb-4 font-mono">
            Configuring budget limit for <span className="text-accent">{month}</span>
          </p>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Category selection mode toggle */}
            <div className="flex items-center justify-between text-xs mb-1">
              <label className="uppercase tracking-wider text-muted font-mono">
                Category
              </label>
              <button
                type="button"
                onClick={() => {
                  setIsCreatingCategory(!isCreatingCategory);
                  setError('');
                }}
                className="text-accent hover:underline"
              >
                {isCreatingCategory ? '← Choose existing' : '+ Create new category'}
              </button>
            </div>

            {isCreatingCategory ? (
              <div>
                <input
                  type="text"
                  maxLength={40}
                  placeholder="e.g. Groceries, Entertainment"
                  value={newCategoryName}
                  onChange={(e) => setNewCategoryName(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent"
                  required
                />
              </div>
            ) : (
              <div>
                <select
                  value={categoryId}
                  onChange={(e) => setCategoryId(e.target.value)}
                  className="w-full bg-[#1C1C24] border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main outline-none focus:border-accent"
                  required
                >
                  <option value="">-- Select a Category --</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Monthly Limit in Rupees */}
            <div>
              <label htmlFor="budget-limit" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                Monthly Limit (₹)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted font-serif">₹</span>
                <input
                  id="budget-limit"
                  type="number"
                  step="any"
                  min="0"
                  placeholder="e.g. 5000"
                  value={limitRupees}
                  onChange={(e) => setLimitRupees(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl pl-8 pr-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                  required
                />
              </div>
            </div>

            {error && (
              <p className="text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-2.5">
                {error}
              </p>
            )}

            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 py-2.5 rounded-xl border border-white/10 text-muted hover:text-main text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="flex-1 py-2.5 rounded-xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.98] transition-all disabled:opacity-50"
              >
                {isSubmitting ? 'Saving...' : 'Set Budget'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}

