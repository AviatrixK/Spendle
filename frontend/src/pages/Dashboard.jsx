import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import GlassCard from '../components/GlassCard';
import ProgressBar from '../components/ProgressBar';
import Fab from '../components/Fab';
import QuickAddSheet from '../components/QuickAddSheet';
import BudgetModal from '../components/BudgetModal';
import { formatINR, monthKey, formatMonthName, prevMonthKey, nextMonthKey } from '../lib/format';
import { dashboardApi, budgetApi, transactionApi, authApi } from '../api';

export default function Dashboard() {
  const [selectedMonth, setSelectedMonth] = useState(() => monthKey());
  const [isQuickAddOpen, setIsQuickAddOpen] = useState(false);
  const [quickAddType, setQuickAddType] = useState('expense');
  const [isBudgetModalOpen, setIsBudgetModalOpen] = useState(false);

  const queryClient = useQueryClient();
  const navigate = useNavigate();

  // Load Dashboard Aggregate data
  const {
    data: dashboardData,
    isLoading: isDashboardLoading,
    error: dashboardError,
  } = useQuery({
    queryKey: ['dashboard', selectedMonth],
    queryFn: () => dashboardApi.get(selectedMonth),
  });

  // Load Categories for Quick-Add chips and Budget modal
  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: budgetApi.getCategories,
  });

  // Load Recent Transactions (limit 10)
  const { data: transactions = [] } = useQuery({
    queryKey: ['transactions', selectedMonth],
    queryFn: () => transactionApi.list(selectedMonth, 10),
  });

  // Delete transaction mutation
  const deleteTxMutation = useMutation({
    mutationFn: (id) => transactionApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboard', selectedMonth] });
      queryClient.invalidateQueries({ queryKey: ['transactions', selectedMonth] });
    },
  });

  // Logout handler
  async function handleLogout() {
    try {
      await authApi.logout();
    } catch {
      // ignore
    } finally {
      queryClient.clear();
      navigate('/login', { replace: true });
    }
  }

  function refreshAllData() {
    queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    queryClient.invalidateQueries({ queryKey: ['transactions'] });
    queryClient.invalidateQueries({ queryKey: ['categories'] });
    queryClient.invalidateQueries({ queryKey: ['budgets'] });
    queryClient.invalidateQueries({ queryKey: ['goals'] });
  }

  const summary = dashboardData?.summary || { income_paise: 0, expense_paise: 0, net_paise: 0 };
  const budgets = dashboardData?.budgets || { total_limit_paise: 0, total_spent_paise: 0, total_left_paise: 0, items: [] };
  const goals = dashboardData?.goals || [];
  const topGoals = goals.slice(0, 3);
  const status = dashboardData?.status || 'break_even';

  return (
    <div className="min-h-screen pb-28 pt-4 px-4">
      <div className="max-w-[720px] mx-auto space-y-6">
        {/* Top Bar */}
        <header className="flex items-center justify-between py-2 border-b border-white/5">
          <div className="flex items-baseline gap-1">
            <span className="font-serif text-2xl text-main tracking-tight">Spendle</span>
            <span className="w-1.5 h-1.5 rounded-full bg-accent" />
          </div>

          {/* Month Switcher */}
          <div className="flex items-center gap-2 bg-white/5 px-3 py-1.5 rounded-full border border-white/10">
            <button
              type="button"
              onClick={() => setSelectedMonth(prevMonthKey(selectedMonth))}
              className="text-muted hover:text-main text-sm px-1.5 transition-colors"
              aria-label="Previous month"
            >
              ‹
            </button>
            <span className="font-serif text-sm tracking-wide text-main min-w-[76px] text-center">
              {formatMonthName(selectedMonth)}
            </span>
            <button
              type="button"
              onClick={() => setSelectedMonth(nextMonthKey(selectedMonth))}
              className="text-muted hover:text-main text-sm px-1.5 transition-colors"
              aria-label="Next month"
            >
              ›
            </button>
          </div>

          <div className="flex items-center gap-3">
            <Link
              to="/goals"
              className="text-xs font-medium text-muted hover:text-accent transition-colors"
            >
              Goals
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              className="text-xs text-muted hover:text-semantic-loss transition-colors"
            >
              Logout
            </button>
          </div>
        </header>

        {isDashboardLoading ? (
          <div className="text-center py-16 text-muted text-sm animate-pulse font-mono">
            Loading dashboard data...
          </div>
        ) : dashboardError ? (
          <GlassCard className="p-6 text-center text-semantic-loss text-sm">
            {dashboardError.message || 'Failed to load dashboard data'}
          </GlassCard>
        ) : (
          <>
            {/* Status Banner */}
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25 }}
            >
              <GlassCard className="text-center p-6 border-white/10 relative overflow-hidden">
                <div className="text-xs font-mono uppercase tracking-wider text-muted mb-1">
                  Monthly Balance
                </div>

                <div className="flex items-center justify-center gap-2 mb-2">
                  <span
                    className={`inline-block px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                      status === 'profit'
                        ? 'bg-semantic-profit/15 text-semantic-profit border border-semantic-profit/30'
                        : status === 'loss'
                        ? 'bg-semantic-loss/15 text-semantic-loss border border-semantic-loss/30'
                        : 'bg-white/10 text-muted border border-white/15'
                    }`}
                  >
                    {status === 'profit'
                      ? 'In profit'
                      : status === 'loss'
                      ? 'In loss'
                      : 'Break-even'}
                  </span>
                </div>

                <div
                  className={`text-4xl sm:text-5xl font-serif font-medium tracking-tight mb-4 ${
                    status === 'profit'
                      ? 'text-semantic-profit'
                      : status === 'loss'
                      ? 'text-semantic-loss'
                      : 'text-main'
                  }`}
                >
                  {summary.net_paise < 0 ? '-' : ''}
                  {formatINR(Math.abs(summary.net_paise))}
                </div>

                <div className="grid grid-cols-2 gap-4 pt-4 border-t border-white/5 font-mono text-xs">
                  <div>
                    <span className="text-muted block mb-0.5">Income</span>
                    <span className="text-semantic-profit font-medium">
                      +{formatINR(summary.income_paise)}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted block mb-0.5">Expenses</span>
                    <span className="text-main font-medium">
                      -{formatINR(summary.expense_paise)}
                    </span>
                  </div>
                </div>
              </GlassCard>
            </motion.div>

            {/* Budgets Section */}
            <motion.section
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25, delay: 0.05 }}
              className="space-y-3"
            >
              <div className="flex items-center justify-between px-1">
                <h2 className="font-serif text-lg text-main">Monthly Budgets</h2>
                <button
                  type="button"
                  onClick={() => setIsBudgetModalOpen(true)}
                  className="text-xs text-accent hover:text-accent-hover font-medium transition-colors"
                >
                  + Set Budget
                </button>
              </div>

              <GlassCard className="space-y-4">
                {budgets.items.length === 0 ? (
                  <div className="text-center py-6">
                    <p className="text-xs text-muted mb-3">
                      No budgets yet. Set one to see what's left.
                    </p>
                    <button
                      type="button"
                      onClick={() => setIsBudgetModalOpen(true)}
                      className="text-xs px-3 py-1.5 rounded-lg bg-accent/15 border border-accent/30 text-accent font-medium hover:bg-accent/25 transition-all"
                    >
                      Set first budget
                    </button>
                  </div>
                ) : (
                  <>
                    <div className="space-y-3.5 divide-y divide-white/5">
                      {budgets.items.map((item) => {
                        const isOver = item.left_paise < 0;
                        const pct = item.limit_paise > 0
                          ? Math.round((item.spent_paise * 100) / item.limit_paise)
                          : 0;

                        return (
                          <div key={item.category_id} className="pt-3 first:pt-0">
                            <div className="flex items-baseline justify-between text-xs mb-1.5">
                              <span className="font-medium text-main">{item.category_name}</span>
                              <span className="font-mono text-muted text-[11px]">
                                {formatINR(item.spent_paise)} of {formatINR(item.limit_paise)}
                              </span>
                            </div>

                            <ProgressBar
                              percent={pct}
                              isOverspent={isOver}
                              height="h-1.5"
                              className="mb-1.5"
                            />

                            <div className="flex items-center justify-between text-[11px] font-mono">
                              <span className="text-muted">{pct}% spent</span>
                              <span className={isOver ? 'text-semantic-loss font-semibold' : 'text-muted'}>
                                {isOver
                                  ? `${formatINR(Math.abs(item.left_paise))} over`
                                  : `${formatINR(item.left_paise)} left`}
                              </span>
                            </div>
                          </div>
                        );
                      })}
                    </div>

                    {/* Total Row */}
                    <div className="pt-3 border-t border-white/10 flex items-center justify-between text-xs font-mono">
                      <span className="text-muted uppercase tracking-wider text-[10px]">Total</span>
                      <div className="text-right">
                        <span className="text-muted mr-2">
                          {formatINR(budgets.total_spent_paise)} of {formatINR(budgets.total_limit_paise)}
                        </span>
                        <span className={budgets.total_left_paise < 0 ? 'text-semantic-loss font-semibold' : 'text-accent font-semibold'}>
                          • {budgets.total_left_paise < 0
                            ? `${formatINR(Math.abs(budgets.total_left_paise))} over`
                            : `${formatINR(budgets.total_left_paise)} left`}
                        </span>
                      </div>
                    </div>
                  </>
                )}
              </GlassCard>
            </motion.section>

            {/* Goals Preview */}
            <motion.section
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25, delay: 0.1 }}
              className="space-y-3"
            >
              <div className="flex items-center justify-between px-1">
                <h2 className="font-serif text-lg text-main">Savings Goals</h2>
                <Link
                  to="/goals"
                  className="text-xs text-accent hover:text-accent-hover font-medium transition-colors"
                >
                  View all ({goals.length}) →
                </Link>
              </div>

              <GlassCard className="space-y-3">
                {topGoals.length === 0 ? (
                  <div className="text-center py-6">
                    <p className="text-xs text-muted mb-2">No savings goals created yet.</p>
                    <Link
                      to="/goals"
                      className="text-xs text-accent hover:underline font-medium"
                    >
                      + Create your first savings goal
                    </Link>
                  </div>
                ) : (
                  topGoals.map((g) => (
                    <Link
                      key={g.id}
                      to="/goals"
                      className="block p-2.5 -mx-2.5 rounded-xl hover:bg-white/5 transition-colors"
                    >
                      <div className="flex items-baseline justify-between text-xs mb-1.5">
                        <span className="font-medium text-main flex items-center gap-1.5">
                          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-white/10 text-muted">
                            #{g.priority}
                          </span>
                          {g.title}
                        </span>
                        <span className="font-mono text-muted text-[11px]">
                          {g.progress_percent}% of {formatINR(g.target_paise)}
                        </span>
                      </div>
                      <ProgressBar percent={g.progress_percent} height="h-1.5" />
                    </Link>
                  ))
                )}
              </GlassCard>
            </motion.section>

            {/* Recent Transactions */}
            <motion.section
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25, delay: 0.15 }}
              className="space-y-3"
            >
              <div className="flex items-center justify-between px-1">
                <h2 className="font-serif text-lg text-main">Recent Transactions</h2>
                <span className="text-xs font-mono text-muted">
                  {transactions.length} this month
                </span>
              </div>

              <GlassCard className="p-0 overflow-hidden divide-y divide-white/5">
                {transactions.length === 0 ? (
                  <div className="text-center py-8 px-4 text-xs text-muted">
                    No transactions logged for {formatMonthName(selectedMonth)} yet.
                    <br />Tap <span className="text-accent font-semibold">+ Expense</span> or <span className="text-main font-semibold">+ Income</span> below to add one.
                  </div>
                ) : (
                  transactions.map((tx) => {
                    const isIncome = tx.type === 'income';

                    return (
                      <div
                        key={tx.id}
                        className="flex items-center justify-between p-3.5 hover:bg-white/[0.02] transition-colors"
                      >
                        <div className="flex-1 min-w-0 pr-3">
                          <div className="flex items-center gap-2 mb-0.5">
                            <span className="text-xs font-medium text-main truncate">
                              {tx.category_name || (isIncome ? 'General Income' : 'Expense')}
                            </span>
                            <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-white/5 text-muted border border-white/5">
                              {tx.payment_mode}
                            </span>
                          </div>
                          <div className="flex items-center gap-2 text-[11px] text-muted truncate">
                            <span className="font-mono">{tx.occurred_on}</span>
                            {tx.note && <span>• {tx.note}</span>}
                          </div>
                        </div>

                        <div className="flex items-center gap-3">
                          <span
                            className={`font-mono text-xs font-medium ${
                              isIncome ? 'text-semantic-profit' : 'text-main'
                            }`}
                          >
                            {isIncome ? '+' : '-'}{formatINR(tx.amount_paise)}
                          </span>

                          <button
                            type="button"
                            onClick={() => deleteTxMutation.mutate(tx.id)}
                            disabled={deleteTxMutation.isPending}
                            className="text-muted/60 hover:text-semantic-loss p-1 rounded transition-colors"
                            title="Delete transaction"
                            aria-label="Delete transaction"
                          >
                            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    );
                  })
                )}
              </GlassCard>
            </motion.section>
          </>
        )}
      </div>

      {/* Floating Action Buttons */}
      <Fab
        onAddExpense={() => {
          setQuickAddType('expense');
          setIsQuickAddOpen(true);
        }}
        onAddIncome={() => {
          setQuickAddType('income');
          setIsQuickAddOpen(true);
        }}
      />

      {/* Quick Add Bottom Sheet */}
      <QuickAddSheet
        isOpen={isQuickAddOpen}
        onClose={() => setIsQuickAddOpen(false)}
        initialType={quickAddType}
        categories={categories}
        onSuccess={refreshAllData}
      />

      {/* Budget Configuration Modal */}
      <BudgetModal
        isOpen={isBudgetModalOpen}
        onClose={() => setIsBudgetModalOpen(false)}
        month={selectedMonth}
        categories={categories}
        onSuccess={refreshAllData}
      />
    </div>
  );
}

