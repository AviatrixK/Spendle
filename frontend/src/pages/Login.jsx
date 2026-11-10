import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import GlassCard from '../components/GlassCard';
import { authApi } from '../api';

export default function Login() {
  const [isSignup, setIsSignup] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const authMutation = useMutation({
    mutationFn: async () => {
      if (isSignup) {
        return authApi.signup(name, email, password);
      }
      return authApi.login(email, password);
    },
    onSuccess: (data) => {
      // Invalidate and pre-populate auth query
      queryClient.setQueryData(['auth', 'me'], data);
      navigate('/', { replace: true });
    },
    onError: (err) => {
      setErrorMessage(err.message || 'Authentication failed');
    },
  });

  function handleSubmit(e) {
    e.preventDefault();
    setErrorMessage('');
    authMutation.mutate();
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 15 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25 }}
        className="w-full max-w-md"
      >
        {/* Brand Wordmark */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-serif tracking-tight text-main inline-flex items-baseline">
            Spendle<span className="text-accent text-3xl font-sans ml-0.5">.</span>
          </h1>
          <p className="text-xs text-muted tracking-widest uppercase mt-2 font-mono">
            Personal Budget & Financial Discipline
          </p>
        </div>

        <GlassCard className="p-8 border-white/10">
          {/* Mode Switcher */}
          <div className="flex bg-white/5 p-1 rounded-xl mb-6 border border-white/10">
            <button
              type="button"
              onClick={() => {
                setIsSignup(false);
                setErrorMessage('');
              }}
              className={`flex-1 py-2 text-xs font-medium rounded-lg transition-all ${
                !isSignup
                  ? 'bg-accent text-[#0E0E11] font-semibold shadow-sm'
                  : 'text-muted hover:text-main'
              }`}
            >
              Sign In
            </button>
            <button
              type="button"
              onClick={() => {
                setIsSignup(true);
                setErrorMessage('');
              }}
              className={`flex-1 py-2 text-xs font-medium rounded-lg transition-all ${
                isSignup
                  ? 'bg-accent text-[#0E0E11] font-semibold shadow-sm'
                  : 'text-muted hover:text-main'
              }`}
            >
              Create Account
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {isSignup && (
              <div>
                <label htmlFor="login-name" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                  Full Name
                </label>
                <input
                  id="login-name"
                  type="text"
                  placeholder="e.g. Mayank Rawat"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent"
                  required
                />
              </div>
            )}

            <div>
              <label htmlFor="login-email" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                Email Address
              </label>
              <input
                id="login-email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent"
                required
              />
            </div>

            <div>
              <label htmlFor="login-password" className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                Password
              </label>
              <input
                id="login-password"
                type="password"
                placeholder="At least 8 characters"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                required
              />
            </div>

            {errorMessage && (
              <motion.p
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                className="text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-3 text-center"
              >
                {errorMessage}
              </motion.p>
            )}

            <button
              type="submit"
              disabled={authMutation.isPending}
              className="w-full py-3 rounded-xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.99] transition-all shadow-lg shadow-accent/20 disabled:opacity-50 mt-2"
            >
              {authMutation.isPending
                ? 'Authenticating...'
                : isSignup
                ? 'Create Spendle Account'
                : 'Sign In'}
            </button>
          </form>
        </GlassCard>
      </motion.div>
    </div>
  );
}

