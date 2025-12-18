import { motion } from 'framer-motion';

const CTA = () => {
  const showComingSoon = () => {
    alert('🚀 Community-portalen lanseras snart!\n\nI mellan tiden:\n\n📞 Autoversio → autoversio.com\n⚡ Liteit → liteit.se');
  };

  return (
    <motion.section
      className="py-20 px-4 max-w-6xl mx-auto text-center"
      initial={{ opacity: 0, y: 50 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
    >
      <div className="fact-box">
        <motion.h2
          className="text-4xl font-bold mb-6 glow-text"
          style={{ color: 'var(--color-primary)' }}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          Gå med i Gräsrotsrörelsen
        </motion.h2>
        <motion.p
          className="text-lg text-gray-300 mb-8 max-w-2xl mx-auto"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          5000 solpaneler + 500k gaming-riggar + 1,7M garage + 84% fiber =
          <span style={{ color: 'var(--color-accent)', fontWeight: 'bold' }}>Sveriges AI-frihet.</span>
        </motion.p>
        <motion.div
          className="flex gap-4 justify-center flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.6 }}
        >
          <motion.button
            onClick={showComingSoon}
            className="px-8 py-4 bg-green-500 hover:bg-green-600 text-black font-bold rounded-lg transition transform hover:scale-105 glow-neon"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            🚀 Starta Min Nod Nu
          </motion.button>
          <motion.button
            className="px-8 py-4 border-2 border-green-500 hover:bg-green-500/10 text-green-400 font-bold rounded-lg transition"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            📚 Läs Dokumentation
          </motion.button>
          <motion.button
            className="px-8 py-4 border-2 border-pink-500 hover:bg-pink-500/10 text-pink-400 font-bold rounded-lg transition"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            💬 Discord Community
          </motion.button>
        </motion.div>
      </div>
    </motion.section>
  );
};

export default CTA;
