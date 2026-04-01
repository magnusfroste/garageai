import { motion } from 'framer-motion';

const Hero = () => {
  return (
    <section
      className="hero-gradient min-h-screen flex flex-col items-center justify-center pt-20 relative overflow-hidden"
      style={{ background: 'var(--color-bg-primary)' }}
    >
      {/* Animated aurora blobs */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div style={{
          position: 'absolute', top: '10%', left: '5%',
          width: '500px', height: '500px',
          background: 'radial-gradient(circle, rgba(0,212,255,0.10) 0%, transparent 70%)',
          borderRadius: '50%', animation: 'aurora-1 18s ease-in-out infinite',
        }} />
        <div style={{
          position: 'absolute', top: '50%', right: '5%',
          width: '400px', height: '400px',
          background: 'radial-gradient(circle, rgba(0,255,136,0.08) 0%, transparent 70%)',
          borderRadius: '50%', animation: 'aurora-2 22s ease-in-out infinite',
        }} />
        <div style={{
          position: 'absolute', top: '30%', left: '45%',
          width: '600px', height: '600px',
          background: 'radial-gradient(circle, rgba(255,0,170,0.05) 0%, transparent 70%)',
          borderRadius: '50%', animation: 'aurora-3 26s ease-in-out infinite',
        }} />
        <div style={{
          position: 'absolute', bottom: '5%', left: '20%',
          width: '350px', height: '350px',
          background: 'radial-gradient(circle, rgba(255,214,10,0.06) 0%, transparent 70%)',
          borderRadius: '50%', animation: 'aurora-1 20s ease-in-out infinite reverse',
        }} />
      </div>

      {/* Breaking news ticker */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="w-full z-10 overflow-hidden"
        style={{ background: 'rgba(220,38,38,0.9)', borderBottom: '1px solid rgba(255,255,255,0.2)' }}
      >
        <div className="flex items-center gap-4 px-4 py-2 max-w-7xl mx-auto">
          <span className="text-white text-xs font-black uppercase tracking-widest shrink-0 bg-red-800 px-2 py-1 rounded">
            BREAKING
          </span>
          <div className="overflow-hidden whitespace-nowrap">
            <motion.span
              animate={{ x: ['0%', '-50%'] }}
              transition={{ duration: 30, repeat: Infinity, ease: 'linear' }}
              className="inline-block text-white text-xs font-medium"
            >
              Sweden announces €100M Sovereign AI Infrastructure Initiative &nbsp;·&nbsp; 10,000 new jobs secured &nbsp;·&nbsp; Europe's first distributed garage-based AI network &nbsp;·&nbsp; "The biggest infrastructure move since broadband" — Swedish Tech Minister &nbsp;·&nbsp; Sweden announces €100M Sovereign AI Infrastructure Initiative &nbsp;·&nbsp; 10,000 new jobs secured &nbsp;·&nbsp; Europe's first distributed garage-based AI network &nbsp;·&nbsp; "The biggest infrastructure move since broadband" — Swedish Tech Minister &nbsp;·&nbsp;
            </motion.span>
          </div>
        </div>
      </motion.div>

      {/* Main content */}
      <div className="relative z-10 text-center max-w-5xl mx-auto px-4 py-16">

        {/* Date badge */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.1 }}
          className="flex justify-center mb-6"
        >
          <span className="text-xs text-gray-400 border border-gray-700 px-3 py-1 rounded-full">
            April 1, 2026 &nbsp;·&nbsp; garageai.eu &nbsp;·&nbsp; <span style={{ color: 'var(--color-primary)' }}>Live</span>
          </span>
        </motion.div>

        <motion.h1
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="apple-heading-1 mb-4"
          style={{ lineHeight: 1.05 }}
        >
          Your Garage.
          <br />
          <span className="gradient-text-cyan">Europe's AI Engine.</span>
        </motion.h1>

        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.35 }}
          className="text-xl mb-4 max-w-3xl mx-auto leading-relaxed"
          style={{ color: 'var(--color-text-secondary)' }}
        >
          Sweden leads the world's first decentralized sovereign AI infrastructure initiative —
          powered by solar panels, EV batteries, and 1.7 million garages.
          <br />
          <span style={{ color: 'var(--color-accent)', fontWeight: 600 }}>
            Europe follows. The rest of the world watches.
          </span>
        </motion.p>

        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.5 }}
          className="text-sm mb-10 max-w-2xl mx-auto"
          style={{ color: 'var(--color-text-muted)' }}
        >
          Garage AI is a concept accelerating in real time. We stand at a paradigm shift in how
          society functions — and the infrastructure for the next era is already sitting in your driveway.
        </motion.p>

        {/* Bento stats grid */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.55 }}
          className="grid grid-cols-2 md:grid-cols-4 gap-3 max-w-4xl mx-auto mb-10"
        >
          {[
            { icon: '🏠', value: '75M+', label: 'Garages in EU', accent: 'var(--color-primary)', span: false },
            { icon: '⚡', value: '€100M', label: 'Initiative Secured', accent: 'var(--color-warning)', span: false },
            { icon: '👷', value: '10,000', label: 'New Jobs', accent: 'var(--color-tertiary)', span: false },
            { icon: '🛡️', value: '100%', label: 'European Sovereignty', accent: 'var(--color-accent)', span: false },
          ].map((stat, i) => (
            <motion.div
              key={i}
              className="text-center p-5 rounded-2xl"
              style={{
                background: 'rgba(255,255,255,0.04)',
                backdropFilter: 'blur(12px)',
                WebkitBackdropFilter: 'blur(12px)',
                border: `1px solid ${stat.accent}22`,
                boxShadow: `0 0 24px ${stat.accent}0d`,
              }}
              whileHover={{ scale: 1.04, y: -3 }}
              transition={{ type: 'spring', stiffness: 300 }}
            >
              <div className="text-2xl mb-2">{stat.icon}</div>
              <div className="text-2xl font-black mb-1" style={{ color: stat.accent }}>{stat.value}</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>{stat.label}</div>
            </motion.div>
          ))}
        </motion.div>

        {/* Open source badge */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.7 }}
          className="flex justify-center gap-3 flex-wrap mb-10"
        >
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-slate-950 border border-slate-700 rounded-full text-green-400 text-sm font-medium">
            🔓 100% Open Source · MIT License
          </span>
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-slate-950 border border-slate-700 rounded-full text-blue-400 text-sm font-medium">
            🇸🇪 Made in Sweden · Built for Europe
          </span>
        </motion.div>

        {/* CTA buttons */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.75 }}
          className="flex gap-4 justify-center flex-wrap"
        >
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/blob/main/docs/GET_STARTED.md', '_blank')}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            🚀 Start My Node
          </motion.button>
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            💬 Join the Community
          </motion.button>
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai', '_blank')}
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            📚 GitHub
          </motion.button>
        </motion.div>

        {/* LinkedIn share */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 1.0 }}
          className="mt-8 flex flex-col items-center gap-3"
        >
          <motion.button
            onClick={() => {
              const url = 'https://www.garageai.eu';
              const text = encodeURIComponent('🇸🇪 Sweden just became the first country in Europe to launch a sovereign, distributed AI infrastructure initiative.\n\n€100M. 10,000 new jobs. 75 million garages across Europe ready to participate.\n\nSolar panels + EV batteries + local AI inference. The prototype is already running on GitHub.\n\n#GarageAI #SovereignAI #Sweden #EuropeAI #Ollama');
              window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(url)}`, '_blank');
            }}
            className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl font-semibold text-sm transition-all"
            style={{ background: '#0A66C2', color: '#fff', boxShadow: '0 4px 16px rgba(10,102,194,0.35)' }}
            whileHover={{ scale: 1.04, boxShadow: '0 6px 24px rgba(10,102,194,0.5)' }}
            whileTap={{ scale: 0.97 }}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
            </svg>
            Share on LinkedIn
          </motion.button>
          <p className="text-xs" style={{ color: 'var(--color-text-muted)', opacity: 0.45 }}>
            * The vision is real. The funding headline? It&apos;s April 1st — but the prototype is running.
          </p>
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
