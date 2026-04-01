import { motion } from 'framer-motion';

const Sponsors = () => {
  return (
    <motion.section
      id="partners"
      className="py-12 px-4 max-w-4xl mx-auto"
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6 }}
    >
      <p className="text-xs text-center mb-8 uppercase tracking-widest" style={{ color: 'var(--color-text-muted)' }}>
        Partners & Acknowledgements
      </p>

      {/* Partners */}
      <div className="grid md:grid-cols-2 gap-4 mb-6">
        {[
          {
            name: 'Autoversio',
            tagline: 'Privat AI för medvetna företag',
            url: 'https://www.autoversio.com',
            color: 'var(--color-primary)',
          },
          {
            name: 'Liteit',
            tagline: 'Ledarskap för agentiska medarbetare',
            url: 'https://www.liteit.se',
            color: 'var(--color-accent)',
          },
        ].map((p) => (
          <motion.a
            key={p.name}
            href={p.url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex flex-col items-center text-center p-6 rounded-2xl transition-all"
            style={{
              background: 'rgba(255,255,255,0.03)',
              border: `1px solid ${p.color}22`,
              textDecoration: 'none',
            }}
            whileHover={{ scale: 1.03, background: 'rgba(255,255,255,0.05)' }}
          >
            <span className="text-lg font-black mb-1" style={{ color: p.color }}>{p.name}</span>
            <span className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{p.tagline}</span>
          </motion.a>
        ))}
      </div>

      {/* Fake minister thank-you */}
      <motion.div
        className="text-center p-5 rounded-2xl"
        style={{ background: 'rgba(255,214,10,0.04)', border: '1px solid rgba(255,214,10,0.12)' }}
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        transition={{ delay: 0.3 }}
      >
        <p className="text-sm" style={{ color: 'var(--color-text-secondary)' }}>
          A special thank you to{' '}
          <span className="font-bold" style={{ color: 'var(--color-warning)' }}>
            Minister of Digitalisation & AI, Anna Svensson*
          </span>
          {' '}for her unwavering support of Sweden's sovereign AI ambitions
          and for personally championing the Garage AI initiative in Brussels.
        </p>
        <p className="text-xs mt-3" style={{ color: 'var(--color-text-muted)', opacity: 0.5 }}>
          * Fictional. Published April 1, 2026. Any resemblance to actual ministers is purely aspirational.
        </p>
      </motion.div>
    </motion.section>
  );
};

export default Sponsors;
