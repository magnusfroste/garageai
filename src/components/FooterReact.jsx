import { motion } from 'framer-motion';

const defaultTagline = '🌍 Garage AI — Sovereign AI infrastructure for Europe. Built in Sweden.';
const defaultSubtagline = 'Decentralize compute. Reclaim sovereignty. Power it with the sun.';
const defaultLinks = [
  { icon: '🚀', text: 'Get Started', url: 'https://github.com/magnusfroste/garageai/blob/main/docs/GET_STARTED.md' },
  { icon: '📖', text: 'Docs', url: 'https://github.com/magnusfroste/garageai/blob/main/docs/GARAGE_AI_KNOWLEDGE_BASE.md' },
  { icon: '🐙', text: 'GitHub', url: 'https://github.com/magnusfroste/garageai' },
  { icon: '💬', text: 'Discussions', url: 'https://github.com/magnusfroste/garageai/discussions' },
  { icon: '📧', text: 'powerup@garageai.eu', url: 'mailto:powerup@garageai.eu' },
];
const defaultPartners = [
  { name: 'Autoversio', url: 'https://www.autoversio.com' },
  { name: 'Liteit', url: 'https://www.liteit.se' },
];
const defaultCopyright = '© 2026 Garage AI · MIT License · Made in Sweden 🇸🇪 · Built for Europe 🇪🇺';

const Footer = ({
  tagline = defaultTagline,
  subtagline = defaultSubtagline,
  links = defaultLinks,
  partners = defaultPartners,
  copyright = defaultCopyright,
}) => {
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
          {tagline}
        </motion.p>
        <motion.p
          className="mb-6 text-xs max-w-xl mx-auto"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.3 }}
        >
          {subtagline}
        </motion.p>

        <motion.div
          className="flex gap-6 justify-center mb-6 flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.4 }}
        >
          {(links || []).map((link) => {
            const isMailto = (link.url || '').startsWith('mailto:');
            return (
              <a
                key={link.url}
                href={link.url}
                {...(isMailto ? {} : { target: '_blank', rel: 'noopener noreferrer' })}
                className="hover:text-green-400 transition"
              >
                {link.icon} {link.text}
              </a>
            );
          })}
        </motion.div>

        <motion.div
          className="mb-4 text-xs"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ delay: 0.5 }}
        >
          <span style={{ color: 'var(--color-text-muted)' }}>Technical partners: </span>
          {(partners || []).map((partner, index) => (
            <span key={partner.url}>
              <a
                href={partner.url}
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-cyan-400 transition"
                style={{ color: 'var(--color-accent)' }}
              >
                {partner.name}
              </a>
              {index < (partners || []).length - 1 && (
                <span style={{ color: 'var(--color-text-muted)' }}> & </span>
              )}
            </span>
          ))}
        </motion.div>

        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ delay: 0.6 }}
          className="text-xs"
        >
          {copyright}
          <br />
          <span style={{ opacity: 0.5 }}>
            AI inference powered by{' '}
            <a
              href="https://kilo.code"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-orange-400 transition"
              style={{ color: 'rgba(255,160,80,0.8)' }}
            >
              Kilo Code
            </a>
            {' '}— produced in our own garage on solar energy! ☀️🏠
          </span>
        </motion.p>
      </div>
    </motion.footer>
  );
};

export default Footer;
