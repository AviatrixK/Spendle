import React, { useState, useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import { authApi } from '../api';

/**
 * EmailVerification
 * A polished 4-digit OTP verification interface featuring:
 * - Single-digit auto-advancing input fields
 * - Paste support for full 4-digit codes
 * - 2-minute expiration countdown timer
 * - 60-second resend cooldown timer
 * - Keyboard navigation (arrow keys & backspace)
 */
export default function EmailVerification({ email, onSuccess, onBack }) {
  const [otp, setOtp] = useState(['', '', '', '']);
  const [error, setError] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);
  const [isResending, setIsResending] = useState(false);

  // 120-second (2 minute) OTP expiration countdown
  const [expirySeconds, setExpirySeconds] = useState(120);
  // 60-second cooldown before a new code can be dispatched
  const [cooldownSeconds, setCooldownSeconds] = useState(60);

  const inputRefs = [useRef(null), useRef(null), useRef(null), useRef(null)];

  // Focus the first input box on initial mount
  useEffect(() => {
    inputRefs[0]?.current?.focus();
  }, []);

  // Expiration countdown ticker
  useEffect(() => {
    if (expirySeconds <= 0) return;
    const timer = setInterval(() => {
      setExpirySeconds((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, [expirySeconds]);

  // Cooldown countdown ticker
  useEffect(() => {
    if (cooldownSeconds <= 0) return;
    const timer = setInterval(() => {
      setCooldownSeconds((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, [cooldownSeconds]);

  // Format seconds as MM:SS
  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`;
  };

  const handleInputChange = (index, value) => {
    // Only accept numeric inputs
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

    // Auto-advance to next input field
    if (index < 3) {
      inputRefs[index + 1]?.current?.focus();
    }
  };

  const handleKeyDown = (index, e) => {
    if (e.key === 'Backspace') {
      if (!otp[index] && index > 0) {
        inputRefs[index - 1]?.current?.focus();
      }
    } else if (e.key === 'ArrowLeft' && index > 0) {
      inputRefs[index - 1]?.current?.focus();
    } else if (e.key === 'ArrowRight' && index < 3) {
      inputRefs[index + 1]?.current?.focus();
    }
  };

  const handlePaste = (e) => {
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

    // Focus last filled or 4th input
    const focusIndex = Math.min(digits.length, 3);
    inputRefs[focusIndex]?.current?.focus();
  };

  const handleSubmit = async (e) => {
    e?.preventDefault();
    setError('');
    setSuccessMsg('');

    const fullCode = otp.join('');
    if (fullCode.length !== 4) {
      setError('Please enter all 4 digits of the verification code.');
      return;
    }

    if (expirySeconds <= 0) {
      setError('This verification code has expired. Please request a new code.');
      return;
    }

    setIsVerifying(true);
    try {
      const data = await authApi.verifyEmail(email, fullCode);
      setSuccessMsg('Email verified successfully! Redirecting...');
      setTimeout(() => {
        onSuccess(data);
      }, 700);
    } catch (err) {
      setError(err.message || 'Verification failed. Please check the code and try again.');
    } finally {
      setIsVerifying(false);
    }
  };

  const handleResend = async () => {
    if (cooldownSeconds > 0 || isResending) return;
    setError('');
    setSuccessMsg('');
    setIsResending(true);

    try {
      await authApi.resendOtp(email, 'email_verification');
      setSuccessMsg('A fresh verification code has been dispatched to your email.');
      setCooldownSeconds(60);
      setExpirySeconds(120);
      setOtp(['', '', '', '']);
      inputRefs[0]?.current?.focus();
    } catch (err) {
      setError(err.message || 'Failed to resend code. Please try again.');
    } finally {
      setIsResending(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-xl font-serif text-main">Check Your Inbox</h2>
        <p className="text-xs text-muted mt-2">
          We sent a 4-digit code to <span className="text-accent font-medium">{email}</span>
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* 4-Digit Input Boxes */}
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
              onChange={(e) => handleInputChange(index, e.target.value)}
              onKeyDown={(e) => handleKeyDown(index, e)}
              onPaste={handlePaste}
              className="w-14 h-16 text-center text-2xl font-mono font-bold bg-white/5 border border-white/10 rounded-2xl text-main focus:border-accent focus:ring-1 focus:ring-accent outline-none transition-all shadow-inner"
            />
          ))}
        </div>

        {/* Timers & Status */}
        <div className="flex justify-between items-center text-xs font-mono px-2 text-muted">
          <span>
            {expirySeconds > 0 ? (
              <>
                Expires in:{' '}
                <span className={expirySeconds <= 30 ? 'text-semantic-loss font-semibold' : 'text-main font-semibold'}>
                  {formatTime(expirySeconds)}
                </span>
              </>
            ) : (
              <span className="text-semantic-loss font-semibold">Code expired</span>
            )}
          </span>

          <button
            type="button"
            onClick={handleResend}
            disabled={cooldownSeconds > 0 || isResending}
            className="text-accent hover:underline disabled:text-muted/50 disabled:no-underline font-sans transition-colors"
          >
            {isResending
              ? 'Sending...'
              : cooldownSeconds > 0
              ? `Resend in ${cooldownSeconds}s`
              : 'Resend Code'}
          </button>
        </div>

        {error && (
          <motion.p
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            className="text-xs text-semantic-loss bg-semantic-loss/10 border border-semantic-loss/20 rounded-xl p-3 text-center"
          >
            {error}
          </motion.p>
        )}

        {successMsg && (
          <motion.p
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            className="text-xs text-semantic-gain bg-semantic-gain/10 border border-semantic-gain/20 rounded-xl p-3 text-center font-medium"
          >
            {successMsg}
          </motion.p>
        )}

        {/* Submit Button */}
        <button
          type="submit"
          disabled={isVerifying || otp.join('').length !== 4 || expirySeconds <= 0}
          className="w-full py-3 rounded-xl bg-accent text-[#0E0E11] font-semibold text-sm hover:bg-accent-hover active:scale-[0.99] transition-all shadow-lg shadow-accent/20 disabled:opacity-50"
        >
          {isVerifying ? 'Verifying Code...' : 'Confirm & Enter Spendle'}
        </button>

        {/* Back Link */}
        <div className="text-center">
          <button
            type="button"
            onClick={onBack}
            className="text-xs text-muted hover:text-main transition-colors underline"
          >
            Back to Sign In / Change Email
          </button>
        </div>
      </form>
    </div>
  );
}

