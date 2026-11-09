import type { APIRoute } from 'astro';
import { getEntry } from 'astro:content';
import llmsSummary from '../../public/llms.txt?raw';
import securityPage from '../legal/security.md?raw';
import bridgePage from '../seo/bridge.md?raw';
import apiPage from '../seo/api.md?raw';
import myMachinePage from '../seo/my-machine.md?raw';

// llms-full.txt: public/llms.txt, every FAQ question and answer, and the security page for garage
// owners, built from the same markdown the site renders, so AI assistants can read it all in one
// request. The security page's own headings move one level down to sit under its section.
export const GET: APIRoute = async () => {
  const summary = llmsSummary.trim();
  const faq = await getEntry('faq', 'general');
  const qa = (faq?.data.faqs ?? []).map((f) => `### ${f.question}\n\n${f.answer}`).join('\n\n');
  const security = securityPage.trim().replace(/^## /gm, '### ');
  // The landing pages: their title and lead from the frontmatter, then their body.
  const page = (raw: string) => {
    const m = raw.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
    const fm = m ? m[1] : '';
    const field = (k: string) => (fm.match(new RegExp(`^${k}: "(.*)"$`, 'm')) || [])[1] || '';
    const body = (m ? m[2] : raw).trim().replace(/^## /gm, '### ');
    return `## ${field('h1')} (https://www.garageai.eu${field('path')})\n\n${field('lead')}\n\n${body}`;
  };
  const pages = [bridgePage, apiPage, myMachinePage].map(page).join('\n\n');
  // The hardware setup guides, from the same entry as the "Pick your setup" cards.
  const setupsEntry = await getEntry('landing', 'setups');
  const setups = ((setupsEntry?.data as any)?.setups ?? [])
    .filter((s: any) => s.url)
    .map((s: any) => {
      const models = (s.models ?? []).map((m: any) => `- ${m.memory}: ${m.model}`).join('\n');
      const steps = (s.steps ?? []).map((x: string, i: number) => `${i + 1}. ${x}`).join('\n');
      const faqs = (s.faqs ?? []).map((f: any) => `**${f.question}** ${f.answer}`).join('\n\n');
      return `## ${s.h1} (https://www.garageai.eu${s.url})\n\n${s.lead}\n\nRuntime: ${s.runtime}.\n\n${s.modelsTitle}:\n${models}\n\nSteps:\n${steps}\n\n${faqs}`;
    })
    .join('\n\n');
  const body =
    `${summary}\n\n## Frequently asked questions\n\n${qa}\n\n` +
    `${pages}\n\n` +
    `${setups}\n\n` +
    `## Is it safe to share from my computer? (https://www.garageai.eu/security)\n\n${security}\n`;
  return new Response(body, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
};
