// One-off generator for public/og-image.png (1200x630 social share card).
// Run with: npm install --no-save sharp && node scripts/generate-og-image.mjs
// The generated PNG is committed to the repo; sharp is NOT a runtime dependency.
import sharp from 'sharp';
import { fileURLToPath } from 'node:url';

const svg = `<svg width="1200" height="630" viewBox="0 0 1200 630" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <radialGradient id="a" cx="18%" cy="22%" r="55%">
      <stop offset="0%" stop-color="#00D4FF" stop-opacity="0.22"/>
      <stop offset="100%" stop-color="#00D4FF" stop-opacity="0"/>
    </radialGradient>
    <radialGradient id="b" cx="88%" cy="30%" r="55%">
      <stop offset="0%" stop-color="#00FF88" stop-opacity="0.18"/>
      <stop offset="100%" stop-color="#00FF88" stop-opacity="0"/>
    </radialGradient>
    <radialGradient id="c" cx="70%" cy="95%" r="60%">
      <stop offset="0%" stop-color="#FF00AA" stop-opacity="0.15"/>
      <stop offset="100%" stop-color="#FF00AA" stop-opacity="0"/>
    </radialGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="0">
      <stop offset="0%" stop-color="#00D4FF"/>
      <stop offset="100%" stop-color="#00FF88"/>
    </linearGradient>
    <style>
      text { font-family: "DejaVu Sans", "FreeSans", sans-serif; }
    </style>
  </defs>

  <rect width="1200" height="630" fill="#080809"/>
  <rect width="1200" height="630" fill="url(#a)"/>
  <rect width="1200" height="630" fill="url(#b)"/>
  <rect width="1200" height="630" fill="url(#c)"/>
  <rect x="0" y="0" width="1200" height="6" fill="url(#accent)"/>

  <!-- brand -->
  <g transform="translate(80, 92)">
    <rect x="0" y="-26" width="40" height="40" rx="9" fill="none" stroke="#00D4FF" stroke-width="3"/>
    <path d="M6 -6 L20 -20 L34 -6" fill="none" stroke="#00D4FF" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
    <text x="58" y="6" font-size="30" font-weight="bold" letter-spacing="4" fill="#EBEBF5">GARAGE AI</text>
  </g>

  <!-- headline -->
  <text x="80" y="290" font-size="94" font-weight="bold" fill="#FFFFFF">Your Garage.</text>
  <text x="80" y="392" font-size="94" font-weight="bold" fill="url(#accent)">Europe's AI Engine.</text>

  <!-- subtitle -->
  <text x="80" y="470" font-size="34" fill="#C7C7CC">Sweden's sovereign AI infrastructure —</text>
  <text x="80" y="514" font-size="34" fill="#C7C7CC">solar + EV batteries + local inference.</text>

  <!-- footer -->
  <text x="80" y="582" font-size="26" font-weight="bold" fill="#00FF88">garageai.eu</text>
  <text x="308" y="582" font-size="26" fill="#8E8E93">Open Source · MIT · Made in Sweden</text>
</svg>`;

const out = fileURLToPath(new URL('../public/og-image.png', import.meta.url));
await sharp(Buffer.from(svg)).png().toFile(out);
console.log('Wrote', out);
