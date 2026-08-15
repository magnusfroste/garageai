import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import mdx from '@astrojs/mdx';
import sitemap from '@astrojs/sitemap';

// Tailwind v3 is processed via postcss.config.js (the @astrojs/tailwind
// integration is not compatible with Astro 6).
export default defineConfig({
  site: 'https://www.garageai.eu',
  integrations: [
    react(),
    mdx(),
    sitemap(),
  ],
  vite: {
    ssr: {
      noExternal: ['framer-motion'],
    },
  },
});
