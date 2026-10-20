import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

// "Technology & Security" section. All copy comes from
// src/content/landing/infrastructure.md.
const Infrastructure = ({
  title,
  subtitle,
  description,
  layers = [],
  securityColumns = [],
  ctaText,
  ctaUrl,
} = {}) => {
  return (
    <motion.section
      id="technology"
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
      <motion.p variants={itemVariants} className="apple-body mb-16 text-center max-w-2xl mx-auto">
        {description}
      </motion.p>

      {/* Stack diagram */}
      <motion.div variants={itemVariants} className="mb-16">
        <div className="space-y-3 max-w-4xl mx-auto">
          {(layers || []).map((layer, i) => (
            <div
              key={i}
              className="flex items-start gap-5 p-5 rounded-xl transition-all"
              style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
            >
              <div className="text-center shrink-0 w-12">
                <div className="text-3xl mb-1">{layer.icon}</div>
                <div className="text-xs font-black" style={{ color: layer.color, opacity: 0.5 }}>{layer.step}</div>
              </div>
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-2 mb-1">
                  <h3 className="font-bold" style={{ color: layer.color }}>{layer.title}</h3>
                  <span className="text-xs px-2 py-0.5 rounded" style={{ background: 'rgba(255,255,255,0.06)', color: layer.color }}>
                    {layer.subtitle}
                  </span>
                </div>
                <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
                  {layer.details}
                </p>
              </div>
            </div>
          ))}
        </div>
      </motion.div>

      {/* Honest security summary */}
      {(securityColumns || []).length > 0 && (
        <motion.div variants={itemVariants} className="apple-card mb-8">
          <h3 className="apple-heading-2 mb-8 text-center">Security, Honestly</h3>
          <div className="grid md:grid-cols-3 gap-6">
            {securityColumns.map((col, i) => (
              <div key={i} className="p-5 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
                <div className="text-3xl mb-2">{col.icon}</div>
                <h4 className="font-bold mb-3" style={{ color: col.color }}>{col.title}</h4>
                <ul className="space-y-2">
                  {(col.points || []).map((pt, j) => (
                    <li key={j} className="text-sm flex gap-2" style={{ color: 'var(--color-text-secondary)' }}>
                      <span style={{ color: col.color }}>→</span>
                      <span>{pt}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </motion.div>
      )}

      {ctaUrl && (
        <motion.div variants={itemVariants} className="text-center">
          <a
            href={ctaUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm hover:text-cyan-400 transition"
            style={{ color: 'var(--color-text-muted)' }}
          >
            {ctaText} →
          </a>
        </motion.div>
      )}
    </motion.section>
  );
};

export default Infrastructure;
