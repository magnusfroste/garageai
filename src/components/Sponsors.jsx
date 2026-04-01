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

      {/* Technical inspiration: Nosana */}
      <motion.a
        href="https://nosana.com"
        target="_blank"
        rel="noopener noreferrer"
        className="flex flex-col md:flex-row items-center gap-4 text-center md:text-left p-6 rounded-2xl mb-6 transition-all"
        style={{
          background: 'rgba(153,69,255,0.05)',
          border: '1px solid rgba(153,69,255,0.25)',
          textDecoration: 'none',
        }}
        whileHover={{ scale: 1.02, background: 'rgba(153,69,255,0.08)' }}
      >
        <div className="shrink-0 text-center">
          <div className="text-3xl mb-1">🏗️</div>
        </div>
        <div className="flex-1">
          <div className="flex flex-wrap items-center gap-2 mb-1 justify-center md:justify-start">
            <span className="text-base font-black" style={{ color: 'rgba(153,69,255,0.95)' }}>Nosana</span>
            <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'rgba(153,69,255,0.15)', color: 'rgba(153,69,255,0.9)' }}>Technical Inspiration</span>
            <span className="text-xs" style={{ color: 'var(--color-text-muted)' }}>Amsterdam, Netherlands</span>
          </div>
          <p className="text-xs leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
            Nosana pioneered the Podman-in-Docker GPU isolation pattern that powers Garage AI nodes.
            Their open-source architecture proved that decentralized GPU networks work at scale:
            2M+ deployments, 826,000+ AI jobs, 2,000+ nodes globally.
            Garage AI's <code style={{ color: 'rgba(153,69,255,0.8)', fontSize: '0.7rem' }}>garage_start.sh</code> is directly inspired by their node setup.
          </p>
        </div>
      </motion.a>

      {/* Partners */}
      <div className="grid md:grid-cols-2 gap-4 mb-6">
        {[
          {
            name: 'Autoversio',
            tagline: 'Private AI for conscious businesses',
            url: 'https://www.autoversio.com',
            color: 'var(--color-primary)',
          },
          {
            name: 'Liteit',
            tagline: 'Leadership for agentic co-workers',
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
