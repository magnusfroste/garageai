import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const landing = defineCollection({
  loader: glob({
    pattern: '*.md',
    base: './src/content/landing',
  }),
  // Each section carries its own structured data in frontmatter (waves,
  // layers, useCases, scenarios, phases, buttons, …). Only the shared fields
  // are validated; section-specific fields pass through to entry.data.
  schema: z
    .object({
      section: z.string(),
      title: z.string(),
      subtitle: z.string().optional(),
      description: z.string().optional(),
      secondaryText: z.string().optional(),
      icon: z.string().optional(),
      order: z.number().default(0),
    })
    .passthrough(),
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
    faqs: z
      .array(
        z.object({
          question: z.string(),
          answer: z.string(),
        })
      )
      .default([]),
  }),
});

export const collections = { landing, faq };
