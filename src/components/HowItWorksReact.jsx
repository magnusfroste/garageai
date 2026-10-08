import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.15 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 },
};

const HowItWorks = ({
  title,
  subtitle,
  description,
  tracks = [],
  whyTitle,
  whyDescription,
  whyItems = [],
}) => {
  return (
    <>
      <motion.section
        id="how-it-works"
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
        <motion.p variants={itemVariants} className="apple-body mb-14 text-center max-w-2xl mx-auto">
          {description}
        </motion.p>

        <div className="grid lg:grid-cols-2 gap-8">
          {(tracks || []).map((track) => (
            <motion.div
              key={track.id}
              id={track.id}
              variants={itemVariants}
              className="apple-card flex flex-col"
              style={{ border: `1px solid ${track.color}` }}
            >
              <div className="flex items-center gap-3 mb-2">
                <span className="text-3xl">{track.icon}</span>
                <h3 className="text-2xl font-black" style={{ color: track.color }}>{track.title}</h3>
              </div>
              <p className="text-sm mb-6" style={{ color: 'var(--color-text-muted)' }}>{track.tagline}</p>

              <ol className="space-y-5 flex-1">
                {(track.steps || []).map((step, i) => (
                  <li key={i} className="flex gap-4">
                    <span
                      className="shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-sm font-black"
                      style={{ background: track.color, color: '#000' }}
                    >
                      {i + 1}
                    </span>
                    <div>
                      <h4 className="font-bold mb-1" style={{ color: 'var(--color-text-primary)' }}>{step.title}</h4>
                      <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{step.text}</p>
                    </div>
                  </li>
                ))}
              </ol>

              {track.helpUrl && (
                <p className="mt-6 text-sm">
                  <a href={track.helpUrl} className="underline" style={{ color: 'var(--color-text-secondary)' }}>
                    {track.helpText}
                  </a>
                </p>
              )}

              {track.ctaUrl && (
                <div className="mt-8">
                  <motion.a
                    href={track.ctaUrl}
                    className="apple-button-primary inline-block no-underline"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    {track.ctaText}
                  </motion.a>
                  {track.noteUrl && (
                    <p className="mt-3 text-sm">
                      <a href={track.noteUrl} className="underline" style={{ color: 'var(--color-text-secondary)' }}>
                        {track.noteText}
                      </a>
                    </p>
                  )}
                  {(track.noteDetail || track.noteLinkUrl) && (
                    <p className="mt-1 text-xs" style={{ color: 'var(--color-text-muted)' }}>
                      {track.noteDetail}
                      {track.noteLinkUrl && (
                        <>
                          {track.noteDetail ? ' ' : ''}
                          <a href={track.noteLinkUrl} className="underline" target="_blank" rel="noopener noreferrer">
                            {track.noteLinkText}
                          </a>
                        </>
                      )}
                    </p>
                  )}
                </div>
              )}
            </motion.div>
          ))}
        </div>
      </motion.section>

      {(whyItems || []).length > 0 && (
        <motion.section
          id="why"
          className="py-16 px-4 max-w-6xl mx-auto"
          variants={containerVariants}
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true }}
        >
          <motion.h2 variants={itemVariants} className="apple-heading-1 mb-4 text-center">
            {whyTitle}
          </motion.h2>
          <motion.p variants={itemVariants} className="apple-body mb-12 text-center max-w-2xl mx-auto">
            {whyDescription}
          </motion.p>
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {whyItems.map((item, i) => (
              <motion.div
                key={i}
                variants={itemVariants}
                className="p-6 rounded-xl"
                style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}
                whileHover={{ y: -4 }}
              >
                <div className="text-3xl mb-3">{item.icon}</div>
                <h3 className="font-bold mb-2" style={{ color: item.color }}>{item.title}</h3>
                <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>{item.text}</p>
              </motion.div>
            ))}
          </div>
        </motion.section>
      )}
    </>
  );
};

export default HowItWorks;
