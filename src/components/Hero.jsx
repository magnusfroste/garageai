import { motion } from 'framer-motion';

const Hero = () => {

  return (
    <section className="hero-gradient min-h-screen flex items-center justify-center pt-20 relative overflow-hidden" style={{ background: 'var(--color-bg-primary)' }}>
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
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>1.7M+</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>Potentiella noder</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">⚡</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>120 tokens/sek</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>per RTX 4090</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">🏆</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>98%</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>fiber-täckning</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">🌱</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>100%</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>lokal integritet</div>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="mb-12"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-slate-950 border border-slate-700 rounded-full text-green-400 text-sm font-medium">
            <span>🔓</span>
            <span>100% Open Source & MIT Licensed</span>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="flex gap-4 justify-center flex-wrap"
        >
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            🚀 Starta min nod
          </motion.button>

          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            💬 Community
          </motion.button>

          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai', '_blank')}
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            📚 Teknisk info
          </motion.button>
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
