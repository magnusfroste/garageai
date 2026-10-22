import type { APIRoute } from 'astro';
import { getEntry } from 'astro:content';
import llmsSummary from '../../public/llms.txt?raw';
import securityPage from '../legal/security.md?raw';

// llms-full.txt: public/llms.txt, every FAQ question and answer, and the security page for garage
// owners, built from the same markdown the site renders, so AI assistants can read it all in one
// request. The security page's own headings move one level down to sit under its section.
export const GET: APIRoute = async () => {
  const summary = llmsSummary.trim();
  const faq = await getEntry('faq', 'general');
  const qa = (faq?.data.faqs ?? []).map((f) => `### ${f.question}\n\n${f.answer}`).join('\n\n');
  const security = securityPage.trim().replace(/^## /gm, '### ');
  const body =
    `${summary}\n\n## Frequently asked questions\n\n${qa}\n\n` +
    `## Is it safe to share from my computer? (https://www.garageai.eu/security)\n\n${security}\n`;
  return new Response(body, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
};
