import { motion } from 'framer-motion';

const Sponsors = () => {
  return (
    <motion.section
      id="partners"
      className="py-8 px-4 max-w-4xl mx-auto"
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true }}
    >
      <div className="apple-card text-center">
        <p className="text-xs mb-3" style={{ color: 'var(--color-text-muted)' }}>
          Technical partners building the foundation
        </p>
        <p className="text-sm">
          <a href="https://www.autoversio.com" target="_blank" rel="noopener noreferrer" className="font-semibold hover:text-primary transition" style={{ color: 'var(--color-accent)' }}>Autoversio</a>
          <span style={{ color: 'var(--color-text-muted)' }}> — Secure AI infrastructure &nbsp;·&nbsp; </span>
          <a href="https://www.liteit.se" target="_blank" rel="noopener noreferrer" className="font-semibold hover:text-primary transition" style={{ color: 'var(--color-accent)' }}>Liteit</a>
          <span style={{ color: 'var(--color-text-muted)' }}> — AI-driven development</span>
        </p>
      </div>
    </motion.section>
  );
};

export default Sponsors;
