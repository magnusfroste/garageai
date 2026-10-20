import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 }
};

const MARK = {
  yes: { symbol: '●', color: 'var(--color-primary)' },
  partial: { symbol: '◐', color: 'var(--color-warning)' },
  no: { symbol: '○', color: 'var(--color-text-secondary)' },
  'n/a': { symbol: '–', color: 'var(--color-text-secondary)' },
};

// Comparison against categories of alternatives. All copy comes from
// src/content/landing/positioning.md.
const Positioning = ({
  title,
  subtitle,
  description,
  claim,
  columns = [],
  rows = [],
  legend = {},
  footnote,
} = {}) => {
  return (
    <motion.section
      id="why-garageai"
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
      <motion.p variants={itemVariants} className="apple-body mb-10 text-center max-w-3xl mx-auto">
        {description}
      </motion.p>

      {claim && (
        <motion.blockquote
          variants={itemVariants}
          className="p-6 rounded-xl mb-12 text-center max-w-3xl mx-auto"
          style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid var(--color-primary)' }}
        >
          <p className="text-lg font-semibold leading-relaxed" style={{ color: 'var(--color-text)' }}>{claim}</p>
        </motion.blockquote>
      )}

      <motion.div variants={itemVariants} className="apple-card">
        <div style={{ overflowX: 'auto', WebkitOverflowScrolling: 'touch' }}>
        <table className="w-full text-sm" style={{ borderCollapse: 'collapse', minWidth: '640px' }}>
          <thead>
            <tr>
              <th className="text-left p-3" style={{ color: 'var(--color-text-secondary)', fontWeight: 500 }}></th>
              {columns.map((c, i) => (
                <th
                  key={i}
                  className="p-3 text-center"
                  style={{ color: i === 0 ? 'var(--color-primary)' : 'var(--color-text-secondary)', fontWeight: i === 0 ? 700 : 500 }}
                >
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} style={{ borderTop: '1px solid rgba(255,255,255,0.07)' }}>
                <td className="p-3 align-top">
                  <div style={{ color: 'var(--color-text)' }}>{r.feature}</div>
                  {r.note && (
                    <div className="text-xs mt-1" style={{ color: 'var(--color-text-secondary)' }}>{r.note}</div>
                  )}
                </td>
                {(r.values || []).map((v, j) => {
                  const m = MARK[v] || MARK['n/a'];
                  return (
                    <td
                      key={j}
                      className="p-3 text-center align-top text-xl"
                      style={{ color: m.color, background: j === 0 ? 'rgba(255,255,255,0.03)' : 'transparent' }}
                      aria-label={legend[v] || v}
                      title={legend[v] || v}
                    >
                      {m.symbol}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
        </div>
        <p className="text-xs text-center mt-2 md:hidden" style={{ color: 'var(--color-text-secondary)' }}>Swipe sideways to see all columns</p>
        <div className="flex flex-wrap gap-4 justify-center mt-4 text-xs" style={{ color: 'var(--color-text-secondary)' }}>
          {Object.entries(MARK).map(([k, m]) => (
            <span key={k}><span style={{ color: m.color }}>{m.symbol}</span> {legend[k] || k}</span>
          ))}
        </div>
      </motion.div>

      {footnote && (
        <motion.p variants={itemVariants} className="text-xs text-center mt-6 max-w-3xl mx-auto" style={{ color: 'var(--color-text-secondary)' }}>
          {footnote}
        </motion.p>
      )}
    </motion.section>
  );
};

export default Positioning;
