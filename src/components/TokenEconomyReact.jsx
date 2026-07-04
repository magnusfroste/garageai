import { motion } from 'framer-motion';
import AnimatedCounter from './AnimatedCounter';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 }
};

const defaultScenarios = [
  {
    label: '2026 Pilot',
    flag: '🇸🇪',
    nodes: 10000,
    directJobs: 1000,
    supportJobs: 150,
    passivePerNode: 30000,
    avgSalary: 333000,
    color: 'var(--color-primary)',
  },
  {
    label: '2027 Scale',
    flag: '🇸🇪',
    nodes: 100000,
    directJobs: 3000,
    supportJobs: 1500,
    passivePerNode: 30000,
    avgSalary: 333000,
    color: 'var(--color-warning)',
  },
  {
    label: '2030 Full Potential',
    flag: '🇸🇪',
    nodes: 1700000,
    directJobs: 40000,
    supportJobs: 17000,
    passivePerNode: 30000,
    avgSalary: 333000,
    color: 'var(--color-accent)',
  },
];

const defaultFootnotes = [
  '¹ MarketsandMarkets: AI Inference Market Report 2025–2030',
  '² Fortune Business Insights: Agentic AI Market Report 2026–2034',
  '³ Gartner, cited in Landbase: 39 Agentic AI Statistics 2026',
  '⁴ Industry forecast cited in VAST Data / Nvidia AI infrastructure analysis 2026',
  '⁵ MIT Technology Review, May 2025: "Inferencing drives 80–90% of all AI compute"',
  '⁶ Anthropic Series G press release, February 2026',
  '⁷ Seeking Alpha / Anthropic internal forecast, raised December 2025',
  '⁸ Sacra / Anthropic revenue tracking, March 2026',
  '⁹ Anthropic API pricing, platform.claude.com, 2026',
  '¹⁰ McKinsey: "Accelerating Europe\'s AI adoption — the role of sovereign AI capabilities", 2025',
  '¹¹ Nvidia at GTC 2026: "Agents, inference and the new token economics"',
];

const TokenEconomy = ({
  title = 'The Inference Economy',
  description = "AI inference is the fastest-growing compute market in history. The third wave — autonomous agents — is accelerating demand to levels that no single company can satisfy. That's where your garage comes in.",
  scenarios,
  footnotes,
} = {}) => {
  const scenarioList = scenarios || defaultScenarios;
  const footnoteList = footnotes || defaultFootnotes;
  return (
    <motion.section
      id="token-economy"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2 variants={itemVariants} className="apple-heading-1 mb-4 text-center">
        {title}
      </motion.h2>
      <motion.p variants={itemVariants} className="apple-body mb-16 text-center max-w-2xl mx-auto">
        {description}
      </motion.p>

      {/* Market size */}
      <motion.div variants={itemVariants} className="apple-card mb-8">
        <h3 className="apple-heading-2 mb-8 text-center">The Market That Never Sleeps</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-10">
          {[
            { num: 106,   prefix: '$', suffix: 'B', label: 'AI inference market 2025', sub: 'current baseline', color: 'var(--color-warning)' },
            { num: 255,   prefix: '$', suffix: 'B', label: 'projected by 2030¹', sub: '+140% in 5 years', color: 'var(--color-primary)' },
            { num: 19.2,  prefix: '',  suffix: '%', label: 'annual growth rate¹', sub: 'CAGR 2025–2030', color: 'var(--color-accent)' },
            { num: 1000,  prefix: '',  suffix: '×', label: 'agentic demand multiplier⁴', sub: 'forecast from agents alone', color: 'rgba(255,59,48,0.9)' },
          ].map((stat, i) => (
            <div key={i} className="text-center p-5 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
              <div className="text-3xl font-black mb-1" style={{ color: stat.color }}>
                <AnimatedCounter value={stat.num} prefix={stat.prefix} suffix={stat.suffix} />
              </div>
              <div className="text-xs font-semibold mb-1" style={{ color: 'var(--color-text-primary)' }}>{stat.label}</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{stat.sub}</div>
            </div>
          ))}
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          <div className="p-5 rounded-xl" style={{ background: 'rgba(255,214,10,0.06)', border: '1px solid rgba(255,214,10,0.2)' }}>
            <p className="font-bold mb-3" style={{ color: 'var(--color-warning)' }}>🤖 Why Agents Change Everything</p>
            <p className="text-sm mb-4" style={{ color: 'var(--color-text-secondary)' }}>
              A human chatting with an AI: maybe 20 inference calls per session.
              An autonomous agent completing a task: potentially <strong>thousands of calls</strong> —
              planning, tool use, reflection, verification, retry loops. The demand per user
              doesn't grow linearly. It compounds.
            </p>
            <ul className="space-y-2 text-xs" style={{ color: 'var(--color-text-secondary)' }}>
              <li>→ Gartner: 40% of enterprise apps will have embedded AI agents by end of 2026, up from 5% in Sept 2025³</li>
              <li>→ IDC: AI agent use cases grow 10× by 2027</li>
              <li>→ Inference now drives 80–90% of all AI compute usage⁵</li>
              <li>→ VAST Data declares 2026 "The Year of AI Inference"</li>
            </ul>
          </div>
          <div className="p-5 rounded-xl" style={{ background: 'rgba(0,255,136,0.05)', border: '1px solid rgba(0,255,136,0.2)' }}>
            <p className="font-bold mb-3" style={{ color: 'var(--color-primary)' }}>📈 The Agentic AI Market Alone</p>
            <ul className="space-y-3 text-sm">
              {[
                ['Agentic AI market 2025', '$7.3B²'],
                ['Agentic AI market 2026', '$9.1B²'],
                ['Agentic AI market 2034', '$139B²'],
                ['CAGR 2026–2034', '40.5%²'],
                ['Enterprise AI spend growth', '+31.9% p.a. to 2029'],
                ['Total AI investment by 2029', '$1.3 trillion'],
              ].map(([label, value], i) => (
                <li key={i} className="flex justify-between border-b border-slate-800 pb-2">
                  <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
                  <span className="font-bold" style={{ color: 'var(--color-primary)' }}>{value}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </motion.div>

      {/* Anthropic as proof of market */}
      <motion.div variants={itemVariants} className="apple-card mb-8">
        <h3 className="apple-heading-2 mb-4 text-center">
          Proof of Market: Anthropic's Token Economy
        </h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          Anthropic doesn't sell software licences. It sells <strong>tokens consumed in real-time</strong>
          as inference runs — billions of them, every hour, at scale. This is the market
          Garage AI participates in from the bottom up.
        </p>

        <div className="grid md:grid-cols-2 gap-8 mb-8">
          <div>
            <p className="font-semibold mb-4" style={{ color: 'var(--color-accent)' }}>📊 Anthropic by the Numbers</p>
            <ul className="space-y-3 text-sm">
              {[
                ['Valuation (Feb 2026, Series G)⁶', '$380 billion'],
                ['Annualised revenue (Mar 2026)⁸', '$19 billion'],
                ['YoY revenue growth⁸', '+1,167%'],
                ['Revenue Dec 2024 → Dec 2025⁸', '$1B → $9B annualised'],
                ['2026 revenue forecast⁷', '$18B'],
                ['Business customers⁸', '300,000+'],
                ['Enterprise share of revenue⁸', '~80%'],
              ].map(([label, value], i) => (
                <li key={i} className="flex justify-between border-b border-slate-800 pb-2">
                  <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
                  <span className="font-bold text-right" style={{ color: 'var(--color-accent)' }}>{value}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="p-6 rounded-xl" style={{ background: 'rgba(0,122,255,0.06)', border: '1px solid rgba(0,122,255,0.2)' }}>
            <p className="font-semibold mb-4" style={{ color: 'var(--color-accent)' }}>🔄 How the Token Revenue Model Works</p>
            <p className="text-sm mb-4" style={{ color: 'var(--color-text-secondary)' }}>
              Every time a developer or enterprise calls the Claude API, tokens are counted and billed
              in real-time — the moment inference runs. Anthropic's entire revenue is, in essence,
              a <strong>live metered stream of compute consumption</strong>.
            </p>
            <div className="space-y-2 text-xs mb-4" style={{ color: 'var(--color-text-secondary)' }}>
              <div className="flex justify-between border-b border-slate-800 pb-1">
                <span>Claude Haiku 4.5</span>
                <span style={{ color: 'var(--color-primary)' }}>$1 / $5 per M tokens (in/out)⁹</span>
              </div>
              <div className="flex justify-between border-b border-slate-800 pb-1">
                <span>Claude Sonnet 4.6</span>
                <span style={{ color: 'var(--color-primary)' }}>$3 / $15 per M tokens⁹</span>
              </div>
              <div className="flex justify-between border-b border-slate-800 pb-1">
                <span>Claude Opus 4.6</span>
                <span style={{ color: 'var(--color-primary)' }}>$5 / $25 per M tokens⁹</span>
              </div>
            </div>
            <p className="text-xs italic" style={{ color: 'var(--color-text-muted)' }}>
              A single agentic workflow consuming 1M output tokens = $5–$25 charged instantly.
              At $19B revenue/year and growing, this is what token demand looks like at scale.
            </p>
          </div>
        </div>

        <div className="p-5 rounded-xl text-center" style={{ background: 'rgba(255,214,10,0.06)', border: '1px solid rgba(255,214,10,0.2)' }}>
          <p className="text-sm font-semibold" style={{ color: 'var(--color-warning)' }}>
            Anthropic's $380B valuation is built on selling tokens consumed in real-time at scale.
            Garage AI is building the distributed infrastructure to serve that same growing demand —
            locally, sustainably, and with value flowing back to node operators, not hyperscalers.¹¹
          </p>
        </div>
      </motion.div>

      {/* Jobs + Passive Income */}
      <motion.div variants={itemVariants} className="apple-card">
        <h3 className="apple-heading-2 mb-4 text-center">Jobs & Passive Income</h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          The inference economy creates two kinds of opportunity: direct employment across
          the installation and support chain, and passive income for anyone who runs a node.
        </p>

        <div className="grid md:grid-cols-2 gap-8 mb-8">
          {/* Jobs */}
          <div>
            <p className="font-bold mb-4" style={{ color: 'var(--color-primary)' }}>
              👷 Jobs Created — The Full Stack
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--color-text-muted)' }}>
              Europe needs €15–20B/year in AI compute infrastructure investment through 2030.¹⁰
              Distributed infrastructure creates jobs where centralised data centres don't —
              spread across every municipality, not concentrated in one industrial park.
            </p>
            <ul className="space-y-3">
              {[
                { role: 'Solar + battery installation', est: 'High volume, existing trades', icon: '☀️' },
                { role: 'EV charger & V2H installation', est: 'Certified electricians, growing fast', icon: '⚡' },
                { role: 'Node setup & local IT support', est: 'New entry-level tech roles', icon: '🖥️' },
                { role: 'Network & fiber technicians', est: 'Ongoing expansion + maintenance', icon: '🌐' },
                { role: 'Regional cluster management', est: 'Skilled sysadmin/DevOps', icon: '🏗️' },
                { role: 'Community energy coordinators', est: 'New role — V2G & grid balancing', icon: '🔋' },
                { role: 'Local AI integration consultants', est: 'Helping SMEs adopt local AI', icon: '🤝' },
              ].map((job, i) => (
                <li key={i} className="flex items-start gap-3 text-sm border-b border-slate-800 pb-2">
                  <span>{job.icon}</span>
                  <div className="flex-1">
                    <span className="font-medium" style={{ color: 'var(--color-text-primary)' }}>{job.role}</span>
                    <span className="text-xs block" style={{ color: 'var(--color-text-muted)' }}>{job.est}</span>
                  </div>
                </li>
              ))}
            </ul>
          </div>

          {/* Passive income */}
          <div>
            <p className="font-bold mb-4" style={{ color: 'var(--color-accent)' }}>
              💰 Passive Income — The Node Operator Opportunity
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--color-text-muted)' }}>
              When your home's AI needs are covered and spare compute capacity exists,
              that capacity can serve the token market. The model is established:
              Render Network (RNDR) already pays GPU owners in tokens for idle compute.
              Nvidia calls inference monetisation the next major revenue layer.¹¹
            </p>
            <div className="space-y-4">
              {[
                {
                  phase: 'Today (prototype)',
                  desc: 'Run your own inference, cut cloud API costs. 1M tokens locally ≈ $3–$25 saved vs Anthropic/OpenAI pricing.',
                  color: 'var(--color-primary)',
                },
                {
                  phase: 'Near term (neighbourhood sharing)',
                  desc: 'Sell spare capacity to local businesses via the shared infrastructure model. Predictable recurring revenue from regional SMEs.',
                  color: 'var(--color-warning)',
                },
                {
                  phase: 'Scaled (token market)',
                  desc: 'As the inference market grows to $255B by 2030, distributed node operators participate directly in a live, metered token economy — the same economic model Anthropic is worth $380B on.',
                  color: 'var(--color-accent)',
                },
              ].map((phase, i) => (
                <div key={i} className="p-4 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid ${phase.color}30` }}>
                  <p className="font-bold text-xs mb-1" style={{ color: phase.color }}>{phase.phase}</p>
                  <p className="text-xs" style={{ color: 'var(--color-text-secondary)' }}>{phase.desc}</p>
                </div>
              ))}
            </div>

            <div className="mt-4 p-4 rounded-xl text-center" style={{ background: 'rgba(0,255,136,0.06)', border: '1px solid rgba(0,255,136,0.2)' }}>
              <p className="text-xs font-semibold" style={{ color: 'var(--color-primary)' }}>
                The market for tokens is real, growing, and open. Anyone with a garage can participate.
              </p>
            </div>
          </div>
        </div>

        {/* Sweden jobs calculator */}
        <div className="mt-8 pt-8 border-t border-slate-800">
          <h4 className="font-black text-lg mb-2 text-center" style={{ color: 'var(--color-primary)' }}>
            🇸🇪 Sweden — Jobs & Income Calculator
          </h4>
          <p className="text-xs text-center mb-6" style={{ color: 'var(--color-text-muted)' }}>
            Based on avg. node passive income SEK 30,000/yr · avg. net salary SEK 333,000/yr ·
            installation 2 workers × 2 days per node.
            Sources: shareai.now GPU income data 2026; SCB average salary Sweden 2025.
          </p>

          <div className="grid md:grid-cols-3 gap-4">
            {(scenarioList || []).map((s) => {
              const passiveTotal = s.nodes * s.passivePerNode;
              const passiveJobs = Math.round(passiveTotal / s.avgSalary);
              const totalJobs = s.directJobs + s.supportJobs + passiveJobs;
              return (
                <div
                  key={s.label}
                  className="p-5 rounded-2xl"
                  style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid ${s.color}30` }}
                >
                  <div className="flex items-center gap-2 mb-4">
                    <span className="text-xl">{s.flag}</span>
                    <span className="font-black text-sm" style={{ color: s.color }}>{s.label}</span>
                  </div>

                  <div className="space-y-2 text-xs mb-4">
                    <div className="flex justify-between border-b border-slate-800 pb-1">
                      <span style={{ color: 'var(--color-text-muted)' }}>Active nodes</span>
                      <span className="font-bold" style={{ color: s.color }}>
                        {s.nodes.toLocaleString()}
                      </span>
                    </div>
                    <div className="flex justify-between border-b border-slate-800 pb-1">
                      <span style={{ color: 'var(--color-text-muted)' }}>Direct install + support jobs</span>
                      <span className="font-bold" style={{ color: 'var(--color-text-primary)' }}>
                        {(s.directJobs + s.supportJobs).toLocaleString()}
                      </span>
                    </div>
                    <div className="flex justify-between border-b border-slate-800 pb-1">
                      <span style={{ color: 'var(--color-text-muted)' }}>Passive income generated</span>
                      <span className="font-bold" style={{ color: 'var(--color-text-primary)' }}>
                        SEK {(passiveTotal / 1e6).toFixed(0)}M/yr
                      </span>
                    </div>
                    <div className="flex justify-between border-b border-slate-800 pb-1">
                      <span style={{ color: 'var(--color-text-muted)' }}>…in salary equivalents</span>
                      <span className="font-bold" style={{ color: 'var(--color-text-primary)' }}>
                        {passiveJobs.toLocaleString()} FTE
                      </span>
                    </div>
                  </div>

                  <div className="p-3 rounded-xl text-center" style={{ background: `${s.color}12` }}>
                    <div className="text-xs mb-1" style={{ color: s.color }}>Total job equivalent</div>
                    <div className="text-2xl font-black" style={{ color: s.color }}>
                      {totalJobs.toLocaleString()}
                    </div>
                    <div className="text-xs mt-1" style={{ color: 'var(--color-text-muted)' }}>
                      direct + passive income streams
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          <p className="text-xs text-center mt-4" style={{ color: 'var(--color-text-muted)', opacity: 0.55 }}>
            Passive income modelled at SEK 2,500/month per node (conservative mid-range based on consumer GPU
            rental data). "Salary equivalent" = passive income ÷ Swedish average net salary.
            Direct jobs include solar/battery installers, electricians, IT support, and cluster management.
          </p>
        </div>

        {/* Footnotes */}
        <div className="pt-4 border-t border-slate-800 space-y-1">
          {(footnoteList || []).map((note, i) => (
            <p key={i} className="text-xs" style={{ color: 'var(--color-text-muted)', opacity: 0.5 }}>{note}</p>
          ))}
        </div>
      </motion.div>

    </motion.section>
  );
};

export default TokenEconomy;
