/**
 * Spendle API client
 * Communicates with the Go backend through Vite's dev-server proxy at /api
 * or via Vercel rewrite / custom base URL.
 * Includes credentials ('include') so the httpOnly cookie is automatically passed.
 */

const API_BASE = import.meta.env.VITE_API_BASE_URL || '';
const BASE_PATH = `${API_BASE}/api/v1`;

export async function request(endpoint, options = {}) {
  const { body, headers = {}, ...customConfig } = options;

  const config = {
    method: options.method || (body ? 'POST' : 'GET'),
    credentials: 'include',
    headers: {
      ...headers,
    },
    ...customConfig,
  };

  if (body !== undefined) {
    config.headers['Content-Type'] = 'application/json';
    config.body = JSON.stringify(body);
  }

  const res = await fetch(`${BASE_PATH}${endpoint}`, config);

  // 204 No Content has no body
  if (res.status === 204) {
    return null;
  }

  let data = null;
  const contentType = res.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    try {
      data = await res.json();
    } catch {
      data = null;
    }
  }

  if (!res.ok) {
    const errorMsg = data && data.error ? data.error : `Request failed with status ${res.status}`;
    const err = new Error(errorMsg);
    err.status = res.status;
    err.data = data;
    throw err;
  }

  return data;
}

// Authentication API
export const authApi = {
  me: () => request('/auth/me'),
  login: (email, password) => request('/auth/login', { body: { email, password } }),
  signup: (name, email, password) => request('/auth/signup', { body: { name, email, password } }),
  verifyEmail: (email, code) => request('/auth/verify-email', { body: { email, code } }),
  resendOtp: (email, purpose = 'email_verification') =>
    request('/auth/resend-otp', { body: { email, purpose } }),
  forgotPassword: (email) => request('/auth/forgot-password', { body: { email } }),
  resetPassword: (email, code, new_password) =>
    request('/auth/reset-password', { body: { email, code, new_password } }),
  logout: () => request('/auth/logout', { method: 'POST' }),
  googleLoginUrl: `${BASE_PATH}/auth/google/login`,
};

// Dashboard API
export const dashboardApi = {
  get: (month) => request(`/dashboard${month ? `?month=${month}` : ''}`),
};

// Budget & Category API
export const budgetApi = {
  getCategories: () => request('/categories'),
  createCategory: (name) => request('/categories', { body: { name } }),
  deleteCategory: (id) => request(`/categories/${id}`, { method: 'DELETE' }),
  getBudgets: (month) => request(`/budgets${month ? `?month=${month}` : ''}`),
  setBudget: (category_id, month, limit_paise) =>
    request('/budgets', { method: 'PUT', body: { category_id, month, limit_paise } }),
};

// Transactions API
export const transactionApi = {
  list: (month, limit = 10) =>
    request(`/transactions?month=${month}&limit=${limit}`),
  create: (data) => request('/transactions', { body: data }),
  delete: (id) => request(`/transactions/${id}`, { method: 'DELETE' }),
  getSummary: (month) => request(`/transactions/summary?month=${month}`),
};

// Savings Goals API
export const goalApi = {
  list: () => request('/goals'),
  create: (data) => request('/goals', { body: data }),
  update: (id, data) => request(`/goals/${id}`, { method: 'PUT', body: data }),
  delete: (id) => request(`/goals/${id}`, { method: 'DELETE' }),
  closeMonth: (month) => request('/months/close', { body: { month } }),
};
