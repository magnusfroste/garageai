import { motion } from 'framer-motion';
import CopyCommand from './CopyCommand.jsx';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.12 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 },
};

// GarageAI Bridge: the one program that connects a garage. All copy comes from
// src/content/landing/bridge.md.
const Bridge = ({
  title,
  subtitle,
  description,
  command,
  commandNote,
  commands = [],
  flow = [],
  features = [],
  nextTitle,
  next = [],
  ctaText,
  ctaUrl,
  secondaryText,
  secondaryUrl,
} = {}) => (
  <motion.section
    id="bridge"
    className="py-20 px-4 max-w-6xl mx-auto"
    variants={containerVariants}
    initial="hidden"
    whileInView="visible"
    viewport={{ once: true }}
  >
    <motion.h2 variants={itemVariants} className="apple-heading-1 mb-4 text-center gradient-text-cyan">
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

    {(command || commands.length > 0) && (
      <motion.div variants={itemVariants} className="max-w-3xl mx-auto mb-10 min-w-0">
        {command && <CopyCommand command={command} note={commandNote} color="var(--color-primary)" />}
        {commands.map((c) => (
          <CopyCommand key={c.label} label={c.label} command={c.command} note={c.note} color="var(--color-primary)" />
        ))}
        {commands.length > 0 && commandNote && (
          <p className="mt-2 text-xs" style={{ color: 'var(--color-text-muted)' }}>{commandNote}</p>
        )}
      </motion.div>
    )}

    {flow.length > 0 && (
      <motion.ol variants={itemVariants} className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-14">
        {flow.map((step, i) => (
          <li key={i} className="flex gap-3 p-4 rounded-xl min-w-0" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
            <span
              className="shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-sm font-black"
              style={{ background: 'var(--color-primary)', color: '#000' }}
            >
              {i + 1}
            </span>
            <div className="min-w-0">
              <h3 className="font-bold mb-1" style={{ color: 'var(--color-text-primary)' }}>{step.title}</h3>
              <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{step.text}</p>
            </div>
          </li>
        ))}
      </motion.ol>
    )}

    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
      {features.map((f, i) => (
        <motion.div
          key={i}
          variants={itemVariants}
          className="apple-card min-w-0"
          style={{ border: `1px solid ${f.color || 'rgba(255,255,255,0.08)'}` }}
        >
          <div className="text-3xl mb-3">{f.icon}</div>
          <h3 className="font-bold mb-2" style={{ color: f.color || 'var(--color-text-primary)' }}>{f.title}</h3>
          <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{f.text}</p>
        </motion.div>
      ))}
    </div>

    {next.length > 0 && (
      <motion.div variants={itemVariants} className="max-w-3xl mx-auto mb-12 text-center">
        <h3 className="font-bold mb-3" style={{ color: 'var(--color-text-primary)' }}>{nextTitle}</h3>
        <ul className="flex flex-wrap justify-center gap-2">
          {next.map((item, i) => (
            <li
              key={i}
              className="text-sm px-3 py-1 rounded-full"
              style={{ background: 'rgba(255,255,255,0.06)', color: 'var(--color-text-secondary)' }}
            >
              {item}
            </li>
          ))}
        </ul>
      </motion.div>
    )}

    {ctaUrl && (
      <motion.div variants={itemVariants} className="text-center">
        <motion.a
          href={ctaUrl}
          className="apple-button-primary inline-block no-underline"
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
        >
          {ctaText}
        </motion.a>
        {secondaryUrl && (
          <p className="mt-4 text-sm">
            <a href={secondaryUrl} className="underline" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--color-text-secondary)' }}>
              {secondaryText}
            </a>
          </p>
        )}
      </motion.div>
    )}
  </motion.section>
);

export default Bridge;
