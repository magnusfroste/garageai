import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

const Community = ({
  title = 'Open Source & Community',
  subtitle,
  description,
  quotes = [],
  stats = [],
  buildTitle = '🛠️ Build & Contribute',
  buildItems = [],
  spreadTitle = '🌍 Spread the Vision',
  spreadItems = [],
  joinTitle = 'Join Early',
  joinText,
  joinButtons = [],
} = {}) => {
  const quoteList = quotes || [];
  const statList = stats || [];

  return (
    <motion.section
      id="open-source"
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
      {subtitle && (
        <motion.p variants={itemVariants} className="apple-heading-2 mb-4 text-center" style={{ color: 'var(--color-text-secondary)' }}>
          {subtitle}
        </motion.p>
      )}
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

      {/* Quotes (only rendered when real, attributable quotes are provided) */}
      {quoteList.length > 0 && (
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
      )}

      {/* Ways to participate */}
      <motion.div variants={itemVariants} className="grid md:grid-cols-2 gap-8 mb-12">
        <div className="fact-box p-7">
          <h4 className="font-black text-lg mb-4" style={{ color: 'var(--color-primary)' }}>
            {buildTitle}
          </h4>
          <ul className="space-y-2 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
            {(buildItems || []).map((item, i) => (
              <li key={i} className="flex gap-2"><span style={{ color: 'var(--color-primary)' }}>→</span>{item}</li>
            ))}
          </ul>
        </div>
        <div className="fact-box p-7">
          <h4 className="font-black text-lg mb-4" style={{ color: 'var(--color-accent)' }}>
            {spreadTitle}
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
            {joinTitle}
          </h3>
          <p className="text-sm mb-6" style={{ color: 'var(--color-text-secondary)' }}>
            {joinText}
          </p>
          <div className="flex gap-4 justify-center flex-wrap">
            {(joinButtons || []).map((button) => (
              <motion.a
                key={button.url}
                href={button.url}
                {...(button.url.startsWith('https://github.com') ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
                className={`${button.variant === 'primary' ? 'apple-button-primary' : 'apple-button-secondary'} inline-block no-underline`}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                {button.text}
              </motion.a>
            ))}
          </div>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Community;
