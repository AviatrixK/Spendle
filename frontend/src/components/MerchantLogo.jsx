import React, { useState } from 'react';

const BRANDS = [
  ['zomato', 'Zomato', '#E23744'], ['swiggy', 'Swiggy', '#FC8019'],
  ['netflix', 'Netflix', '#E50914'], ['spotify', 'Spotify', '#1ED760'],
  ['youtube', 'YouTube', '#FF0000'], ['uber', 'Uber', '#111111'],
  ['paytm', 'Paytm', '#00BAF2'], ['phonepe', 'PhonePe', '#5F259F'],
  ['googlepay', 'Google Pay', '#4285F4'], ['gpay', 'Google Pay', '#4285F4'],
  ['jio', 'Jio', '#0A2885'], ['airtel', 'Airtel', '#E40000'],
  ['bigbasket', 'bigbasket', '#559A35'],
];
const CATEGORY_GLYPHS = { food: '🍽', groceries: '🛒', transport: '🚆', bills: '⚡', entertainment: '🎬', shopping: '🛍', health: '✚', freelance: '💻', salary: '▤' };

export function merchantFromText(text = '') {
  const value = String(text).toLowerCase();
  const normalized = value.replace(/[^a-z0-9]/g, '');
  return BRANDS.find(([key]) => normalized.includes(key))
    || (normalized.includes('gpay') ? BRANDS.find(([key]) => key === 'googlepay') : null)
    || null;
}

export default function MerchantLogo({ name = '', note = '', category = '', className = '' }) {
  const brand = merchantFromText(name || note);
  const [failed, setFailed] = useState(false);
  const key = brand?.[0];
  if (brand && !failed) {
    return <span className={`merchant-logo ${className}`} title={brand[1]}><img src={`/brands/${key}.png`} alt={`${brand[1]} logo`} width="40" height="40" onError={() => setFailed(true)} /></span>;
  }
  if (brand) {
    return <span className={`merchant-logo merchant-monogram ${className}`} title={brand[1]} aria-label={`${brand[1]} logo`} style={{ '--brand-color': brand[2] }}>{brand[1].slice(0, 1)}</span>;
  }
  const cleanCat = category.toLowerCase();
  const glyph = Object.entries(CATEGORY_GLYPHS).find(([key]) => cleanCat.includes(key))?.[1] || '₹';
  return <span className={`merchant-logo merchant-category ${className}`} title={category || 'Transaction'} aria-label={category || 'Transaction'}>{glyph}</span>;
}
