import type { APIRoute } from 'astro';
import { getEntry } from 'astro:content';
import llmsSummary from '../../public/llms.txt?raw';

// llms-full.txt: public/llms.txt plus every FAQ question and answer, built from the same
// markdown the page renders, so AI assistants can read the whole site in one request.
export const GET: APIRoute = async () => {
  const summary = llmsSummary.trim();
  const faq = await getEntry('faq', 'general');
  const qa = (faq?.data.faqs ?? []).map((f) => `### ${f.question}\n\n${f.answer}`).join('\n\n');
  const body = `${summary}\n\n## Frequently asked questions\n\n${qa}\n`;
  return new Response(body, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
};
