import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const landing = defineCollection({
  loader: glob({
    pattern: '*.md',
    base: './src/content/landing',
  }),
  schema: z.object({
    section: z.string(),
    title: z.string(),
    subtitle: z.string().optional(),
    description: z.string().optional(),
    secondaryText: z.string().optional(),
    icon: z.string().optional(),
    order: z.number().default(0),
  }),
});

const faq = defineCollection({
  loader: glob({
    pattern: '**/*.md',
    base: './src/content/faq',
  }),
  schema: z.object({
    section: z.string(),
    title: z.string(),
    description: z.string().optional(),
    order: z.number().default(0),
  }),
});

export const collections = { landing, faq };
