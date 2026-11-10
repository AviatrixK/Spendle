import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import GlassCard from '../components/GlassCard';
import ProgressBar from '../components/ProgressBar';
import GoalModal from '../components/GoalModal';
import { formatINR, prevMonthKey, monthKey, formatMonthName } from '../lib/format';
import { goalApi } from '../api';

export default function Goals() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingGoal, setEditingGoal] = useState(null);
  const [closeMonthInput, setCloseMonthInput] = useState(() => prevMonthKey(monthKey()));
  const [closeResultMsg, setCloseResultMsg] = useState('');
  const [closeErrorMsg, setCloseErrorMsg] = useState('');

  const queryClient = useQueryClient();

  // Load Goals list (ordered by priority ASC, created_at ASC)
  const { data: goals = [], isLoading, error } = useQuery({
    queryKey: ['goals'],
    queryFn: goalApi.list,
  });

  // Delete goal mutation
  const deleteMutation = useMutation({
    mutationFn: (id) => goalApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  // Toggle status (active <-> paused) mutation
  const toggleStatusMutation = useMutation({
    mutationFn: async ({ id, currentGoal }) => {
      const nextStatus = currentGoal.status === 'active' ? 'paused' : 'active';
      return goalApi.update(id, {
        title: currentGoal.title,
        target_paise: currentGoal.target_paise,
        deadline: currentGoal.deadline,
        monthly_amount_paise: currentGoal.monthly_amount_paise,
        priority: currentGoal.priority,
        status: nextStatus,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  // Month-end sweep rollover mutation
  const closeMonthMutation = useMutation({
    mutationFn: (m) => goalApi.closeMonth(m),
    onSuccess: (data) => {
      setCloseErrorMsg('');
      if (data.already_closed) {
        setCloseResultMsg(`Month ${closeMonthInput} was already closed. No duplicate rollover applied.`);
      } else {
        setCloseResultMsg(
          `Rollover complete! Moved ${formatINR(data.moved_paise)} across ${data.goals_funded} goal${
            data.goals_funded === 1 ? '' : 's'
          }.`
        );
      }
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
    onError: (err) => {
      setCloseResultMsg('');
      setCloseErrorMsg(err.message || 'Failed to close month');
    },
  });

  function handleOpenAdd() {
    setEditingGoal(null);
    setIsModalOpen(true);
  }

  function handleOpenEdit(g) {
    setEditingGoal(g);
    setIsModalOpen(true);
  }

  function handleCloseMonthSubmit(e) {
    e.preventDefault();
    if (!closeMonthInput) return;
    setCloseResultMsg('');
    setCloseErrorMsg('');
    closeMonthMutation.mutate(closeMonthInput);
  }

  return (
    <div className="min-h-screen pb-16 pt-4 px-4">
      <div className="max-w-[720px] mx-auto space-y-6">
        {/* Top Header */}
        <header className="flex items-center justify-between py-2 border-b border-white/5">
          <div className="flex items-center gap-3">
            <Link
              to="/"
              className="text-xs text-muted hover:text-main px-2.5 py-1.5 rounded-lg bg-white/5 border border-white/10 transition-colors"
            >
              ← Dashboard
            </Link>
            <h1 className="font-serif text-2xl text-main tracking-tight">Savings Goals</h1>
          </div>

          <button
            type="button"
            onClick={handleOpenAdd}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl bg-accent text-[#0E0E11] text-xs font-semibold hover:bg-accent-hover transition-all active:scale-95 shadow-md shadow-accent/20"
          >
            <span className="text-sm font-bold leading-none">+</span>
            <span>New Goal</span>
          </button>
        </header>

        {/* Manual Month Close Panel (Viva Demo Shortcut) */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
        >
          <GlassCard className="p-4 border-accent/20 bg-accent/[0.02]">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono uppercase tracking-wider text-accent font-semibold">
                    Month-End Rollover Sweep
                  </span>
                  <span className="text-[10px] bg-accent/20 text-accent px-1.5 py-0.5 rounded font-mono">
                    Demo Control
                  </span>
                </div>
                <p className="text-[11px] text-muted mt-0.5">
                  Manual trigger to simulate the background worker sweeping unspent budget into priority goals.
                </p>
              </div>

              <form onSubmit={handleCloseMonthSubmit} className="flex items-center gap-2">
                <input
                  type="month"
                  value={closeMonthInput}
                  onChange={(e) => setCloseMonthInput(e.target.value)}
                  className="bg-white/5 border border-white/10 rounded-xl px-2.5 py-1.5 text-xs text-main font-mono outline-none focus:border-accent"
                  required
                />
                <button
                  type="submit"
                  disabled={closeMonthMutation.isPending}
                  className="px-3.5 py-1.5 rounded-xl bg-white/10 hover:bg-white/15 border border-white/15 text-main text-xs font-medium transition-all active:scale-95 disabled:opacity-50 whitespace-nowrap"
                >
                  {closeMonthMutation.isPending ? 'Sweeping...' : 'Close Month'}
                </button>
              </form>
            </div>

            {closeResultMsg && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                className="mt-3 text-xs text-semantic-profit bg-semantic-profit/10 border border-semantic-profit/20 rounded-xl p-2.5"
              >
                {closeResultMsg}
              </motion.div>
            )}

            {closeErrorMsg && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                className="mt-3 text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-2.5"
              >
                {closeErrorMsg}
              </motion.div>
            )}
          </GlassCard>
        </motion.div>

        {/* Goals List */}
        {isLoading ? (
          <div className="text-center py-16 text-muted text-sm animate-pulse font-mono">
            Loading savings goals...
          </div>
        ) : error ? (
          <GlassCard className="p-6 text-center text-semantic-loss text-sm">
            {error.message || 'Failed to load savings goals'}
          </GlassCard>
        ) : goals.length === 0 ? (
          <GlassCard className="text-center py-12 px-6">
            <h3 className="font-serif text-lg text-main mb-1">No savings goals yet</h3>
            <p className="text-xs text-muted mb-4 max-w-sm mx-auto">
              Set prioritized savings goals to allocate your monthly leftover budget automatically.
            </p>
            <button
              type="button"
              onClick={handleOpenAdd}
              className="px-4 py-2 rounded-xl bg-accent text-[#0E0E11] text-xs font-semibold hover:bg-accent-hover transition-all"
            >
              + Create First Goal
            </button>
          </GlassCard>
        ) : (
          <div className="space-y-4">
            {goals.map((goal, idx) => {
              const isCompleted = goal.status === 'completed';
              const isPaused = goal.status === 'paused';

              return (
                <motion.div
                  key={goal.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.2, delay: idx * 0.04 }}
                >
                  <GlassCard className="p-5 relative space-y-4 border-white/10 hover:border-white/20 transition-all">
                    {/* Header Row */}
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-mono text-xs px-2 py-0.5 rounded-full bg-accent/15 text-accent border border-accent/30 font-semibold">
                            Rank #{goal.priority}
                          </span>

                          <span
                            className={`text-[10px] uppercase tracking-wider font-mono px-2 py-0.5 rounded-full border ${
                              isCompleted
                                ? 'bg-semantic-profit/15 text-semantic-profit border-semantic-profit/30'
                                : isPaused
                                ? 'bg-white/5 text-muted border-white/10'
                                : 'bg-white/10 text-main border-white/15'
                            }`}
                          >
                            {goal.status}
                          </span>
                        </div>

                        <h2 className={`font-serif text-xl tracking-tight ${isCompleted ? 'text-muted line-through' : 'text-main'}`}>
                          {goal.title}
                        </h2>
                      </div>

                      {/* Card Action Controls */}
                      <div className="flex items-center gap-1.5">
                        <button
                          type="button"
                          onClick={() => handleOpenEdit(goal)}
                          className="px-2.5 py-1 rounded-lg text-xs font-medium text-muted hover:text-main hover:bg-white/5 border border-transparent hover:border-white/10 transition-colors"
                        >
                          Edit
                        </button>

                        {!isCompleted && (
                          <button
                            type="button"
                            onClick={() => toggleStatusMutation.mutate({ id: goal.id, currentGoal: goal })}
                            className="px-2.5 py-1 rounded-lg text-xs font-medium text-muted hover:text-main hover:bg-white/5 border border-transparent hover:border-white/10 transition-colors"
                          >
                            {isPaused ? 'Resume' : 'Pause'}
                          </button>
                        )}

                        <button
                          type="button"
                          onClick={() => deleteMutation.mutate(goal.id)}
                          className="p-1 rounded-lg text-muted/60 hover:text-semantic-loss hover:bg-white/5 transition-colors"
                          title="Delete goal"
                          aria-label="Delete goal"
                        >
                          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
                    </div>

                    {/* Progress Bar */}
                    <div className="space-y-1.5">
                      <div className="flex items-baseline justify-between text-xs font-mono">
                        <span className="text-main font-semibold">
                          {formatINR(goal.saved_paise)}{' '}
                          <span className="text-muted font-normal">of {formatINR(goal.target_paise)}</span>
                        </span>
                        <span className="text-accent font-semibold">
                          {goal.progress_percent}%
                        </span>
                      </div>
                      <ProgressBar percent={goal.progress_percent} height="h-2" />
                    </div>

                    {/* Computed Timeline Metrics */}
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 pt-2 border-t border-white/5 text-xs font-mono">
                      <div>
                        <span className="text-muted block text-[10px] uppercase">Remaining</span>
                        <span className="text-main font-medium">
                          {formatINR(goal.remaining_paise)}
                        </span>
                      </div>

                      {goal.deadline && (
                        <div>
                          <span className="text-muted block text-[10px] uppercase">Deadline</span>
                          <span className="text-main font-medium">{goal.deadline}</span>
                          {goal.monthly_needed_paise != null && (
                            <span className="text-accent block text-[11px]">
                              {formatINR(goal.monthly_needed_paise)}/mo
                            </span>
                          )}
                        </div>
                      )}

                      {goal.monthly_amount_paise && (
                        <div>
                          <span className="text-muted block text-[10px] uppercase">Monthly Target</span>
                          <span className="text-main font-medium">
                            {formatINR(goal.monthly_amount_paise)}/mo
                          </span>
                          {goal.eta_month && (
                            <span className="text-accent block text-[11px]">
                              ETA: {formatMonthName(goal.eta_month)}
                            </span>
                          )}
                        </div>
                      )}
                    </div>
                  </GlassCard>
                </motion.div>
              );
            })}
          </div>
        )}
      </div>

      {/* Goal Creation / Edit Modal */}
      <GoalModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        goal={editingGoal}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['goals'] });
          queryClient.invalidateQueries({ queryKey: ['dashboard'] });
        }}
      />
    </div>
  );
}

