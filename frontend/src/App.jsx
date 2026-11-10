import React, { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { authApi } from './api';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Goals from './pages/Goals';

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
    return (
      <div className="min-h-screen flex items-center justify-center text-muted font-mono text-xs">
        Checking session...
      </div>
    );
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
    return (
      <div className="min-h-screen flex items-center justify-center text-muted font-mono text-xs">
        Loading...
      </div>
    );
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
    return (
      <div className="min-h-screen flex items-center justify-center text-muted font-mono text-xs">
        Loading...
      </div>
    );
  }

  if (user) {
    return <Dashboard />;
  }

  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center text-muted font-mono text-xs">
          Loading...
        </div>
      }
    >
      <Landing />
    </Suspense>
  );
}

export default function App() {
  return (
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
  );
}
