import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import GlassCard from '../components/GlassCard';
import EmailVerification from '../components/EmailVerification';
import ForgotPasswordModal from '../components/ForgotPasswordModal';
import { authApi } from '../api';

export default function Login() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Mode: 'signin', 'signup', or 'verify'
  const [mode, setMode] = useState('signin');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [isForgotOpen, setIsForgotOpen] = useState(false);

  // Parse URL search parameters on mount (e.g. OAuth error or direct verification link)
  useEffect(() => {
    const errorParam = searchParams.get('error');
    if (errorParam) {
      if (errorParam === 'google_auth_failed' || errorParam === 'oauth_token_error') {
        setErrorMessage('Google authentication could not be completed. Please try again or use email.');
      } else {
        setErrorMessage(decodeURIComponent(errorParam));
      }
    }

    const verifyParam = searchParams.get('verify');
    const emailParam = searchParams.get('email');
    if (verifyParam && emailParam) {
      setEmail(emailParam);
      setMode('verify');
    }
  }, [searchParams]);

  const loginMutation = useMutation({
    mutationFn: async () => {
      return authApi.login(email, password);
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['auth', 'me'], data);
      navigate('/', { replace: true });
    },
    onError: (err) => {
      if (err.data && err.data.error === 'email_unverified') {
        // Account exists but is unverified -> transition directly to OTP challenge
        setErrorMessage('');
        setMode('verify');
        return;
      }
      setErrorMessage(err.message || 'Invalid email or password');
    },
  });

  const signupMutation = useMutation({
    mutationFn: async () => {
      return authApi.signup(name, email, password);
    },
    onSuccess: () => {
      setErrorMessage('');
      // User registered -> move to 4-digit OTP email verification
      setMode('verify');
    },
    onError: (err) => {
      setErrorMessage(err.message || 'Registration failed. Please check the information provided.');
    },
  });

  function handleSubmit(e) {
    e.preventDefault();
    setErrorMessage('');

    if (mode === 'signup') {
      if (password.length < 8) {
        setErrorMessage('Password must be at least 8 characters long.');
        return;
      }
      if (password !== confirmPassword) {
        setErrorMessage('Passwords do not match.');
        return;
      }
      signupMutation.mutate();
    } else {
      loginMutation.mutate();
    }
  }

  function handleVerificationSuccess(data) {
    queryClient.setQueryData(['auth', 'me'], data);
    navigate('/', { replace: true });
  }

  function handleGoogleLogin() {
    window.location.href = authApi.googleLoginUrl;
  }

  const isSubmitting = loginMutation.isPending || signupMutation.isPending;

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
          {mode === 'verify' ? (
            <EmailVerification
              email={email}
              onSuccess={handleVerificationSuccess}
              onBack={() => {
                setMode('signin');
                setErrorMessage('');
              }}
            />
          ) : (
            <>
              {/* Mode Switcher */}
              <div className="flex bg-white/5 p-1 rounded-xl mb-6 border border-white/10">
                <button
                  type="button"
                  onClick={() => {
                    setMode('signin');
                    setErrorMessage('');
                  }}
                  className={`flex-1 py-2 text-xs font-medium rounded-lg transition-all ${
                    mode === 'signin'
                      ? 'bg-accent text-[#0E0E11] font-semibold shadow-sm'
                      : 'text-muted hover:text-main'
                  }`}
                >
                  Sign In
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setMode('signup');
                    setErrorMessage('');
                  }}
                  className={`flex-1 py-2 text-xs font-medium rounded-lg transition-all ${
                    mode === 'signup'
                      ? 'bg-accent text-[#0E0E11] font-semibold shadow-sm'
                      : 'text-muted hover:text-main'
                  }`}
                >
                  Create Account
                </button>
              </div>

              {/* Google Sign-In Button */}
              <button
                type="button"
                onClick={handleGoogleLogin}
                className="w-full py-2.5 px-4 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-main text-xs font-medium flex items-center justify-center gap-2.5 transition-all mb-5 hover:border-white/20 active:scale-[0.99]"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24">
                  <path
                    fill="#EA4335"
                    d="M12 5c1.6 0 3 .6 4.1 1.7l3.1-3.1C17.3 1.8 14.8 1 12 1 7.5 1 3.7 3.6 1.9 7.3l3.7 2.9C6.5 7.4 9 5 12 5z"
                  />
                  <path
                    fill="#4285F4"
                    d="M23.5 12.3c0-.8-.1-1.7-.2-2.3H12v4.5h6.5c-.3 1.5-1.1 2.8-2.4 3.7l3.7 2.9c2.2-2 3.7-5 3.7-8.8z"
                  />
                  <path
                    fill="#FBBC05"
                    d="M5.6 14.8c-.2-.7-.4-1.5-.4-2.8s.2-2.1.4-2.8L1.9 6.3C.7 8.7 0 10.8 0 12s.7 3.3 1.9 5.7l3.7-2.9z"
                  />
                  <path
                    fill="#34A853"
                    d="M12 23c3.2 0 6-1.1 8-3l-3.7-2.9c-1.1.7-2.5 1.2-4.3 1.2-3 0-5.5-2.4-6.4-5.2L1.9 16c1.8 3.7 5.6 7 10.1 7z"
                  />
                </svg>
                <span>Continue with Google</span>
              </button>

              <div className="relative flex items-center justify-center mb-5">
                <div className="border-t border-white/10 w-full" />
                <span className="bg-[#121216] px-2 text-[10px] uppercase tracking-wider text-muted font-mono absolute">
                  or with email
                </span>
              </div>

              <form onSubmit={handleSubmit} className="space-y-4">
                {mode === 'signup' && (
                  <div>
                    <label
                      htmlFor="login-name"
                      className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono"
                    >
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
                  <label
                    htmlFor="login-email"
                    className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono"
                  >
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
                  <div className="flex justify-between items-center mb-1">
                    <label
                      htmlFor="login-password"
                      className="block text-xs uppercase tracking-wider text-muted font-mono"
                    >
                      Password
                    </label>
                    {mode === 'signin' && (
                      <button
                        type="button"
                        onClick={() => setIsForgotOpen(true)}
                        className="text-[11px] text-accent hover:underline font-sans"
                      >
                        Forgot password?
                      </button>
                    )}
                  </div>
                  <div className="relative">
                    <input
                      id="login-password"
                      type={showPassword ? 'text' : 'password'}
                      placeholder="At least 8 characters"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono pr-12"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-2.5 text-xs text-muted hover:text-main select-none font-mono"
                    >
                      {showPassword ? 'Hide' : 'Show'}
                    </button>
                  </div>
                </div>

                {mode === 'signup' && (
                  <div>
                    <label
                      htmlFor="login-confirm-password"
                      className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono"
                    >
                      Confirm Password
                    </label>
                    <input
                      id="login-confirm-password"
                      type={showPassword ? 'text' : 'password'}
                      placeholder="Repeat your password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                      required
                    />
                  </div>
                )}

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
                  disabled={isSubmitting}
                  className="w-full py-3 rounded-xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.99] transition-all shadow-lg shadow-accent/20 disabled:opacity-50 mt-2"
                >
                  {isSubmitting
                    ? 'Processing...'
                    : mode === 'signup'
                    ? 'Create Spendle Account'
                    : 'Sign In'}
                </button>
              </form>
            </>
          )}
        </GlassCard>
      </motion.div>

      {/* Forgot Password Modal */}
      <ForgotPasswordModal
        isOpen={isForgotOpen}
        onClose={() => setIsForgotOpen(false)}
        initialEmail={email}
      />
    </div>
  );
}
