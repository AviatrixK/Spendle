/**
 * Monetary & Date Formatting Helpers for Spendle
 * All money internally is strictly represented as integer paise (1 Rupee = 100 Paise).
 */

/**
 * Formats integer paise into an Indian Rupee string with Indian numbering grouping (e.g. ₹90,000).
 * Displays decimals only when paise remainder is non-zero.
 */
export function formatINR(paise) {
  if (paise == null || isNaN(paise)) return '₹0';
  const num = Number(paise);
  const rupees = num / 100;
  const hasDecimals = num % 100 !== 0;

  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: hasDecimals ? 2 : 0,
    maximumFractionDigits: 2,
  }).format(rupees);
}

/**
 * Converts a typed rupee amount (string or number) into integer paise.
 * Uses Math.round to protect against IEEE 754 floating point inaccuracies.
 */
export function rupeesToPaise(input) {
  if (input === '' || input == null) return 0;
  const num = typeof input === 'number' ? input : parseFloat(String(input).replace(/,/g, ''));
  if (isNaN(num) || num <= 0) return 0;
  return Math.round(num * 100);
}

/**
 * Converts integer paise back into a plain rupee string for input editing.
 */
export function paiseToRupees(paise) {
  if (paise == null || paise === 0) return '';
  const val = Number(paise) / 100;
  return val % 1 === 0 ? String(val) : val.toFixed(2);
}

/**
 * Returns current month key in "YYYY-MM" format in UTC.
 */
export function monthKey(date = new Date()) {
  const d = new Date(date);
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  return `${year}-${month}`;
}

/**
 * Returns the human-readable month name (e.g. "Oct 2026").
 */
export function formatMonthName(monthStr) {
  if (!monthStr) return '';
  const [yearStr, monthPart] = monthStr.split('-');
  const year = parseInt(yearStr, 10);
  const month = parseInt(monthPart, 10);
  if (isNaN(year) || isNaN(month)) return monthStr;
  const d = new Date(year, month - 1, 1);
  return d.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
}

/**
 * Calculates the previous calendar month key ("YYYY-MM").
 */
export function prevMonthKey(monthStr) {
  if (!monthStr) return monthKey();
  const [yearStr, monthPart] = monthStr.split('-');
  const year = parseInt(yearStr, 10);
  const month = parseInt(monthPart, 10);
  const d = new Date(year, month - 2, 1);
  return monthKey(d);
}

/**
 * Calculates the next calendar month key ("YYYY-MM").
 */
export function nextMonthKey(monthStr) {
  if (!monthStr) return monthKey();
  const [yearStr, monthPart] = monthStr.split('-');
  const year = parseInt(yearStr, 10);
  const month = parseInt(monthPart, 10);
  const d = new Date(year, month, 1);
  return monthKey(d);
}

