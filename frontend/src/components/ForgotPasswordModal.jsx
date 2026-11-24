import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import GlassCard from './GlassCard';
import { authApi } from '../api';

export default function ForgotPasswordModal({ isOpen, onClose, initialEmail = '' }) {
  const [step, setStep] = useState(1); // 1: Email input, 2: Code & New Password
  const [email, setEmail] = useState(initialEmail);
  const [otp, setOtp] = useState(['', '', '', '']);
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  // Timers for Step 2
  const [expirySeconds, setExpirySeconds] = useState(120);
  const [cooldownSeconds, setCooldownSeconds] = useState(60);

  const inputRefs = [useRef(null), useRef(null), useRef(null), useRef(null)];

  useEffect(() => {
    if (initialEmail) {
      setEmail(initialEmail);
    }
  }, [initialEmail]);

  useEffect(() => {
    if (step === 2 && inputRefs[0]?.current) {
      inputRefs[0].current.focus();
    }
  }, [step]);

  // Timers
  useEffect(() => {
    if (step !== 2 || expirySeconds <= 0) return;
    const timer = setInterval(() => {
      setExpirySeconds((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, [step, expirySeconds]);

  useEffect(() => {
    if (step !== 2 || cooldownSeconds <= 0) return;
    const timer = setInterval(() => {
      setCooldownSeconds((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, [step, cooldownSeconds]);

  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`;
  };

  const handleOtpChange = (index, value) => {
    const cleanValue = value.replace(/\D/g, '');
    if (!cleanValue) {
      const updated = [...otp];
      updated[index] = '';
      setOtp(updated);
      return;
    }

    const digit = cleanValue.slice(-1);
    const updated = [...otp];
    updated[index] = digit;
    setOtp(updated);
    setError('');

    if (index < 3) {
      inputRefs[index + 1]?.current?.focus();
    }
  };

  const handleOtpKeyDown = (index, e) => {
    if (e.key === 'Backspace' && !otp[index] && index > 0) {
      inputRefs[index - 1]?.current?.focus();
    }
  };

  const handleOtpPaste = (e) => {
    e.preventDefault();
    const pasteData = e.clipboardData.getData('text').trim().replace(/\D/g, '');
    if (!pasteData) return;

    const digits = pasteData.slice(0, 4).split('');
    const updated = ['', '', '', ''];
    digits.forEach((d, i) => {
      updated[i] = d;
    });
    setOtp(updated);
    setError('');
    const focusIndex = Math.min(digits.length, 3);
    inputRefs[focusIndex]?.current?.focus();
  };

  const handleRequestOTP = async (e) => {
    e.preventDefault();
    if (!email) {
      setError('Please provide your email address.');
      return;
    }

    setError('');
    setLoading(true);
    try {
      await authApi.forgotPassword(email);
      setStep(2);
      setExpirySeconds(120);
      setCooldownSeconds(60);
    } catch (err) {
      setError(err.message || 'Failed to request reset code.');
    } finally {
      setLoading(false);
    }
  };

  const handleResetPassword = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    const code = otp.join('');
    if (code.length !== 4) {
      setError('Please enter all 4 digits of the reset code.');
      return;
    }

    if (newPassword.length < 8) {
      setError('New password must be at least 8 characters long.');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    setLoading(true);
    try {
      await authApi.resetPassword(email, code, newPassword);
      setSuccess('Your password has been reset successfully! You can now log in.');
      setTimeout(() => {
        handleClose();
      }, 1500);
    } catch (err) {
      setError(err.message || 'Failed to reset password. Please check the code.');
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    if (cooldownSeconds > 0 || loading) return;
    setError('');
    setLoading(true);
    try {
      await authApi.resendOtp(email, 'password_reset');
      setCooldownSeconds(60);
      setExpirySeconds(120);
      setOtp(['', '', '', '']);
      inputRefs[0]?.current?.focus();
    } catch (err) {
      setError(err.message || 'Failed to resend reset code.');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setStep(1);
    setOtp(['', '', '', '']);
    setNewPassword('');
    setConfirmPassword('');
    setError('');
    setSuccess('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          className="w-full max-w-md"
        >
          <GlassCard className="p-7 border-white/10 relative">
            <button
              onClick={handleClose}
              className="absolute top-4 right-4 text-muted hover:text-main text-lg font-mono transition-colors"
            >
              ✕
            </button>

            {step === 1 ? (
              <div>
                <h3 className="text-xl font-serif text-main mb-2">Reset Password</h3>
                <p className="text-xs text-muted mb-6">
                  Enter your email address and we will send you a 4-digit verification code to reset your password.
                </p>

                <form onSubmit={handleRequestOTP} className="space-y-4">
                  <div>
                    <label className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                      Email Address
                    </label>
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="you@example.com"
                      className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent"
                      required
                    />
                  </div>

                  {error && (
                    <p className="text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-3 text-center">
                      {error}
                    </p>
                  )}

                  <button
                    type="submit"
                    disabled={loading || !email}
                    className="w-full py-3 rounded-xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.99] transition-all disabled:opacity-50"
                  >
                    {loading ? 'Sending Code...' : 'Send Reset Code'}
                  </button>
                </form>
              </div>
            ) : (
              <div>
                <h3 className="text-xl font-serif text-main mb-2">Enter Reset Code</h3>
                <p className="text-xs text-muted mb-5">
                  We sent a 4-digit code to <span className="text-accent font-medium">{email}</span>
                </p>

                <form onSubmit={handleResetPassword} className="space-y-4">
                  {/* OTP inputs */}
                  <div>
                    <label className="block text-xs uppercase tracking-wider text-muted mb-2 font-mono text-center">
                      4-Digit Verification Code
                    </label>
                    <div className="flex justify-center gap-3">
                      {otp.map((digit, index) => (
                        <input
                          key={index}
                          ref={inputRefs[index]}
                          type="text"
                          inputMode="numeric"
                          pattern="[0-9]*"
                          maxLength={1}
                          value={digit}
                          onChange={(e) => handleOtpChange(index, e.target.value)}
                          onKeyDown={(e) => handleOtpKeyDown(index, e)}
                          onPaste={handleOtpPaste}
                          className="w-12 h-14 text-center text-xl font-mono font-bold bg-white/5 border border-white/10 rounded-xl text-main focus:border-accent outline-none transition-all shadow-inner"
                        />
                      ))}
                    </div>
                  </div>

                  {/* Timers */}
                  <div className="flex justify-between items-center text-xs font-mono text-muted px-1">
                    <span>
                      {expirySeconds > 0 ? `Expires in ${formatTime(expirySeconds)}` : 'Expired'}
                    </span>
                    <button
                      type="button"
                      onClick={handleResend}
                      disabled={cooldownSeconds > 0 || loading}
                      className="text-accent hover:underline disabled:text-muted/50 disabled:no-underline font-sans"
                    >
                      {cooldownSeconds > 0 ? `Resend (${cooldownSeconds}s)` : 'Resend Code'}
                    </button>
                  </div>

                  {/* New Password */}
                  <div>
                    <label className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                      New Password
                    </label>
                    <div className="relative">
                      <input
                        type={showPassword ? 'text' : 'password'}
                        value={newPassword}
                        onChange={(e) => setNewPassword(e.target.value)}
                        placeholder="At least 8 characters"
                        className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                        required
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-2.5 text-xs text-muted hover:text-main"
                      >
                        {showPassword ? 'Hide' : 'Show'}
                      </button>
                    </div>
                  </div>

                  {/* Confirm Password */}
                  <div>
                    <label className="block text-xs uppercase tracking-wider text-muted mb-1 font-mono">
                      Confirm New Password
                    </label>
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder="Repeat new password"
                      className="w-full bg-white/5 border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-main placeholder:text-muted/60 outline-none focus:border-accent font-mono"
                      required
                    />
                  </div>

                  {error && (
                    <p className="text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-3 text-center">
                      {error}
                    </p>
                  )}

                  {success && (
                    <p className="text-xs text-semantic-gain bg-semantic-gain/10 border border-semantic-gain/20 rounded-xl p-3 text-center font-medium">
                      {success}
                    </p>
                  )}

                  <button
                    type="submit"
                    disabled={loading || otp.join('').length !== 4 || !newPassword || !confirmPassword}
                    className="w-full py-3 rounded-xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.99] transition-all disabled:opacity-50"
                  >
                    {loading ? 'Updating Password...' : 'Reset Password'}
                  </button>
                </form>
              </div>
            )}
          </GlassCard>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}

