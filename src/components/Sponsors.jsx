import { motion } from 'framer-motion';

const Sponsors = () => {
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.3
      }
    }
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 50 },
    visible: { opacity: 1, y: 0 }
  };

  return (
    <motion.section
      id="sponsors"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="text-4xl md:text-5xl font-bold mb-16 glow-text text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        🤝 Sponsors & Lösningar
      </motion.h2>

      <motion.div
        variants={itemVariants}
        className="grid md:grid-cols-2 gap-12"
      >
        {/* AUTOVERSIO CARD */}
        <motion.div
          variants={itemVariants}
          className="card-hover fact-box"
        >
          <div className="flex items-center gap-3 mb-6">
            <div className="text-3xl">🔐</div>
            <h3 className="text-2xl font-bold" style={{ color: 'var(--color-accent)' }}>Autoversio</h3>
          </div>

          <p className="text-sm text-gray-300 mb-6">
            <strong>Privat & Säker AI-Inferens för Svenska Företag</strong>
          </p>

          <div className="bg-slate-950/50 p-4 rounded border border-green-500/20 mb-6">
            <p className="text-sm font-bold text-green-400 mb-3">🎯 Integritet & Säkerhet Först</p>
            <ul className="space-y-2 text-sm text-gray-300">
              <li>✅ <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>AI i alla processer</span> – från dataintag till output</li>
              <li>✅ <strong>Dataflöden:</strong> Privat AI-baserad ETL och databearbetning</li>
              <li>✅ <strong>Produktutveckling:</strong> AI-accelererad innovation med säkerhet</li>
              <li>✅ <strong>Mjukvaruutveckling:</strong> Agentic AI för kodgenerering & QA</li>
              <li>✅ <strong>Interna system:</strong> Säker AI-automatisering av kärnprocesser</li>
            </ul>
          </div>

          <p className="text-xs text-gray-400 mb-6">
            Autoversio levererar end-to-end säker AI-infrastruktur för svenska organisationer där integritet och GDPR-compliance är icke-förhandlingsbar.
          </p>

          <motion.a
            href="https://www.autoversio.com"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-block px-6 py-3 bg-green-500 hover:bg-green-600 text-black font-bold rounded-lg transition"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            📞 Kontakta Autoversio
          </motion.a>
        </motion.div>

        {/* LITEIT CARD */}
        <motion.div
          variants={itemVariants}
          className="card-hover fact-box"
        >
          <div className="flex items-center gap-3 mb-6">
            <div className="text-3xl">⚡</div>
            <h3 className="text-2xl font-bold" style={{ color: 'var(--color-warning)' }}>Liteit</h3>
          </div>

          <p className="text-sm text-gray-300 mb-6">
            <strong>AI-Driven Mjukvaruutveckling – 20x Snabbare</strong>
          </p>

          <div className="bg-slate-950/50 p-4 rounded border border-yellow-500/20 mb-6">
            <p className="text-sm font-bold text-yellow-400 mb-3">🚀 Agentic AI Code Generation</p>
            <ul className="space-y-2 text-sm text-gray-300">
              <li>⏱️ <span style={{ color: 'var(--color-accent)', fontWeight: 'bold' }}>Applikationer utvecklas ~20x snabbare</span> än traditionell kod</li>
              <li>📊 <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Lika hög eller högre kvalité</span> – AI-code agenter på nivå med erfarna utvecklare</li>
              <li>🧠 <span style={{ color: 'var(--color-primary)' }}>Context window evolution:</span> 20k → Full GitHub repo-förståelse (senaste 24 mån)</li>
              <li>🔍 <span style={{ color: 'var(--color-primary)' }}>Helhetsbild av projekt:</span> Minskad risk, ökad säkerhet, exploit-detektering</li>
              <li>🛡️ <span style={{ color: 'var(--color-primary)' }}>AI-detekterade säkerhetshål:</span> Automatisk remediation i kodgen</li>
            </ul>
          </div>

          <div className="bg-gradient-to-r from-slate-950/50 to-slate-950/70 p-4 rounded border border-yellow-500/30 mb-6">
            <p className="text-xs text-gray-400">
              <span style={{ color: 'var(--color-warning)', fontWeight: 'bold' }}>Faktum:</span> Moderna AI-code agenter presterar på samma nivå som erfarna utvecklare – skillnaden är en <span style={{ color: 'var(--color-warning)' }}>faktor x20 i hastighet.</span> Kod är producerad, testbar och deploybar på sekunder istället för dagar.
            </p>
          </div>

          <motion.a
            href="https://www.liteit.se"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-block px-6 py-3 bg-yellow-500 hover:bg-yellow-600 text-black font-bold rounded-lg transition"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            🚀 Läs Mer om Liteit
          </motion.a>
        </motion.div>
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="mt-12 text-center p-6 rounded border border-green-500/30 bg-slate-950/30"
      >
        <p className="text-sm text-gray-300 mb-4">
          <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Garage AI Community</span> +
          <span style={{ color: 'var(--color-accent)', fontWeight: 'bold' }}>Autoversio (Säker Enterprise AI)</span> +
          <span style={{ color: 'var(--color-warning)', fontWeight: 'bold' }}>Liteit (AI-Driven Dev)</span> =
          <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Sveriges AI-revolution</span>
        </p>
        <p className="text-xs text-gray-400">
          Gräsrötter möter enterprise. Lokala noder. Säker infrastruktur. Snabbaste utveckling. Suveränitet.
        </p>
      </motion.div>
    </motion.section>
  );
};

export default Sponsors;
