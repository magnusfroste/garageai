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

// "Pricing & Economics" section. All copy comes from
// src/content/landing/token-economy.md.
const TokenEconomy = ({
  title,
  subtitle,
  description,
  buyOptions = [],
  buyerNotes = [],
  operatorPoints = [],
  poolTitle,
  poolDescription,
  poolSteps = [],
  poolNote,
  marketTitle,
  marketDescription,
  marketStats = [],
  whyAgents = [],
  footnotes = [],
} = {}) => {
  return (
    <motion.section
      id="pricing"
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

      <motion.div variants={itemVariants} className="grid lg:grid-cols-2 gap-8 mb-10">
        {/* Buyers */}
        <div className="apple-card">
          <h3 className="text-xl font-black mb-6" style={{ color: 'var(--color-accent)' }}>⚡ For buyers</h3>
          <div className="grid sm:grid-cols-2 gap-4 mb-6">
            {(buyOptions || []).map((opt, i) => (
              <div key={i} className="p-5 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid ${opt.color}` }}>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-2xl">{opt.icon}</span>
                  <h4 className="font-black" style={{ color: opt.color }}>{opt.title}</h4>
                </div>
                <p className="text-xs mb-3" style={{ color: 'var(--color-text-muted)' }}>{opt.tag}</p>
                <ul className="space-y-2">
                  {(opt.points || []).map((pt, j) => (
                    <li key={j} className="text-xs flex gap-2" style={{ color: 'var(--color-text-secondary)' }}>
                      <span style={{ color: opt.color }}>✓</span>{pt}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
          <ul className="space-y-2">
            {(buyerNotes || []).map((note, i) => (
              <li key={i} className="text-sm flex gap-2" style={{ color: 'var(--color-text-secondary)' }}>
                <span style={{ color: 'var(--color-accent)' }}>→</span>{note}
              </li>
            ))}
          </ul>
        </div>

        {/* Operators */}
        <div className="apple-card">
          <h3 className="text-xl font-black mb-6" style={{ color: 'var(--color-primary)' }}>🏠 For operators</h3>
          <ul className="space-y-4">
            {(operatorPoints || []).map((pt, i) => (
              <li key={i} className="flex gap-3 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
                <span style={{ color: 'var(--color-primary)' }}>✓</span>{pt}
              </li>
            ))}
          </ul>
        </div>
      </motion.div>

      {/* How the pool works */}
      {(poolSteps || []).length > 0 && (
        <motion.div variants={itemVariants} className="apple-card mb-10">
          <h3 className="apple-heading-2 mb-2 text-center">{poolTitle}</h3>
          {poolDescription && (
            <p className="text-sm mb-8 text-center max-w-2xl mx-auto" style={{ color: 'var(--color-text-secondary)' }}>{poolDescription}</p>
          )}
          <ol className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {poolSteps.map((step, i) => (
              <li key={i} className="p-5 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
                <div className="text-xs font-black mb-2" style={{ color: 'var(--color-primary)', opacity: 0.7 }}>
                  {String(i + 1).padStart(2, '0')}
                </div>
                <h4 className="font-bold text-sm mb-2" style={{ color: 'var(--color-text-primary)' }}>{step.title}</h4>
                <p className="text-xs leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{step.text}</p>
              </li>
            ))}
          </ol>
          {poolNote && (
            <p className="text-xs mt-6 text-center" style={{ color: 'var(--color-text-muted)' }}>{poolNote}</p>
          )}
        </motion.div>
      )}

      {/* Market context (third-party estimates, clearly labelled) */}
      {(marketStats || []).length > 0 && (
        <motion.div variants={itemVariants} className="apple-card">
          <h3 className="apple-heading-2 mb-2 text-center">{marketTitle}</h3>
          <p className="text-sm mb-8 text-center" style={{ color: 'var(--color-text-muted)' }}>{marketDescription}</p>
          <div className="flex flex-col sm:flex-row justify-center gap-4 mb-8">
            {marketStats.map((stat, i) => (
              <div key={i} className="text-center p-5 rounded-xl sm:w-1/3" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
                <div className="text-3xl font-black mb-1" style={{ color: stat.color }}>
                  <AnimatedCounter value={stat.num} prefix={stat.prefix} suffix={stat.suffix} />
                </div>
                <div className="text-xs font-semibold mb-1" style={{ color: 'var(--color-text-primary)' }}>{stat.label}</div>
                <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{stat.sub}</div>
              </div>
            ))}
          </div>
          {(whyAgents || []).length > 0 && (
            <ul className="space-y-2 text-sm max-w-3xl mx-auto mb-6" style={{ color: 'var(--color-text-secondary)' }}>
              {whyAgents.map((item, i) => (
                <li key={i} className="flex gap-2"><span style={{ color: 'var(--color-warning)' }}>→</span>{item}</li>
              ))}
            </ul>
          )}
          <div className="pt-4 border-t border-slate-800 space-y-1">
            {(footnotes || []).map((note, i) => (
              <p key={i} className="text-xs" style={{ color: 'var(--color-text-muted)', opacity: 0.6 }}>{note}</p>
            ))}
          </div>
        </motion.div>
      )}
    </motion.section>
  );
};

export default TokenEconomy;
