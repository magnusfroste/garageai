import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 }
};

// "For Business" section. All copy comes from src/content/landing/b2b.md.
const B2B = ({
  title,
  subtitle,
  description,
  offerings = [],
  useCases = [],
  privacyNote,
  ctaText,
  ctaButtonText,
  ctaUrl,
} = {}) => {
  return (
    <motion.section
      id="for-business"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2 variants={itemVariants} className="apple-heading-1 mb-4 text-center">
        {title}
      </motion.h2>
      {subtitle && (
        <motion.p variants={itemVariants} className="apple-heading-2 mb-4 text-center" style={{ color: 'var(--color-text-secondary)' }}>
          {subtitle}
        </motion.p>
      )}
      <motion.p variants={itemVariants} className="apple-body mb-14 text-center max-w-2xl mx-auto">
        {description}
      </motion.p>

      <motion.div variants={itemVariants} className="grid md:grid-cols-3 gap-6 mb-10">
        {(offerings || []).map((item, i) => (
          <div key={i} className="p-6 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid ${item.color}` }}>
            <div className="text-3xl mb-3">{item.icon}</div>
            <h3 className="font-bold mb-3" style={{ color: item.color }}>{item.title}</h3>
            <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{item.text}</p>
          </div>
        ))}
      </motion.div>

      {(useCases || []).length > 0 && (
        <motion.div variants={itemVariants} className="apple-card mb-8">
          <h3 className="apple-heading-2 mb-8 text-center">Good Fits Today</h3>
          <div className="grid md:grid-cols-3 gap-4">
            {useCases.map((uc, i) => (
              <div
                key={i}
                className="p-5 rounded-xl"
                style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-2xl">{uc.icon}</span>
                  <span className="font-bold text-sm" style={{ color: 'var(--color-accent)' }}>{uc.sector}</span>
                </div>
                <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{uc.example}</p>
              </div>
            ))}
          </div>
        </motion.div>
      )}

      {privacyNote && (
        <motion.div
          variants={itemVariants}
          className="p-5 rounded-xl mb-10 text-center"
          style={{ background: 'rgba(255,214,10,0.06)', border: '1px solid rgba(255,214,10,0.2)' }}
        >
          <p className="text-sm max-w-3xl mx-auto" style={{ color: 'var(--color-text-secondary)' }}>
            <strong style={{ color: 'var(--color-warning)' }}>About data: </strong>{privacyNote}
          </p>
        </motion.div>
      )}

      {ctaUrl && (
        <motion.div variants={itemVariants} className="text-center">
          <p className="text-sm mb-4" style={{ color: 'var(--color-text-secondary)' }}>{ctaText}</p>
          <motion.a
            href={ctaUrl}
            className="apple-button-primary inline-block no-underline"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            {ctaButtonText}
          </motion.a>
        </motion.div>
      )}
    </motion.section>
  );
};

export default B2B;
