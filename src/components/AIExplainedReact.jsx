import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

const waves = [
  {
    number: '01',
    era: '1980s',
    title: 'Personal Computing',
    subtitle: 'The First Wave',
    description:
      'Compute power left the mainframe. Bill Gates, Steve Jobs, and a generation of builders put a computer on every desk and in every home. Processing power was no longer the exclusive domain of corporations and governments — it belonged to individuals. The pattern was set: centralized technology decentralizes, and the world changes.',
    icon: '🖥️',
    color: 'rgba(120,120,255,0.9)',
    outcome: 'Every desk got a computer. Computing power became personal.',
  },
  {
    number: '02',
    era: '1990s – 2000s',
    title: 'The Internet',
    subtitle: 'The Second Wave',
    description:
      'Information left the library and became borderless. Anyone could publish, connect, and trade. The infrastructure was centralized — servers, ISPs, data centers — but the value flowed to billions of individuals. E-commerce, social media, and open knowledge followed. A new layer on top of the previous wave.',
    icon: '🌐',
    color: 'var(--color-accent)',
    outcome: 'Every home got connected. Knowledge and commerce became global.',
  },
  {
    number: '03',
    era: '2022 – 2024',
    title: 'Generative AI',
    subtitle: 'The Third Wave',
    description:
      'Creation was democratized. Anyone could generate text, images, code and analysis at scale. But this wave re-centralized what computing had decentralized: the most capable AI ran exclusively on hyperscaler infrastructure — OpenAI, Anthropic, Google. Powerful, transformative, and controlled by a handful of companies.',
    icon: '✨',
    color: 'var(--color-warning)',
    outcome: 'Creation became accessible. But sovereignty moved back to the cloud.',
  },
  {
    number: '04',
    era: '2025 →',
    title: 'Autonomous Agents',
    subtitle: 'The Fourth Wave — Happening Now',
    description:
      'Action is being democratized. AI agents plan, decide, and execute autonomously — making thousands of model calls per task. Demand for inference is exploding at a pace no centralised infrastructure can sustainably serve. The key insight: agents running on private, local models unlock true autonomy. The intelligence that left the mainframe in the 1980s is now leaving the cloud — and coming back to your garage.',
    icon: '🤖',
    color: 'var(--color-primary)',
    outcome: 'Action becomes autonomous. Intelligence comes home.',
    highlight: true,
  },
];

const AIExplained = () => {
  return (
    <motion.section
      id="three-waves"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="apple-heading-1 mb-4 text-center gradient-text-cyan"
      >
        Four Waves. One Direction.
      </motion.h2>
      <motion.p
        variants={itemVariants}
        className="apple-body mb-16 text-center max-w-2xl mx-auto"
      >
        Every wave of technological change has followed the same arc: power concentrates,
        then decentralizes back to individuals. We're living through the fourth —
        and it's bringing intelligence back home.
      </motion.p>

      {/* Three waves timeline */}
      <motion.div variants={itemVariants} className="space-y-6 mb-16">
        {waves.map((wave) => (
          <div
            key={wave.number}
            className="apple-card relative overflow-hidden"
            style={wave.highlight ? { border: `1px solid ${wave.color}`, background: 'rgba(0,255,136,0.04)' } : {}}
          >
            {wave.highlight && (
              <div className="absolute top-4 right-4">
                <span className="text-xs font-black px-3 py-1 rounded-full text-black"
                  style={{ background: wave.color }}>
                  NOW
                </span>
              </div>
            )}
            <div className="flex flex-col md:flex-row gap-6 items-start">
              <div className="text-center md:text-left shrink-0">
                <div className="text-5xl mb-2">{wave.icon}</div>
                <div className="text-5xl font-black" style={{ color: wave.color, opacity: 0.3 }}>{wave.number}</div>
              </div>
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-3 mb-2">
                  <h3 className="text-2xl font-black" style={{ color: wave.color }}>{wave.title}</h3>
                  <span className="text-xs px-2 py-1 rounded border" style={{ color: wave.color, borderColor: wave.color, opacity: 0.7 }}>
                    {wave.subtitle}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{wave.era}</span>
                </div>
                <p className="text-sm mb-4 leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
                  {wave.description}
                </p>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold" style={{ color: wave.color }}>Outcome:</span>
                  <span className="text-xs font-semibold" style={{ color: 'var(--color-text-primary)' }}>{wave.outcome}</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </motion.div>

      {/* OpenClaw & local inference explosion */}
      <motion.div variants={itemVariants} className="apple-card mb-8">
        <h3 className="apple-heading-2 mb-4 text-center">The Catalyst: OpenClaw</h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          <a href="https://openclaw.ai" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--color-primary)', fontWeight: 700 }}>OpenClaw</a>
          {' '}is the viral phenomenon that crystallised the fourth wave —
          an open-source personal AI agent that runs entirely on your own hardware,
          integrating with WhatsApp, Telegram, iMessage and 50+ other services.
          More downloads in its first months than Linux accumulated in 30 years.
          Users say: <em style={{ color: 'var(--color-text-secondary)' }}>"It's running my company."</em>
        </p>

        <div className="grid md:grid-cols-3 gap-6 mb-8">
          {[
            {
              icon: '🔒',
              title: 'Private by Default',
              text: 'Data never leaves your hardware. OpenClaw runs locally — browser control, file operations, shell commands, persistent memory. Exactly what Garage AI infrastructure was built to power.',
              color: 'var(--color-primary)'
            },
            {
              icon: '🤖',
              title: 'Truly Agentic',
              text: 'Manages email, calendar, files and workflows autonomously. One agent, thousands of model calls per day. The inference demand is real — and it needs to be local, fast, and private.',
              color: 'var(--color-accent)'
            },
            {
              icon: '💻',
              title: 'The Mac Mini Moment',
              text: 'A Mac Mini M4 Pro (64GB unified memory) runs Llama 70B at ~15 tokens/sec — entirely locally. The household AI node is already here. The garage is the logical next step.',
              color: 'var(--color-warning)'
            }
          ].map((card, i) => (
            <div key={i} className="p-5 rounded-xl text-center" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
              <div className="text-4xl mb-3">{card.icon}</div>
              <h4 className="font-bold mb-2" style={{ color: card.color }}>{card.title}</h4>
              <p className="text-xs leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{card.text}</p>
            </div>
          ))}
        </div>

        {/* Agents in the Wild */}
        <div className="mb-6">
          <p className="font-semibold text-center mb-5" style={{ color: 'var(--color-text-secondary)' }}>
            Agents in the Wild — Three Examples of the Fourth Wave
          </p>
          <div className="grid md:grid-cols-3 gap-4">
            {[
              {
                name: 'OpenClaw',
                url: 'https://openclaw.ai',
                tag: 'Personal AI agent',
                desc: 'Runs locally on your machine. Controls browser, files, calendar and 50+ integrations via WhatsApp or Telegram. Private by design — built for local inference.',
                color: 'var(--color-primary)',
                icon: '🦾',
              },
              {
                name: 'FlowWink',
                url: 'https://www.flowwink.com',
                tag: 'AI-powered CMS & marketing',
                desc: 'AI agents manage content workflows and marketing operations autonomously — reducing the manual overhead of digital publishing to near zero.',
                color: 'var(--color-warning)',
                icon: '✨',
              },
              {
                name: 'Silicon Soap',
                url: 'https://www.siliconsoap.com',
                tag: 'Multi-agent debate platform',
                desc: '2–4 AI agents (Llama, DeepSeek, Qwen, Mistral) debate any topic autonomously across multiple rounds. A live demonstration of multi-agent coordination at work.',
                color: 'var(--color-accent)',
                icon: '🎭',
              },
            ].map((ex) => (
              <a
                key={ex.name}
                href={ex.url}
                target="_blank"
                rel="noopener noreferrer"
                className="block p-5 rounded-xl transition-all"
                style={{
                  background: 'rgba(255,255,255,0.03)',
                  border: `1px solid ${ex.color}22`,
                  textDecoration: 'none',
                }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-xl">{ex.icon}</span>
                  <span className="font-black text-sm" style={{ color: ex.color }}>{ex.name}</span>
                  <span className="text-xs px-2 py-0.5 rounded-full ml-auto" style={{ background: `${ex.color}18`, color: ex.color }}>{ex.tag}</span>
                </div>
                <p className="text-xs leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{ex.desc}</p>
              </a>
            ))}
          </div>
        </div>

        <div className="p-4 rounded-xl text-center" style={{ background: 'rgba(0,122,255,0.07)', border: '1px solid rgba(0,122,255,0.2)' }}>
          <p className="text-sm" style={{ color: 'var(--color-accent)' }}>
            <strong>These agents are running right now.</strong> Each one generates hundreds of inference calls per session.
            As adoption scales, the demand for local, private, sovereign compute becomes existential — not optional.
            That is what Garage AI is built for.
          </p>
        </div>
      </motion.div>

      {/* The connected home analogy */}
      <motion.div variants={itemVariants} className="apple-card">
        <h3 className="apple-heading-2 mb-4 text-center">Your Home as a Connected Tesla</h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          In the future, every home will function like a connected Tesla — aware of its energy,
          optimized for seasons, and integrated with everything around it.
        </p>
        <div className="grid md:grid-cols-2 gap-8">
          <div>
            <p className="font-semibold mb-4" style={{ color: 'var(--color-primary)' }}>🏠 The Integrated Home Stack</p>
            <ul className="space-y-3">
              {[
                ['Security & cameras', 'Local AI processes, zero cloud'],
                ['Fire & smoke detection', 'Real-time local response'],
                ['Heating & ventilation', 'AI-optimized for summer/winter'],
                ['Garden & irrigation', 'Weather-aware, fully automated'],
                ['Kitchen appliances', 'Connected to same local gateway'],
                ['EV charging', 'Smart scheduling via V2G'],
              ].map(([system, benefit], i) => (
                <li key={i} className="flex justify-between text-sm border-b border-slate-800 pb-2">
                  <span style={{ color: 'var(--color-text-secondary)' }}>{system}</span>
                  <span className="text-xs font-medium" style={{ color: 'var(--color-primary)' }}>{benefit}</span>
                </li>
              ))}
            </ul>
          </div>
          <div>
            <p className="font-semibold mb-4" style={{ color: 'var(--color-accent)' }}>⚠️ The Problem with Cloud Dependency</p>
            <div className="p-5 rounded-xl mb-4" style={{ background: 'rgba(255,59,48,0.07)', border: '1px solid rgba(255,59,48,0.2)' }}>
              <p className="text-sm mb-3" style={{ color: 'var(--color-text-secondary)' }}>
                Systems like Ngenic, smart locks, and home automation hubs are increasingly powerful —
                but they rely on central cloud compute. That sounds exactly like what it is:
              </p>
              <p className="text-sm font-bold text-center" style={{ color: 'rgba(255,59,48,0.9)' }}>
                Completely wrong.
              </p>
            </div>
            <div className="p-5 rounded-xl" style={{ background: 'rgba(0,255,136,0.05)', border: '1px solid rgba(0,255,136,0.2)' }}>
              <p className="text-sm font-semibold mb-2" style={{ color: 'var(--color-primary)' }}>The Garage AI gateway:</p>
              <p className="text-sm" style={{ color: 'var(--color-text-secondary)' }}>
                One local AI node connects all home systems. No data leaves.
                No dependency on distant servers. When the internet goes down,
                your home keeps running. <strong>That is true autonomy.</strong>
              </p>
            </div>
          </div>
        </div>
      </motion.div>

    </motion.section>
  );
};

export default AIExplained;
