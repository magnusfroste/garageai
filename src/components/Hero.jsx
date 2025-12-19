import { motion } from 'framer-motion';

const Hero = () => {
  const showComingSoon = () => {
    alert('🚀 Community-portalen lanseras snart!\n\nI mellan tiden:\n\n📞 Autoversio → autoversio.com\n⚡ Liteit → liteit.se');
  };

  return (
    <section className="hero-gradient min-h-screen flex items-center justify-center pt-20 relative overflow-hidden" style={{ background: 'var(--color-bg-secondary)' }}>
      {/* Subtle Apple-style background */}
      <div className="absolute inset-0 opacity-30">
        <div
          style={{
            position: 'absolute',
            top: '20%',
            left: '10%',
            width: '200px',
            height: '200px',
            background: 'radial-gradient(circle, rgba(0, 122, 255, 0.1) 0%, transparent 70%)',
            borderRadius: '50%'
          }}
        />
        <div
          style={{
            position: 'absolute',
            top: '60%',
            right: '15%',
            width: '150px',
            height: '150px',
            background: 'radial-gradient(circle, rgba(52, 199, 89, 0.1) 0%, transparent 70%)',
            borderRadius: '50%'
          }}
        />
      </div>

      <div className="relative z-10 text-center max-w-4xl mx-auto px-4">
        <motion.h1
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          className="apple-heading-1 mb-8"
        >
          GARAGE AI
        </motion.h1>

        <motion.h2
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="apple-heading-2 mb-8"
        >
          Din garage → Europas AI-motor
        </motion.h2>

        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.4 }}
          className="apple-body mb-6 max-w-2xl mx-auto"
        >
          Demokratisera AI genom decentraliserad infrastruktur. Var med och bygg Europas mest hållbara AI-nätverk -
          drivet av gaming-datorer och förnybar energi från svenska hem.
        </motion.p>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.5 }}
          className="mb-8"
        >
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto">
            <div className="text-center">
              <div className="text-2xl mb-1">🚀</div>
              <div className="text-sm font-medium text-gray-300">500+ Nodes</div>
              <div className="text-xs text-gray-500">Aktiva deltagare</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">⚡</div>
              <div className="text-sm font-medium text-gray-300">45 tokens/sek</div>
              <div className="text-xs text-gray-500">RTX 4090 cluster</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">🏆</div>
              <div className="text-sm font-medium text-gray-300">vs 200K+ stars</div>
              <div className="text-xs text-gray-500">konkurrerande projekt</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">🌱</div>
              <div className="text-sm font-medium text-gray-300">85% mindre CO₂</div>
              <div className="text-xs text-gray-500">vs molntjänster</div>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="mb-12"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-green-500/10 border border-green-500/20 rounded-full text-green-400 text-sm font-medium">
            <span>🔓</span>
            <span>100% Open Source & MIT Licensed</span>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="flex gap-6 justify-center flex-wrap"
        >
          <motion.button
            onClick={showComingSoon}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            🚀 Bli medlem i communityt
          </motion.button>

          <motion.button
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            📚 Läs mer om tekniken
          </motion.button>
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
