import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.15 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 },
};

// Recommended runtime, model and steps per hardware profile. All copy comes from
// src/content/landing/setups.md.
const Setups = ({
  title,
  subtitle,
  description,
  setups = [],
  benefitsTitle,
  benefits = [],
  ctaText,
  ctaUrl,
  note,
} = {}) => (
  <motion.section
    id="pick-your-setup"
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
    <motion.p variants={itemVariants} className="apple-body mb-14 text-center max-w-3xl mx-auto">
      {description}
    </motion.p>

    <div className="grid md:grid-cols-2 gap-6 mb-14">
      {setups.map((s) => (
        <motion.div
          key={s.id}
          id={s.id}
          variants={itemVariants}
          className="apple-card flex flex-col"
          style={{ border: `1px solid ${s.color}` }}
        >
          <div className="flex items-center gap-3 mb-1 flex-wrap">
            <span className="text-3xl">{s.icon}</span>
            <h3 className="text-xl font-black" style={{ color: s.color }}>{s.title}</h3>
            {s.badge && (
              <span className="text-xs px-2 py-1 rounded-full" style={{ background: 'rgba(255,255,255,0.06)', color: 'var(--color-text-secondary)' }}>
                {s.badge}
              </span>
            )}
          </div>
          <p className="text-sm mb-4" style={{ color: 'var(--color-text-secondary)' }}>{s.tagline}</p>

          <p className="text-sm mb-3">
            <span style={{ color: 'var(--color-text-secondary)' }}>Runtime: </span>
            <strong>{s.runtime}</strong>
          </p>

          <div className="mb-5">
            <div className="text-xs mb-2" style={{ color: 'var(--color-text-secondary)' }}>{s.modelsTitle}</div>
            <table className="w-full text-sm" style={{ borderCollapse: 'collapse' }}>
              <tbody>
                {(s.models || []).map((m, i) => (
                  <tr key={i} style={{ borderTop: '1px solid rgba(255,255,255,0.07)' }}>
                    <td className="py-1.5 pr-3 align-top" style={{ color: 'var(--color-text-secondary)' }}>{m.memory}</td>
                    <td className="py-1.5 align-top">{m.model}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <ol className="space-y-3 text-sm flex-1">
            {(s.steps || []).map((step, i) => (
              <li key={i} className="flex gap-3">
                <span className="font-bold shrink-0" style={{ color: s.color }}>{i + 1}</span>
                <span style={{ color: 'var(--color-text-secondary)' }}>{step}</span>
              </li>
            ))}
          </ol>
        </motion.div>
      ))}
    </div>

    {benefits.length > 0 && (
      <motion.div variants={itemVariants} className="mb-10">
        <h3 className="apple-heading-2 mb-6 text-center">{benefitsTitle}</h3>
        <div className="grid md:grid-cols-3 gap-6">
          {benefits.map((b, i) => (
            <div key={i} className="apple-card">
              <div className="text-3xl mb-3">{b.icon}</div>
              <h4 className="font-bold mb-2">{b.title}</h4>
              <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{b.text}</p>
            </div>
          ))}
        </div>
      </motion.div>
    )}

    {ctaUrl && (
      <motion.div variants={itemVariants} className="text-center">
        <a href={ctaUrl} className="apple-button-primary inline-block no-underline">{ctaText}</a>
        {note && <p className="text-xs mt-4 max-w-2xl mx-auto" style={{ color: 'var(--color-text-secondary)' }}>{note}</p>}
      </motion.div>
    )}
  </motion.section>
);

export default Setups;
