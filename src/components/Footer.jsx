import { motion } from 'framer-motion';

const Footer = () => {
  return (
    <motion.footer
      className="py-12 px-4 border-t mt-20 text-center text-sm"
      style={{ borderColor: 'rgba(0,255,136,0.15)', color: 'var(--color-text-muted)' }}
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
    >
      <div className="max-w-6xl mx-auto">
        <motion.p
          className="mb-2 font-semibold"
          style={{ color: 'var(--color-text-secondary)' }}
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.2 }}
        >
          🌍 Garage AI — Sovereign AI infrastructure for Europe. Built in Sweden.
        </motion.p>
        <motion.p
          className="mb-6 text-xs max-w-xl mx-auto"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.3 }}
        >
          Decentralize compute. Reclaim sovereignty. Power it with the sun.
        </motion.p>

        <motion.div
          className="flex gap-6 justify-center mb-6 flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.4 }}
        >
          <a href="https://github.com/magnusfroste/garageai/blob/main/docs/GET_STARTED.md" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">🚀 Get Started</a>
          <a href="https://github.com/magnusfroste/garageai/blob/main/docs/GARAGE_AI_KNOWLEDGE_BASE.md" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">📖 Docs</a>
          <a href="https://github.com/magnusfroste/garageai" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">🐙 GitHub</a>
          <a href="https://github.com/magnusfroste/garageai/discussions" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">💬 Discussions</a>
          <a href="mailto:powerup@garageai.eu" className="hover:text-green-400 transition">📧 powerup@garageai.eu</a>
        </motion.div>

        <motion.div
          className="mb-4 text-xs"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ delay: 0.5 }}
        >
          <span style={{ color: 'var(--color-text-muted)' }}>Technical partners: </span>
          <a href="https://www.autoversio.com" target="_blank" rel="noopener noreferrer" className="hover:text-cyan-400 transition" style={{ color: 'var(--color-accent)' }}>Autoversio</a>
          <span style={{ color: 'var(--color-text-muted)' }}> & </span>
          <a href="https://www.liteit.se" target="_blank" rel="noopener noreferrer" className="hover:text-cyan-400 transition" style={{ color: 'var(--color-accent)' }}>Liteit</a>
        </motion.div>

        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ delay: 0.6 }}
          className="text-xs"
        >
          © 2026 Garage AI · MIT License · Made in Sweden 🇸🇪 · Built for Europe 🇪🇺
          <br />
          <span style={{ opacity: 0.5 }}>
            Built with{' '}
            <a
              href="https://claude.ai/code"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-orange-400 transition"
              style={{ color: 'rgba(255,160,80,0.8)' }}
            >
              Claude Code
            </a>
            {' '}— our AI co-worker, architect, and fact-checker.
          </span>
          <br />
          <span style={{ opacity: 0.35 }}>
            The €100M funding announcement was published on April 1, 2026. The prototype, the vision, and the community are entirely real.
          </span>
        </motion.p>
      </div>
    </motion.footer>
  );
};

export default Footer;
