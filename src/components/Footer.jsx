import { motion } from 'framer-motion';

const Footer = () => {
  return (
    <motion.footer
      className="py-12 px-4 border-t border-green-500/20 mt-20 text-center text-gray-500 text-sm"
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
    >
      <div className="max-w-6xl mx-auto">
        <motion.p
          className="mb-4"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          🌍 Garage AI – Bygg nästa lager ovanpå fiber & internet. Suveränitet för alla.
        </motion.p>
        <motion.div
          className="flex gap-6 justify-center mb-4 flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          <a href="https://github.com/magnusfroste/garageai/blob/main/docs/GET_STARTED.md" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">🚀 Komma Igång</a>
          <a href="https://github.com/magnusfroste/garageai" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">📖 Teknisk Docs</a>
          <a href="https://github.com/magnusfroste/garageai" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">🐙 GitHub</a>
          <a href="https://github.com/magnusfroste/garageai/discussions" target="_blank" rel="noopener noreferrer" className="hover:text-green-400 transition">💬 Diskussioner</a>
        </motion.div>
        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.6 }}
        >
          © 2025 Garage AI. Sponsrat av Autoversio & Liteit. Made in Sweden 🇸🇪
        </motion.p>
      </div>
    </motion.footer>
  );
};

export default Footer;
