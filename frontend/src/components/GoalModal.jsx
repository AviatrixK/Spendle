import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { rupeesToPaise, paiseToRupees } from '../lib/format';
import { goalApi } from '../api';

export default function GoalModal({
  isOpen,
  onClose,
  goal = null, // null for create, object for edit
  onSuccess,
}) {
  const isEditing = Boolean(goal);

  const [title, setTitle] = useState('');
  const [targetRupees, setTargetRupees] = useState('');
  const [timelineType, setTimelineType] = useState('deadline'); // 'deadline' | 'monthly' | 'none'
  const [deadline, setDeadline] = useState('');
  const [monthlyRupees, setMonthlyRupees] = useState('');
  const [priority, setPriority] = useState('');
  const [status, setStatus] = useState('active');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setError('');
      if (goal) {
        setTitle(goal.title || '');
        setTargetRupees(paiseToRupees(goal.target_paise));
        setPriority(goal.priority ? String(goal.priority) : '1');
        setStatus(goal.status || 'active');

        if (goal.deadline) {
          setTimelineType('deadline');
          setDeadline(goal.deadline);
          setMonthlyRupees('');
        } else if (goal.monthly_amount_paise) {
          setTimelineType('monthly');
          setMonthlyRupees(paiseToRupees(goal.monthly_amount_paise));
          setDeadline('');
        } else {
          setTimelineType('none');
          setDeadline('');
          setMonthlyRupees('');
        }
      } else {
        setTitle('');
        setTargetRupees('');
        setTimelineType('deadline');
        setDeadline('');
        setMonthlyRupees('');
        setPriority('');
        setStatus('active');
      }
    }
  }, [isOpen, goal]);

  if (!isOpen) return null;

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');

    const targetPaise = rupeesToPaise(targetRupees);
    if (targetPaise <= 0) {
      setError('Target amount must be greater than ₹0');
      return;
    }

    let deadlineVal = null;
    let monthlyAmountPaise = null;

    if (timelineType === 'deadline') {
      if (!deadline) {
        setError('Please pick a deadline date');
        return;
      }
      deadlineVal = deadline;
    } else if (timelineType === 'monthly') {
      const parsedMonthly = rupeesToPaise(monthlyRupees);
      if (parsedMonthly <= 0) {
        setError('Monthly contribution amount must be greater than ₹0');
        return;
      }
      monthlyAmountPaise = parsedMonthly;
    }

    const payload = {
      title: title.trim(),
      target_paise: targetPaise,
      deadline: deadlineVal,
      monthly_amount_paise: monthlyAmountPaise,
    };

    if (priority && parseInt(priority, 10) > 0) {
      payload.priority = parseInt(priority, 10);
    }

    try {
      setIsSubmitting(true);
      if (isEditing) {
        payload.priority = payload.priority || goal.priority || 1;
        payload.status = status;
        await goalApi.update(goal.id, payload);
      } else {
        await goalApi.create(payload);
      }

      if (onSuccess) onSuccess();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to save savings goal');
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

        {/* Modal Content */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{ duration: 0.2 }}
          className="relative w-full max-w-md bg-[#16161B] border border-white/10 rounded-2xl p-6 shadow-2xl z-10 max-h-[90vh] overflow-y-auto"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-serif text-main">
              {isEditing ? 'Edit Savings Goal' : 'New Savings Goal'}
            </h2>
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

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Title */}
            <div>
              <label htmlFor="goal-title" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                Goal Title
              </label>
              <input
                id="goal-title"
                type="text"
                maxLength={60}
                placeholder="e.g. MacBook Pro, Trip to Goa"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent"
                required
              />
            </div>

            {/* Target in Rupees */}
            <div>
              <label htmlFor="goal-target" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                Target Amount (₹)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted font-serif">₹</span>
                <input
                  id="goal-target"
                  type="number"
                  step="any"
                  min="1"
                  placeholder="e.g. 90000"
                  value={targetRupees}
                  onChange={(e) => setTargetRupees(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl pl-8 pr-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                  required
                />
              </div>
            </div>

            {/* Timeline Strategy: Deadline OR Monthly Amount (mutually exclusive) */}
            <div className="bg-white/[0.02] border border-white/5 rounded-xl p-3.5 space-y-3">
              <span className="block text-xs uppercase tracking-wider text-muted font-mono">
                Timeline Plan (Choose At Most One)
              </span>

              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setTimelineType('deadline');
                    setMonthlyRupees('');
                  }}
                  className={`py-2 px-3 rounded-lg text-xs font-medium border text-center transition-all ${
                    timelineType === 'deadline'
                      ? 'bg-accent/15 border-accent text-accent font-semibold'
                      : 'bg-white/5 border-white/10 text-muted hover:text-main'
                  }`}
                >
                  By Target Deadline
                </button>

                <button
                  type="button"
                  onClick={() => {
                    setTimelineType('monthly');
                    setDeadline('');
                  }}
                  className={`py-2 px-3 rounded-lg text-xs font-medium border text-center transition-all ${
                    timelineType === 'monthly'
                      ? 'bg-accent/15 border-accent text-accent font-semibold'
                      : 'bg-white/5 border-white/10 text-muted hover:text-main'
                  }`}
                >
                  By Monthly Target
                </button>
              </div>

              {timelineType === 'deadline' && (
                <div>
                  <label htmlFor="goal-deadline" className="block text-xs text-muted mb-1">
                    Target Completion Date
                  </label>
                  <input
                    id="goal-deadline"
                    type="date"
                    value={deadline}
                    onChange={(e) => setDeadline(e.target.value)}
                    className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-xs text-main outline-none focus:border-accent font-mono"
                    required
                  />
                </div>
              )}

              {timelineType === 'monthly' && (
                <div>
                  <label htmlFor="goal-monthly" className="block text-xs text-muted mb-1">
                    Planned Monthly Contribution (₹)
                  </label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted font-serif text-xs">₹</span>
                    <input
                      id="goal-monthly"
                      type="number"
                      step="any"
                      min="1"
                      placeholder="e.g. 5000"
                      value={monthlyRupees}
                      onChange={(e) => setMonthlyRupees(e.target.value)}
                      className="w-full bg-white/5 border border-white/10 rounded-xl pl-7 pr-3 py-2 text-xs text-main outline-none focus:border-accent font-mono"
                      required
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Priority & Status */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label htmlFor="goal-priority" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                  Priority Rank
                </label>
                <input
                  id="goal-priority"
                  type="number"
                  min="1"
                  placeholder={isEditing ? '1' : 'Auto (Next)'}
                  value={priority}
                  onChange={(e) => setPriority(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                />
                <span className="text-[10px] text-muted block mt-0.5">1 = highest priority</span>
              </div>

              {isEditing && (
                <div>
                  <label htmlFor="goal-status" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                    Status
                  </label>
                  <select
                    id="goal-status"
                    value={status}
                    onChange={(e) => setStatus(e.target.value)}
                    className="w-full bg-[#1C1C24] border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main outline-none focus:border-accent"
                  >
                    <option value="active">Active</option>
                    <option value="paused">Paused</option>
                    <option value="completed">Completed</option>
                  </select>
                </div>
              )}
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
                {isSubmitting ? 'Saving...' : isEditing ? 'Update Goal' : 'Create Goal'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}

