import { motion } from 'framer-motion';

const AIExplained = () => {
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.2
      }
    }
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 50 },
    visible: { opacity: 1, y: 0 }
  };

  return (
    <motion.section
      id="ai-explained"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="text-4xl md:text-5xl font-bold mb-12 glow-text text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        🧠 Vad är AI-Inferens? Enkelt Förklarat.
      </motion.h2>

      <motion.div
        variants={itemVariants}
        className="grid md:grid-cols-3 gap-8 mb-12"
      >
        <motion.div variants={itemVariants} className="card-hover fact-box">
          <h3 className="text-xl font-bold mb-4" style={{ color: 'var(--color-accent)' }}>1. Modell (Tränad AI)</h3>
          <p className="text-sm text-gray-300 mb-4">
            En stor språkmodell från Hugging Face (gratis, 7-70 GB). Redan tränad på miljontals texter.
          </p>
          <p className="text-xs text-gray-500">
            💾 T.ex. "Llama-7B" = 7 miljarder parametrar = ~13 GB fil
          </p>
        </motion.div>

        <motion.div variants={itemVariants} className="card-hover fact-box">
          <h3 className="text-xl font-bold mb-4" style={{ color: 'var(--color-warning)' }}>2. Prompt (Din Input)</h3>
          <p className="text-sm text-gray-300 mb-4">
            Du ställer en fråga. Prompten stannar 100% lokalt i ditt garage via fiber.
          </p>
          <p className="text-xs text-gray-500">
            🗣️ Ingen data lämnar ditt garage – allt är lokalt!
          </p>
        </motion.div>

        <motion.div variants={itemVariants} className="card-hover fact-box">
          <h3 className="text-xl font-bold mb-4" style={{ color: 'var(--color-primary)' }}>3. Inferens (Beräkning)</h3>
          <p className="text-sm text-gray-300 mb-4">
            Din RTX 5090 kör modellen. GPU:n gissar nästa ord tills svar är klart. 15-50 tokens/sekund!
          </p>
          <p className="text-xs text-gray-500">
            ⚡ RTX 5090: 2.6x snabbare än A100
          </p>
        </motion.div>
      </motion.div>

      <motion.div variants={itemVariants} className="fact-box mb-8">
        <h3 className="text-2xl font-bold mb-6" style={{ color: 'var(--color-primary)' }}>🔄 Inferens-Flöde: Cloud Services & Garage AI</h3>
        <div className="grid md:grid-cols-2 gap-8">
          <div>
            <p className="text-sm font-bold text-blue-400 mb-4">☁️ Cloud AI Services</p>
            <div className="text-sm space-y-2 text-gray-300">
              <div className="flex items-start gap-2">
                <span className="text-blue-400">1.</span>
                <span>Perfekt för snabb prototyping & skalning</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-blue-400">2.</span>
                <span>Säker kryptering under transport</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-blue-400">3.</span>
                <span>State-of-the-art modeller tillgängliga direkt</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-blue-400">4.</span>
                <span>Data hanteras enligt tjänstens policy</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-blue-400">💰</span>
                <span><span style={{ color: 'var(--color-warning)' }}>Pay-per-use prissättning</span></span>
              </div>
            </div>
          </div>

          <div>
            <p className="text-sm font-bold text-green-400 mb-4">🏠 Garage AI (Din RTX)</p>
            <div className="text-sm space-y-2 text-gray-300">
              <div className="flex items-start gap-2">
                <span style={{ color: 'var(--color-primary)' }}>1.</span>
                <span>Prompt stannar <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>100% lokalt</span></span>
              </div>
              <div className="flex items-start gap-2">
                <span style={{ color: 'var(--color-primary)' }}>2.</span>
                <span>✅ Din RTX kör modellen (gratis från Hugging Face)</span>
              </div>
              <div className="flex items-start gap-2">
                <span style={{ color: 'var(--color-primary)' }}>3.</span>
                <span>✅ <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Full lokal kontroll</span> – maximal integritet</span>
              </div>
              <div className="flex items-start gap-2">
                <span style={{ color: 'var(--color-primary)' }}>4.</span>
                <span>✅ Du äger datan 100%, ingen extern lagring</span>
              </div>
              <div className="flex items-start gap-2">
                <span style={{ color: 'var(--color-primary)' }}>💰</span>
                <span><span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Endast din el-kostnad</span></span>
              </div>
            </div>
          </div>
        </div>
      </motion.div>

      <motion.div variants={itemVariants} className="fact-box">
        <h3 className="text-2xl font-bold mb-6 glow-text" style={{ color: 'var(--color-accent)' }}>
          🏘️ Self-Hosted AI för Hela Grannlaget
        </h3>
        <p className="text-sm text-gray-300 mb-6">
          Garage AI är inte bara för dig – det är för ditt helt grannskap. En <strong>self-hosted API</strong> i ditt garage blir infrastrukturen för din lokala gemenskap.
        </p>

        <div className="grid md:grid-cols-2 gap-8 mb-8">
          <div className="bg-slate-950/50 p-4 rounded border border-pink-500/20">
            <p className="text-sm font-bold text-pink-400 mb-3">🏠 Din Garage = Lokal API-Server</p>
            <ul className="space-y-2 text-sm text-gray-300">
              <li>✅ RTX 5090-kluster kör privat modell</li>
              <li>✅ API exponeras lokalt på fiber</li>
              <li>✅ Hela grannskapet får åtkomst utan BigTech</li>
              <li>✅ Du kontrollerar modell, data, regler</li>
            </ul>
          </div>

          <div className="bg-slate-950/50 p-4 rounded border border-green-500/20">
            <p className="text-sm font-bold text-green-400 mb-3">👥 Grannarna Drar Nytta</p>
            <ul className="space-y-2 text-sm text-gray-300">
              <li>📱 Enskilda: Koppla sina enheter till din API</li>
              <li>🏢 Små företag: Säker privat AI för kärnprocesser</li>
              <li>🎓 Skolor: Lokal AI (no tracking)</li>
              <li>🤖 IoT: Smarta hemmet får AI-intelligens</li>
            </ul>
          </div>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default AIExplained;
