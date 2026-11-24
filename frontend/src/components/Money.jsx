import React from 'react';

/**
 * Money component
 * Formats integer paise using the Indian numbering system (en-IN).
 * Decimals are displayed in a lighter tone (opacity: 0.5) per the editorial design spec.
 * Uses tabular figures for clean vertical alignment in lists and dashboards.
 *
 * @param {number} paise - Amount in integer paise (e.g. 425000 = ₹4,250.00)
 * @param {boolean|string} sign - If true, displays + for positive amounts. If explicitly '+', forces plus.
 * @param {boolean} showCurrency - Whether to prepend the ₹ symbol (default: true)
 * @param {boolean} showDecimals - Whether to render .XX decimals (default: true)
 * @param {string} className - Additional CSS classes
 * @param {string} decimalClassName - Additional classes for decimal span
 */
export default function Money({
  paise = 0,
  sign = false,
  showCurrency = true,
  showDecimals = true,
  className = '',
  decimalClassName = '',
}) {
  const safePaise = Number.isFinite(paise) ? Math.round(paise) : 0;
  const isNegative = safePaise < 0;
  const absPaise = Math.abs(safePaise);

  const rupees = Math.floor(absPaise / 100);
  const fraction = absPaise % 100;

  // Indian digit grouping (e.g. 1,50,000)
  const formattedRupees = new Intl.NumberFormat('en-IN').format(rupees);
  const formattedFraction = `.${fraction.toString().padStart(2, '0')}`;

  let signPrefix = '';
  if (isNegative) {
    signPrefix = '-';
  } else if (sign === true || sign === '+') {
    signPrefix = '+';
  }

  return (
    <span
      className={`font-mono tabular-nums inline-flex items-baseline ${className}`}
      style={{ fontVariantNumeric: 'tabular-nums' }}
    >
      {signPrefix && <span className="mr-0.5">{signPrefix}</span>}
      {showCurrency && <span className="mr-0.5 font-sans font-normal opacity-90">₹</span>}
      <span>{formattedRupees}</span>
      {showDecimals && (
        <span
          className={`opacity-50 text-[0.82em] font-normal tracking-tight ${decimalClassName}`}
        >
          {formattedFraction}
        </span>
      )}
    </span>
  );
}

