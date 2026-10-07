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
    // /partners is unlisted: reachable by link, not in the sitemap.
    sitemap({ filter: (page) => !page.includes('/partners') }),
  ],
  vite: {
    ssr: {
      noExternal: ['framer-motion'],
    },
  },
});
