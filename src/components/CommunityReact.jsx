import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

const quotes = [
  {
    quote: 'Running a private inference server at home felt niche two years ago. Now it feels like the obvious future. Garage AI is the structure this movement was missing.',
    author: 'Early node operator',
    location: 'Stockholm, Sweden',
    avatar: '🚀',
  },
  {
    quote: 'The parallels to solar are striking. Same technology adoption curve, same decentralization pressure, same sovereignty argument. We\'re early.',
    author: 'Energy systems researcher',
    location: 'Gothenburg, Sweden',
    avatar: '☀️',
  },
  {
    quote: 'When autonomous agents need hundreds of inference calls per task, local infrastructure stops being optional — it becomes the only rational choice.',
    author: 'AI infrastructure engineer',
    location: 'Malmö, Sweden',
    avatar: '🤖',
  },
];

const communityStats = [
  { icon: '🖥️', value: 'Growing', label: 'Node Operators' },
  { icon: '🌍', value: '🇸🇪→🇪🇺', label: 'Sweden First' },
  { icon: '🔓', value: 'MIT', label: 'Open Source License' },
  { icon: '⚡', value: 'Live', label: 'Prototype Running' },
];

const Community = ({
  title = 'The Community',
  description = "Builders, researchers, energy enthusiasts, and AI practitioners — united around a shared conviction that Europe's AI future should be local, sovereign, and community-owned.",
  quotes: quotesProp,
  stats,
  buildItems = ['Set up a garage node (docs on GitHub)', 'Contribute to the open-source inference cluster', 'Test V2G and solar integration', 'Document your setup for others'],
  spreadItems = ['Share this initiative with your municipality', 'Talk to local energy companies about V2G', 'Connect with EU policy makers', 'Tell your neighbours what\'s possible'],
} = {}) => {
  const quoteList = quotesProp || quotes;
  const statList = stats || communityStats;

  return (
    <motion.section
      id="community"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="apple-heading-1 mb-4 text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        {title}
      </motion.h2>
      <motion.p
        variants={itemVariants}
        className="apple-body mb-12 text-center max-w-2xl mx-auto"
      >
        {description}
      </motion.p>

      {/* Stats */}
      <motion.div variants={itemVariants} className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-16">
        {(statList || []).map((stat, i) => (
          <motion.div
            key={i}
            variants={itemVariants}
            className="text-center p-6 rounded-xl"
            style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}
            whileHover={{ scale: 1.05 }}
          >
            <div className="text-3xl mb-2">{stat.icon}</div>
            <div className="text-2xl font-black mb-1" style={{ color: 'var(--color-primary)' }}>{stat.value}</div>
            <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{stat.label}</div>
          </motion.div>
        ))}
      </motion.div>

      {/* Quotes */}
      <motion.div variants={itemVariants} className="grid md:grid-cols-3 gap-6 mb-16">
        {(quoteList || []).map((q, i) => (
          <motion.div
            key={i}
            variants={itemVariants}
            className="fact-box p-6"
            whileHover={{ y: -4 }}
          >
            <div className="text-4xl mb-4 text-center">{q.avatar}</div>
            <blockquote className="text-sm italic mb-4 leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
              "{q.quote}"
            </blockquote>
            <div className="text-right">
              <div className="text-sm font-semibold" style={{ color: 'var(--color-primary)' }}>{q.author}</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{q.location}</div>
            </div>
          </motion.div>
        ))}
      </motion.div>

      {/* Ways to participate */}
      <motion.div variants={itemVariants} className="grid md:grid-cols-2 gap-8 mb-12">
        <div className="fact-box p-7">
          <h4 className="font-black text-lg mb-4" style={{ color: 'var(--color-primary)' }}>
            🛠️ Build & Contribute
          </h4>
          <ul className="space-y-2 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
            {(buildItems || []).map((item, i) => (
              <li key={i} className="flex gap-2"><span style={{ color: 'var(--color-primary)' }}>→</span>{item}</li>
            ))}
          </ul>
        </div>
        <div className="fact-box p-7">
          <h4 className="font-black text-lg mb-4" style={{ color: 'var(--color-accent)' }}>
            🌍 Spread the Vision
          </h4>
          <ul className="space-y-2 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
            {(spreadItems || []).map((item, i) => (
              <li key={i} className="flex gap-2"><span style={{ color: 'var(--color-accent)' }}>→</span>{item}</li>
            ))}
          </ul>
        </div>
      </motion.div>

      {/* Join CTA */}
      <motion.div variants={itemVariants}>
        <div className="text-center p-8 rounded-2xl" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
          <h3 className="text-xl font-black mb-4" style={{ color: 'var(--color-primary)' }}>
            Join the Movement
          </h3>
          <p className="text-sm mb-6" style={{ color: 'var(--color-text-secondary)' }}>
            500+ builders are already running nodes. The prototype is live. Europe's sovereign AI future starts in a garage.
          </p>
          <div className="flex gap-4 justify-center flex-wrap">
            <motion.button
              onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
              className="apple-button-primary"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              💬 GitHub Discussions
            </motion.button>
            <motion.button
              onClick={() => window.open('https://github.com/magnusfroste/garageai', '_blank')}
              className="apple-button-secondary"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              🔓 View on GitHub
            </motion.button>
          </div>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Community;
