import React from 'react';
import {
  siZomato,
  siSwiggy,
  siNetflix,
  siSpotify,
  siYoutube,
  siUber,
  siPaytm,
  siPhonepe,
  siGooglepay,
  siJio,
  siAirtel,
  siBigbasket,
} from 'simple-icons';

/**
 * Registry of known brands with official simple-icons vectors
 */
const BRAND_ICONS = {
  zomato: { name: 'Zomato', icon: siZomato, color: '#E23744' },
  swiggy: { name: 'Swiggy', icon: siSwiggy, color: '#FC8019' },
  netflix: { name: 'Netflix', icon: siNetflix, color: '#E50914' },
  spotify: { name: 'Spotify', icon: siSpotify, color: '#1ED760' },
  youtube: { name: 'YouTube', icon: siYoutube, color: '#FF0000' },
  uber: { name: 'Uber', icon: siUber, color: '#000000' },
  paytm: { name: 'Paytm', icon: siPaytm, color: '#00BAF2' },
  phonepe: { name: 'PhonePe', icon: siPhonepe, color: '#5F259F' },
  googlepay: { name: 'Google Pay', icon: siGooglepay, color: '#4285F4' },
  gpay: { name: 'Google Pay', icon: siGooglepay, color: '#4285F4' },
  jio: { name: 'Jio', icon: siJio, color: '#0A2885' },
  airtel: { name: 'Airtel', icon: siAirtel, color: '#E40000' },
  bigbasket: { name: 'bigbasket', icon: siBigbasket, color: '#A5CD39' },
};

/**
 * Registry of brands absent from simple-icons requiring monogram chip fallback
 * (Never hand-drawn, traced or invented)
 */
const MONOGRAM_BRANDS = {
  amazon: { name: 'Amazon', letter: 'A', color: '#FF9900' },
  flipkart: { name: 'Flipkart', letter: 'F', color: '#2874F0' },
  ola: { name: 'Ola', letter: 'O', color: '#F5B83D' },
  blinkit: { name: 'Blinkit', letter: 'B', color: '#F8CB46' },
  zepto: { name: 'Zepto', letter: 'Z', color: '#8229F6' },
  irctc: { name: 'IRCTC', letter: 'I', color: '#213D77' },
};

/**
 * Category fallback icons (SVG glyphs) when note contains no known brand
 */
const CATEGORY_GLYPHS = {
  food: '🍽️',
  groceries: '🛒',
  transport: '🚇',
  bills: '⚡',
  entertainment: '🎬',
  shopping: '🛍️',
  health: '💊',
  freelance: '💻',
  salary: '💼',
};

/**
 * Extracts a known merchant from user note or title text
 * @param {string} text - e.g. "Swiggy dinner with friends" or "Metro recharge"
 * @returns {object|null} Match details or null
 */
export function merchantFromText(text = '') {
  if (!text || typeof text !== 'string') return null;
  const lower = text.toLowerCase();

  // 1. Check official simple-icons brands
  const iconKeys = Object.keys(BRAND_ICONS);
  for (const key of iconKeys) {
    if (new RegExp(`\\b${key}\\b`, 'i').test(lower) || lower.includes(key)) {
      return { type: 'icon', ...BRAND_ICONS[key] };
    }
  }

  // 2. Check monogram brands
  const monoKeys = Object.keys(MONOGRAM_BRANDS);
  for (const key of monoKeys) {
    if (new RegExp(`\\b${key}\\b`, 'i').test(lower) || lower.includes(key)) {
      return { type: 'monogram', ...MONOGRAM_BRANDS[key] };
    }
  }

  return null;
}

/**
 * Calculates relative luminance to check contrast against dark UI
 */
function isTooDarkForTheme(hex = '#000000') {
  const clean = hex.replace('#', '');
  const r = parseInt(clean.substring(0, 2), 16) / 255;
  const g = parseInt(clean.substring(2, 4), 16) / 255;
  const b = parseInt(clean.substring(4, 6), 16) / 255;
  const luminance = 0.2126 * r + 0.7152 * g + 0.0722 * b;
  return luminance < 0.22;
}

/**
 * MerchantLogo
 * Renders a 36px round chip for transaction lists, cards, and marquee strips.
 *
 * @param {string} name - Explicit merchant name (e.g. "zomato")
 * @param {string} note - Transaction note to parse (e.g. "Zomato biryani")
 * @param {string} category - Category fallback name (e.g. "Food")
 * @param {string} className - Additional CSS classes
 */
export default function MerchantLogo({
  name = '',
  note = '',
  category = '',
  className = '',
}) {
  const merchant = merchantFromText(name || note);

  // 1. Matched brand with simple-icons vector
  if (merchant && merchant.type === 'icon') {
    const brandColor = merchant.color || `#${merchant.icon.hex}`;
    const tooDark = isTooDarkForTheme(brandColor);
    const iconColor = tooDark ? '#EDEAE3' : brandColor;

    return (
      <div
        className={`w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0 transition-transform ${className}`}
        style={{
          backgroundColor: `${brandColor}26`, // ~15% alpha
          borderColor: `${brandColor}4D`, // ~30% border alpha
          borderWidth: '1px',
        }}
        title={merchant.name}
        aria-label={merchant.name}
      >
        <svg
          role="img"
          viewBox="0 0 24 24"
          className="w-4 h-4"
          fill={iconColor}
        >
          <path d={merchant.icon.path} />
        </svg>
      </div>
    );
  }

  // 2. Matched brand requiring monogram chip fallback
  if (merchant && merchant.type === 'monogram') {
    const brandColor = merchant.color;
    return (
      <div
        className={`w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0 font-sans font-semibold text-xs tracking-wide ${className}`}
        style={{
          backgroundColor: `${brandColor}26`,
          borderColor: `${brandColor}4D`,
          borderWidth: '1px',
          color: brandColor,
        }}
        title={merchant.name}
        aria-label={merchant.name}
      >
        {merchant.letter}
      </div>
    );
  }

  // 3. Fallback to Category Icon
  const cleanCat = (category || '').toLowerCase();
  let emoji = '💸';
  for (const [catKey, glyph] of Object.entries(CATEGORY_GLYPHS)) {
    if (cleanCat.includes(catKey)) {
      emoji = glyph;
      break;
    }
  }

  return (
    <div
      className={`w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0 bg-white/5 border border-white/10 text-xs ${className}`}
      title={category || 'Transaction'}
      aria-label={category || 'Transaction'}
    >
      <span>{emoji}</span>
    </div>
  );
}

