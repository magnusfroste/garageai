import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

const phases = [
  {
    flag: '🇸🇪',
    period: '2025',
    title: 'Sweden: Foundation',
    badge: 'DONE',
    badgeColor: 'rgba(120,120,120,0.9)',
    status: 'done',
    items: [
      'Distributed inference cluster built',
      'Multi-node repo published on GitHub (open source)',
      'Local AI on consumer hardware validated',
      'Proof-of-concept for garage nodes confirmed',
      'Security & performance benchmarking completed',
    ],
    color: 'rgba(180,180,180,0.8)',
  },
  {
    flag: '🇸🇪',
    period: '2026',
    title: 'Sweden: Pilot',
    badge: 'NOW',
    badgeColor: 'var(--color-primary)',
    status: 'current',
    items: [
      'Swedish pilot launched — first live nodes',
      'Solar + V2G energy integration',
      'Neighbourhood sharing protocol live',
      'Community governance structure established',
      '10,000 jobs initiative formally announced',
    ],
    color: 'var(--color-primary)',
  },
  {
    flag: '🇪🇺',
    period: '2027',
    title: 'Europe: Accelerate',
    badge: 'NEXT',
    badgeColor: 'var(--color-accent)',
    status: 'upcoming',
    items: [
      'EU expansion — Germany, France, Netherlands first',
      'Sweden model as reference architecture for all EU states',
      '1M+ European garage nodes target',
      'Pan-European inference mesh live',
      'EU regulatory framework alignment',
      'The rest of the world follows',
    ],
    color: 'var(--color-accent)',
  },
];

const Roadmap = () => {
  return (
    <motion.section
      id="roadmap"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2 variants={itemVariants} className="apple-heading-1 mb-4 text-center gradient-text-cyan">
        Sweden First. Europe Next.
      </motion.h2>
      <motion.p variants={itemVariants} className="apple-body mb-16 text-center max-w-2xl mx-auto">
        Sweden's innovation culture, fiber infrastructure, EV leadership and technical competence
        make it the natural starting point for Europe's sovereign AI revolution.
      </motion.p>

      <motion.div variants={itemVariants} className="grid lg:grid-cols-3 gap-8 mb-16">
        {phases.map((phase, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="apple-card relative"
            style={phase.status === 'current' ? { border: `1px solid ${phase.color}` } : {}}
          >
            <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
              <span
                className="text-xs font-black px-3 py-1 rounded-full text-black"
                style={{ background: phase.badgeColor }}
              >
                {phase.status === 'current' ? '🔴 ' : ''}{phase.badge}
              </span>
            </div>

            <div className="text-center mb-6">
              <div className="text-4xl mb-2">{phase.flag}</div>
              <h3 className="text-3xl font-black mb-1" style={{ color: phase.color }}>
                {phase.period}
              </h3>
              <h4 className="font-semibold text-sm" style={{ color: 'var(--color-text-secondary)' }}>
                {phase.title}
              </h4>
            </div>

            <ul className="space-y-3">
              {phase.items.map((item, i) => (
                <li key={i} className="flex items-start gap-3">
                  <span style={{ color: phase.color }}>•</span>
                  <span className="text-sm" style={{ color: 'var(--color-text-secondary)' }}>{item}</span>
                </li>
              ))}
            </ul>
          </motion.div>
        ))}
      </motion.div>

      {/* Why act now */}
      <motion.div variants={itemVariants} className="apple-card max-w-4xl mx-auto">
        <h3 className="apple-heading-2 mb-8 text-center">Why Act Now?</h3>
        <div className="grid md:grid-cols-3 gap-8 mb-10">
          {[
            { icon: '🏆', title: 'Be an Early Builder', text: 'First movers shape the protocol, the community governance, and the defaults that millions will use.' },
            { icon: '🇸🇪', title: 'Sweden Sets the Standard', text: 'What Sweden builds today becomes the reference architecture for every EU member state that follows.' },
            { icon: '🌍', title: 'Sovereignty Can\'t Wait', text: 'EU AI dependency on US hyperscalers grows by the day. Every quarter delayed is more lock-in to reverse.' },
          ].map((item, i) => (
            <div key={i} className="text-center">
              <div className="text-4xl mb-3">{item.icon}</div>
              <h4 className="font-bold mb-2" style={{ color: 'var(--color-primary)' }}>{item.title}</h4>
              <p className="text-sm" style={{ color: 'var(--color-text-secondary)' }}>{item.text}</p>
            </div>
          ))}
        </div>

        <div className="text-center">
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            🚀 Join the Community
          </motion.button>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Roadmap;
