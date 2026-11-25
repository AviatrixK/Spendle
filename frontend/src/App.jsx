import React, { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { authApi } from './api';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Goals from './pages/Goals';
import AmbientBackground from './components/AmbientBackground';
import CustomCursor from './components/CustomCursor';
import SpendleLoader from './components/SpendleLoader';

// Lazy-load landing page so app bundle stays small
const Landing = lazy(() => import('./landing/Landing'));

/**
 * ProtectedRoute Guard
 * Verifies authenticated state against the backend cookie via GET /api/v1/auth/me.
 * Redirects unauthenticated users to /login.
 */
function ProtectedRoute({ children }) {
  const { data: user, isLoading, isError } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: authApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return <SpendleLoader label="Checking your session" />;
  }

  if (isError || !user) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

/**
 * PublicRoute Guard
 * If user is already authenticated, redirects them away from /login to the dashboard.
 */
function PublicRoute({ children }) {
  const { data: user, isLoading } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: authApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return <SpendleLoader label="Checking your session" />;
  }

  if (user) {
    return <Navigate to="/" replace />;
  }

  return children;
}

/**
 * HomeRoute Dispatcher
 * Logged-out visitors see the editorial Landing page.
 * Logged-in users see their live financial Dashboard.
 */
function HomeRoute() {
  const { data: user, isLoading } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: authApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return <SpendleLoader label="Preparing your dashboard" />;
  }

  if (user) {
    return <Dashboard />;
  }

  return (
    <Suspense
      fallback={
        <SpendleLoader label="Loading Spendle" />
      }
    >
      <Landing />
    </Suspense>
  );
}

function WelcomeRoute() {
  return <Suspense fallback={<SpendleLoader label="Loading Spendle" />}><Landing /></Suspense>;
}

export default function App() {
  return (
    <div className="relative min-h-screen bg-bg text-main selection:bg-accent selection:text-bg">
      {/* 1. Global Ambient Background: Dot-Grid Canvas & 500px Radial Spotlight at -z-10 */}
      <AmbientBackground intensity="full" />

      {/* 2. Global Custom Cursor: Gold Dot & Lagging Ring at z-[9999] */}
      <CustomCursor />

      <Routes>
        <Route
          path="/login"
          element={
            <PublicRoute>
              <Login />
            </PublicRoute>
          }
        />
        <Route path="/" element={<HomeRoute />} />
        <Route path="/welcome" element={<WelcomeRoute />} />
        <Route
          path="/goals"
          element={
            <ProtectedRoute>
              <Goals />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}
