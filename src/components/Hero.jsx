import { motion } from 'framer-motion';

const Hero = () => {
  const showComingSoon = () => {
    alert('🚀 Community-portalen lanseras snart!\n\nI mellan tiden:\n\n📞 Autoversio → autoversio.com\n⚡ Liteit → liteit.se');
  };

  return (
    <section className="hero-gradient min-h-screen flex items-center justify-center pt-20 relative overflow-hidden">
      <div className="absolute inset-0 flex items-center justify-center opacity-20">
        <div
          style={{
            width: '500px',
            height: '500px',
            background: 'radial-gradient(circle, rgba(0, 255, 136, 0.5) 0%, transparent 70%)',
            filter: 'blur(80px)'
          }}
        />
      </div>

      <div className="relative z-10 text-center max-w-4xl mx-auto px-4">
        <motion.h1
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          className="text-7xl md:text-8xl font-black mb-6 glow-text"
          style={{ color: 'var(--color-primary)' }}
        >
          GARAGE AI
        </motion.h1>

        <motion.h2
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="text-3xl md:text-5xl font-bold mb-6 text-white"
        >
          Din garage → Europas AI-motor
        </motion.h2>

        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.4 }}
          className="text-lg md:text-xl text-gray-300 mb-12 max-w-2xl mx-auto"
        >
          Bygg privat inferens på RTX 5090-kluster. <span style={{ color: 'var(--color-primary)' }}>5000 solpaneler.</span>
          <span style={{ color: 'var(--color-accent)' }}>500k gaming-riggar.</span>
          <span style={{ color: 'var(--color-warning)' }}>Tjäna GAI-tokens.</span> Komplettera globala AI-tjänster.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="flex gap-4 justify-center flex-wrap"
        >
          <motion.button
            onClick={showComingSoon}
            className="px-8 py-4 bg-green-500 hover:bg-green-600 text-black font-bold rounded-lg transition transform hover:scale-105 glow-neon"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            🚀 Gå med i rörelsen
          </motion.button>

          <motion.button
            className="px-8 py-4 border-2 border-green-500 hover:bg-green-500/10 text-green-400 font-bold rounded-lg transition"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            📊 Se leaderboard
          </motion.button>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.8 }}
          className="mt-16 flex justify-center gap-8 items-center flex-wrap"
        >
          <div className="text-center">
            <div className="text-2xl font-bold text-green-400">Autoversio</div>
            <a
              href="https://www.autoversio.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-gray-400 hover:text-green-400"
            >
              autoversio.com
            </a>
          </div>

          <div className="text-2xl text-gray-500">×</div>

          <div className="text-center">
            <div className="text-2xl font-bold text-green-400">Liteit</div>
            <a
              href="https://www.liteit.se"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-gray-400 hover:text-green-400"
            >
              liteit.se
            </a>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
