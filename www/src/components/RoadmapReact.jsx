import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

const Roadmap = ({
  title,
  description,
  phases = [],
  whyActTitle = 'Why Join Early?',
  whyActItems = [],
  ctaText = '🚀 Get Started',
  ctaUrl = 'https://app.garageai.eu/auth',
} = {}) => {
  const phaseList = phases || [];
  const whyList = whyActItems || [];
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
        {title}
      </motion.h2>
      <motion.p variants={itemVariants} className="apple-body mb-16 text-center max-w-2xl mx-auto">
        {description}
      </motion.p>

      <motion.div variants={itemVariants} className="grid lg:grid-cols-3 gap-8 mb-16">
        {(phaseList || []).map((phase, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="apple-card relative"
            style={{
              border: phase.status === 'current'
                ? `1px solid ${phase.color}`
                : phase.status === 'done'
                  ? '1px solid rgba(255,255,255,0.06)'
                  : undefined,
              opacity: phase.status === 'done' ? 0.65 : 1,
            }}
          >
            <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
              <span
                className="text-xs font-black px-3 py-1 rounded-full"
                style={{
                  background: phase.badgeColor,
                  color: phase.status === 'done' ? 'rgba(255,255,255,0.8)' : '#000',
                }}
              >
                {phase.status === 'current' ? '● ' : phase.status === 'done' ? '✓ ' : ''}{phase.badge}
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
        <h3 className="apple-heading-2 mb-8 text-center">{whyActTitle}</h3>
        <div className="grid md:grid-cols-3 gap-8 mb-10">
          {(whyList || []).map((item, i) => (
            <div key={i} className="text-center">
              <div className="text-4xl mb-3">{item.icon}</div>
              <h4 className="font-bold mb-2" style={{ color: 'var(--color-primary)' }}>{item.title}</h4>
              <p className="text-sm" style={{ color: 'var(--color-text-secondary)' }}>{item.text}</p>
            </div>
          ))}
        </div>

        <div className="text-center">
          <motion.button
            onClick={() => { window.location.href = ctaUrl; }}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            {ctaText}
          </motion.button>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Roadmap;
