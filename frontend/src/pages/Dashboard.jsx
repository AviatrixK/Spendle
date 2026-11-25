import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import GlassCard from '../components/GlassCard';
import ProgressBar from '../components/ProgressBar';
import Fab from '../components/Fab';
import MerchantLogo from '../components/MerchantLogo';
import QuickAddSheet from '../components/QuickAddSheet';
import BudgetModal from '../components/BudgetModal';
import { formatINR, monthKey, formatMonthName, prevMonthKey, nextMonthKey } from '../lib/format';
import { dashboardApi, budgetApi, transactionApi, authApi } from '../api';

const iconFor = (name = '') => {
  const value = name.toLowerCase();
  if (/food|dining|restaurant/.test(value)) return '🍜';
  if (/grocer|market/.test(value)) return '🛒';
  if (/transport|travel|fuel|cab/.test(value)) return '🚕';
  if (/bill|utilities|rent/.test(value)) return '⚡';
  if (/entertain|movie|subscription/.test(value)) return '🎬';
  if (/shop/.test(value)) return '🛍';
  if (/health|medical/.test(value)) return '✚';
  if (/education|study/.test(value)) return '📚';
  return '◌';
};

function AreaChart({ transactions, type }) {
  const dayCount = Math.max(1, new Date(Number(transactions.month?.slice(0, 4)), Number(transactions.month?.slice(5, 7)) || 1, 0).getDate());
  const values = Array.from({ length: dayCount }, () => 0);
  transactions.items.forEach((tx) => {
    if (tx.type !== type) return;
    const day = Number(tx.occurred_on?.slice(-2)) - 1;
    if (day >= 0 && day < dayCount) values[day] += Number(tx.amount_paise || 0);
  });
  let running = 0;
  const cumulative = values.map((value) => (running += value));
  const max = Math.max(...cumulative, 1);
  const points = cumulative.map((value, i) => `${(i / Math.max(1, dayCount - 1)) * 100},${48 - (value / max) * 35}`).join(' ');
  const last = cumulative.at(-1) || 0;
  return (
    <div className="balance-chart" aria-label={`${type === 'income' ? 'Income' : 'Expense'} trend this month`}>
      <svg viewBox="0 0 100 54" preserveAspectRatio="none" role="img" aria-hidden="true">
        <defs><linearGradient id="balance-fill" x1="0" x2="0" y1="0" y2="1"><stop offset="0" stopColor="#F2C76A" stopOpacity=".28" /><stop offset="1" stopColor="#F2C76A" stopOpacity="0" /></linearGradient></defs>
        <polygon points={`0,54 ${points} 100,54`} fill="url(#balance-fill)" />
        <polyline points={points} fill="none" stroke="#F2C76A" strokeWidth="1.2" vectorEffect="non-scaling-stroke" />
      </svg>
      <span className="chart-tooltip">{formatINR(last)}</span>
      <div className="chart-axis"><span>1 {formatMonthName(transactions.month).split(' ')[0]}</span><span>{dayCount} {formatMonthName(transactions.month).split(' ')[0]}</span></div>
    </div>
  );
}

function GoalRing({ percent = 0 }) {
  const progress = Math.max(0, Math.min(100, Number(percent)));
  return <span className="goal-ring" style={{ '--progress': `${progress * 3.6}deg` }}><span>{progress}%</span></span>;
}

export default function Dashboard() {
  const [selectedMonth, setSelectedMonth] = useState(() => monthKey());
  const [chartType, setChartType] = useState('expense');
  const [isQuickAddOpen, setIsQuickAddOpen] = useState(false);
  const [quickAddType, setQuickAddType] = useState('expense');
  const [isBudgetModalOpen, setIsBudgetModalOpen] = useState(false);
  const [isAccountMenuOpen, setIsAccountMenuOpen] = useState(false);
  const accountMenuRef = useRef(null);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const location = useLocation();
  const currentUser = queryClient.getQueryData(['auth', 'me']);

  useEffect(() => {
    const action = new URLSearchParams(location.search).get('quickAdd');
    if (action === 'expense' || action === 'income') {
      setQuickAddType(action);
      setIsQuickAddOpen(true);
      navigate('/', { replace: true });
    }
  }, [location.search, navigate]);

  useEffect(() => {
    if (!isAccountMenuOpen) return undefined;
    function handlePointerDown(event) {
      if (!accountMenuRef.current?.contains(event.target)) setIsAccountMenuOpen(false);
    }
    function handleKeyDown(event) {
      if (event.key === 'Escape') setIsAccountMenuOpen(false);
    }
    document.addEventListener('pointerdown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('pointerdown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isAccountMenuOpen]);
  const { data: dashboardData, isLoading: isDashboardLoading, error: dashboardError, refetch } = useQuery({
    queryKey: ['dashboard', selectedMonth], queryFn: () => dashboardApi.get(selectedMonth),
  });
  const { data: previousData } = useQuery({
    queryKey: ['dashboard', prevMonthKey(selectedMonth)], queryFn: () => dashboardApi.get(prevMonthKey(selectedMonth)),
    retry: false,
  });
  const { data: categories = [] } = useQuery({ queryKey: ['categories'], queryFn: budgetApi.getCategories });
  const { data: transactions = [] } = useQuery({
    queryKey: ['transactions', selectedMonth, 200], queryFn: () => transactionApi.list(selectedMonth, 200),
  });
  const deleteTxMutation = useMutation({
    mutationFn: (id) => transactionApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboard', selectedMonth] });
      queryClient.invalidateQueries({ queryKey: ['transactions', selectedMonth] });
    },
  });

  async function handleLogout() {
    try { await authApi.logout(); } catch { /* session is cleared locally below */ }
    queryClient.clear();
    navigate('/login', { replace: true });
  }
  function refreshAllData() {
    ['dashboard', 'transactions', 'categories', 'budgets', 'goals'].forEach((key) => queryClient.invalidateQueries({ queryKey: [key] }));
  }

  const summary = dashboardData?.summary || { income_paise: 0, expense_paise: 0, net_paise: 0 };
  const budgets = dashboardData?.budgets || { total_limit_paise: 0, total_spent_paise: 0, total_left_paise: 0, items: [] };
  const goals = dashboardData?.goals || [];
  const status = dashboardData?.status || 'break_even';
  const overBudget = budgets.items.filter((item) => item.left_paise < 0);
  const budgetPct = budgets.total_limit_paise > 0 ? Math.round(budgets.total_spent_paise * 100 / budgets.total_limit_paise) : 0;
  const showAlert = overBudget.length > 0 || (budgets.total_limit_paise > 0 && budgetPct >= 80);
  const trend = previousData?.summary?.expense_paise > 0
    ? Math.round((summary.expense_paise - previousData.summary.expense_paise) * 100 / previousData.summary.expense_paise)
    : null;
  const chartData = useMemo(() => ({ month: selectedMonth, items: transactions }), [selectedMonth, transactions]);
  const recentGroups = useMemo(() => {
    const groups = new Map();
    transactions.slice(0, 10).forEach((tx) => {
      if (!groups.has(tx.occurred_on)) groups.set(tx.occurred_on, []);
      groups.get(tx.occurred_on).push(tx);
    });
    return [...groups.entries()];
  }, [transactions]);

  return (
    <main className="dashboard-page">
      <div className="dashboard-wrap">
        <header className="dashboard-topbar">
          <Link to="/" className="brand-wordmark" aria-label="Spendle home">Spendle<span /></Link>
          <div className="month-control" aria-label="Choose month">
            <button type="button" aria-label="Previous month" onClick={() => setSelectedMonth(prevMonthKey(selectedMonth))}>‹</button>
            <span>{formatMonthName(selectedMonth)}</span>
            <button type="button" aria-label="Next month" onClick={() => setSelectedMonth(nextMonthKey(selectedMonth))}>›</button>
          </div>
          <div className="topbar-actions">
            <div className="account-menu-wrap" ref={accountMenuRef}>
              <button type="button" className="avatar-chip" title="Account options" aria-label="Account options" aria-haspopup="menu" aria-expanded={isAccountMenuOpen} onClick={() => setIsAccountMenuOpen((open) => !open)}>
                {(currentUser?.name || currentUser?.email || 'S').trim().charAt(0).toUpperCase()}
              </button>
              {isAccountMenuOpen && <div className="account-menu" role="menu" aria-label="Account options">
                <div className="account-menu-heading"><strong>Your account</strong><span>{currentUser?.email || 'Signed in to Spendle'}</span></div>
                <Link role="menuitem" to="/goals" onClick={() => setIsAccountMenuOpen(false)}>Savings goals <span aria-hidden="true">→</span></Link>
                <button role="menuitem" type="button" onClick={handleLogout}>Log out</button>
              </div>}
            </div>
            <button type="button" className="calendar-button" aria-label="Go to current month" onClick={() => setSelectedMonth(monthKey())}>
              <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3.5" y="5" width="17" height="16" rx="3" /><path d="M7.5 3.5v3M16.5 3.5v3M4 9.5h16" /></svg>
            </button>
          </div>
        </header>

        {isDashboardLoading ? (
          <section className="dashboard-skeleton" aria-label="Loading dashboard"><div /><div /><div /><div /></section>
        ) : dashboardError ? (
          <section className="dashboard-error" role="alert"><p>Couldn’t load this month’s numbers.</p><button onClick={() => refetch()}>Try again</button></section>
        ) : (
          <>
            <motion.section className="balance-panel" initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
              <div className="balance-panel-top">
                <div><span className="eyebrow">{selectedMonth === monthKey() ? 'YOUR MONTH AT A GLANCE' : formatMonthName(selectedMonth).toUpperCase()}</span><h1>Kitna bacha?</h1></div>
                <span className={`status-chip status-${status}`}>{status === 'profit' ? 'In profit' : status === 'loss' ? 'In loss' : 'Break-even'}</span>
              </div>
              <p className={`balance-number ${budgets.total_left_paise < 0 ? 'is-over' : ''}`}>
                {formatINR(budgets.total_limit_paise ? budgets.total_left_paise : summary.net_paise)}
              </p>
              <div className="balance-meta">
                <span>{budgets.total_limit_paise ? 'left in your budget' : 'income after expenses'}</span>
                {trend !== null && <span className={`trend-chip ${trend > 0 ? 'trend-up' : 'trend-down'}`} aria-label={`${Math.abs(trend)} percent change vs last month`}>{trend > 0 ? '↑' : '↓'} {Math.abs(trend)}% vs last month</span>}
              </div>
              <div className="balance-segment" role="group" aria-label="Chart measure">
                <button type="button" className={chartType === 'income' ? 'selected' : ''} onClick={() => setChartType('income')}>Income</button>
                <button type="button" className={chartType === 'expense' ? 'selected' : ''} onClick={() => setChartType('expense')}>Expenses</button>
              </div>
              <AreaChart transactions={chartData} type={chartType} />
            </motion.section>

            {showAlert && <aside className="budget-alert" role="status"><span className="alert-icon">!</span><p>{overBudget.length ? `${overBudget[0].category_name} is over by ${formatINR(Math.abs(overBudget[0].left_paise))}` : `You’ve used ${budgetPct}% of your monthly budget`}</p><button type="button" onClick={() => setIsBudgetModalOpen(true)}>Review</button></aside>}

            <section className="budget-card">
              <div className="section-heading"><div><span className="eyebrow">THIS MONTH</span><h2>Budget overview</h2></div><button type="button" className="text-action" onClick={() => setIsBudgetModalOpen(true)}>{budgets.items.length ? 'Edit' : 'Set budget'} <span aria-hidden="true">↗</span></button></div>
              {budgets.items.length === 0 ? <div className="inline-empty"><p>Set a budget to see exactly what’s left.</p><button type="button" onClick={() => setIsBudgetModalOpen(true)}>Set your first budget</button></div> : <>
                <div className="budget-total"><div><span className="eyebrow">AVAILABLE</span><strong className={budgets.total_left_paise < 0 ? 'is-over' : ''}>{formatINR(budgets.total_left_paise)}</strong></div><div className="budget-spent"><span>spent</span><b>{formatINR(budgets.total_spent_paise)}</b><span>of {formatINR(budgets.total_limit_paise)}</span></div></div>
                <ProgressBar percent={budgetPct} isOverspent={budgets.total_left_paise < 0} height="h-2" />
                <div className="segment-bar" role="img" aria-label="Budget split by category">{budgets.items.map((item, i) => <span key={item.category_id} style={{ width: `${Math.max(2, (item.limit_paise / Math.max(1, budgets.total_limit_paise)) * 100)}%`, background: ['#e9bd61', '#74d6a6', '#e28a7d', '#a99cd7', '#80b9c8'][i % 5] }} />)}</div>
                <div className="budget-legend">{budgets.items.slice(0, 4).map((item, i) => <span key={item.category_id}><i style={{ background: ['#e9bd61', '#74d6a6', '#e28a7d', '#a99cd7'][i % 4] }} />{item.category_name}</span>)}</div>
              </>}
            </section>

            <section className="categories-section">
              <div className="section-heading"><div><span className="eyebrow">SPENDING PLAN</span><h2>Categories</h2></div><button type="button" className="text-action" onClick={() => setIsBudgetModalOpen(true)}>View all <span aria-hidden="true">→</span></button></div>
              {budgets.items.length ? <div className="category-grid">{budgets.items.slice(0, 6).map((item, i) => {
                const percent = item.limit_paise ? item.spent_paise * 100 / item.limit_paise : 0;
                const over = item.left_paise < 0;
                return <button type="button" key={item.category_id} className="category-card" onClick={() => setIsBudgetModalOpen(true)}>
                  <span className="category-top"><span className="category-icon" style={{ '--icon-tint': ['#f4dca8', '#b8e8d3', '#f1c6c0', '#d3c9f0'][i % 4] }}>{iconFor(item.category_name)}</span><span className={`category-status ${over ? 'is-over' : ''}`}>{over ? 'Over' : `${Math.round(percent)}%`}</span></span>
                  <span className="category-name">{item.category_name}</span><strong>{formatINR(item.spent_paise)}</strong>
                  <ProgressBar percent={percent} isOverspent={over} height="h-1.5" />
                  <span className={`category-left ${over ? 'is-over' : ''}`}>{over ? `Over by ${formatINR(Math.abs(item.left_paise))}` : `${formatINR(item.left_paise)} left`}</span>
                </button>;
              })}</div> : <div className="category-empty"><span>◌</span><p>Category budgets will show up here.</p><button onClick={() => setIsBudgetModalOpen(true)}>Add a budget</button></div>}
            </section>

            <section className="goals-section">
              <div className="section-heading"><div><span className="eyebrow">SAVING FOR SOMETHING</span><h2>Goals</h2></div><Link className="text-action" to="/goals">View all <span aria-hidden="true">→</span></Link></div>
              {goals.length ? <div className="goal-scroller">{goals.slice(0, 6).map((goal) => <Link to="/goals" className="goal-card" key={goal.id}><GoalRing percent={goal.progress_percent} /><span className="goal-copy"><strong>{goal.title}</strong><span>{formatINR(goal.saved_paise)} saved</span><small>of {formatINR(goal.target_paise)}</small></span><span className="goal-arrow">↗</span></Link>)}</div> : <div className="goal-empty"><p>Give your next big thing a place to grow.</p><Link to="/goals">Create a goal <span>→</span></Link></div>}
            </section>

            <section className="transactions-section">
              <div className="section-heading"><div><span className="eyebrow">THE LATEST</span><h2>Recent transactions</h2></div><span className="count-chip">{transactions.slice(0, 10).length} this month</span></div>
              {transactions.length === 0 ? <div className="transaction-empty"><p>No spends logged for {formatMonthName(selectedMonth)}.</p><button type="button" onClick={() => { setQuickAddType('expense'); setIsQuickAddOpen(true); }}>Add your first spend</button></div> : <div className="transaction-list">{recentGroups.map(([date, entries]) => <div className="transaction-day" key={date}><h3>{date === new Date().toISOString().slice(0, 10) ? 'Today' : new Date(`${date}T12:00:00`).toLocaleDateString('en-IN', { weekday: 'long', day: 'numeric', month: 'short' })}</h3>{entries.map((tx) => {
                const isIncome = tx.type === 'income';
                return <div className="transaction-row" key={tx.id}>
                  <MerchantLogo name={tx.note} note={tx.note} category={tx.category_name || ''} />
                  <div className="transaction-info"><strong>{tx.note || tx.category_name || (isIncome ? 'Income' : 'Spend')}</strong><span>{tx.category_name || (isIncome ? 'Income' : 'Expense')}</span></div>
                  <div className="transaction-amount"><strong className={isIncome ? 'income-amount' : ''}>{isIncome ? '+' : '−'}{formatINR(tx.amount_paise)}</strong><span>{tx.payment_mode}</span></div>
                  <button type="button" className="delete-transaction" onClick={() => deleteTxMutation.mutate(tx.id)} disabled={deleteTxMutation.isPending} aria-label="Delete transaction">×</button>
                </div>;
              })}</div>)}</div>}
            </section>
          </>
        )}
      </div>
      <Fab onHome={() => window.scrollTo({ top: 0, behavior: 'smooth' })} onAddExpense={() => { setQuickAddType('expense'); setIsQuickAddOpen(true); }} onAddIncome={() => { setQuickAddType('income'); setIsQuickAddOpen(true); }} />
      <QuickAddSheet isOpen={isQuickAddOpen} onClose={() => setIsQuickAddOpen(false)} initialType={quickAddType} categories={categories} onSuccess={refreshAllData} />
      <BudgetModal isOpen={isBudgetModalOpen} onClose={() => setIsBudgetModalOpen(false)} month={selectedMonth} categories={categories} onSuccess={refreshAllData} />
    </main>
  );
}
