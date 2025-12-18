import { motion } from 'framer-motion';

const Community = () => {
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
      id="community"
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
        👥 Community: Stupid Simple Onboarding
      </motion.h2>

      <motion.div
        variants={itemVariants}
        className="grid md:grid-cols-3 gap-8 mb-12"
      >
        <motion.div variants={itemVariants} className="card-hover fact-box">
          <h3 className="text-lg font-bold mb-4" style={{ color: 'var(--color-primary)' }}>🔐 Steg 1 (10 sek)</h3>
          <p className="text-sm">Google-login → Auto-detect GPU via WebGPU</p>
        </motion.div>
        <motion.div variants={itemVariants} className="card-hover fact-box">
          <h3 className="text-lg font-bold mb-4" style={{ color: 'var(--color-accent)' }}>⚙️ Steg 2 (20 sek)</h3>
          <p className="text-sm">Toggle "Aktivera nod" → Du är live!</p>
        </motion.div>
        <motion.div variants={itemVariants} className="card-hover fact-box">
          <h3 className="text-lg font-bold mb-4" style={{ color: 'var(--color-warning)' }}>📊 Steg 3 (30 sek)</h3>
          <p className="text-sm">"Din nod #472: 2x RTX5090, 15 tokens/sek! Rank #47"</p>
        </motion.div>
      </motion.div>

      {/* Social Proof - Early Adopters */}
      <motion.div variants={itemVariants} className="grid md:grid-cols-2 gap-8 mb-12">
        <motion.div variants={itemVariants} className="fact-box">
          <h3 className="text-xl font-bold mb-6" style={{ color: 'var(--color-primary)' }}>💬 Vad säger våra Early Adopters?</h3>
          <div className="space-y-6">
            <div className="bg-slate-950/50 p-4 rounded border-l-4" style={{ borderColor: 'var(--color-accent)' }}>
              <p className="text-sm italic text-gray-300 mb-2">"Efter 3 veckor tjänade jag tillräckligt för att köpa en RTX 5090 till. Fantastiskt!"</p>
              <p className="text-xs text-green-400 font-bold">- Marcus K., Stockholm</p>
            </div>
            <div className="bg-slate-950/50 p-4 rounded border-l-4" style={{ borderColor: 'var(--color-primary)' }}>
              <p className="text-sm italic text-gray-300 mb-2">"Lokala företag kan äntligen använda AI utan att oroa sig för datasekretess."</p>
              <p className="text-xs text-green-400 font-bold">- Anna L., Göteborg</p>
            </div>
          </div>
        </motion.div>

        <motion.div variants={itemVariants} className="fact-box">
          <h3 className="text-xl font-bold mb-6" style={{ color: 'var(--color-accent)' }}>🚀 Community Milestones</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm">Beta-testare onboardade</span>
              <span className="text-sm font-bold text-green-400">247 / 500</span>
            </div>
            <div className="w-full bg-slate-950/50 rounded-full h-2">
              <div className="bg-green-500 h-2 rounded-full" style={{ width: '49%' }}></div>
            </div>

            <div className="flex items-center justify-between mt-4">
              <span className="text-sm">Total GAI-tokens distribuerade</span>
              <span className="text-sm font-bold text-yellow-400">127,450</span>
            </div>

            <div className="flex items-center justify-between mt-4">
              <span className="text-sm">Aktiva noder idag</span>
              <span className="text-sm font-bold text-blue-400">89</span>
            </div>
          </div>
        </motion.div>
      </motion.div>

      <motion.div variants={itemVariants} className="fact-box">
        <h3 className="text-2xl font-bold mb-6" style={{ color: 'var(--color-primary)' }}>🎯 Bidra & Tjäna GAI-Tokens</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <tr className="border-b border-green-500/30">
              <th className="text-left py-3 px-4" style={{ color: 'var(--color-primary)' }}>Handling</th>
              <th className="text-left py-3 px-4" style={{ color: 'var(--color-primary)' }}>Belöning</th>
              <th className="text-left py-3 px-4" style={{ color: 'var(--color-primary)' }}>Komplexitet</th>
            </tr>
            <tr className="border-b border-green-500/20">
              <td className="py-3 px-4">Kör nod 24h</td>
              <td className="py-3 px-4">10 GAI/RTX-dag</td>
              <td className="py-3 px-4">1-klick ✓</td>
            </tr>
            <tr className="border-b border-green-500/20">
              <td className="py-3 px-4">Dela tutorial</td>
              <td className="py-3 px-4">50 GAI</td>
              <td className="py-3 px-4">Markdown upload</td>
            </tr>
            <tr className="border-b border-green-500/20">
              <td className="py-3 px-4">Testa beta-jobb</td>
              <td className="py-3 px-4">5 GAI/jobbet</td>
              <td className="py-3 px-4">Accept-knapp</td>
            </tr>
            <tr>
              <td className="py-3 px-4">Solpaneler kopplade</td>
              <td className="py-3 px-4">+20% GAI bonus</td>
              <td className="py-3 px-4">Auto-detect</td>
            </tr>
          </table>
        </div>
      </motion.div>

      {/* Early Adopter Incentives */}
      <motion.div variants={itemVariants} className="mt-8 bg-gradient-to-r from-green-500/10 to-blue-500/10 border border-green-500/30 p-6 rounded">
        <h3 className="text-xl font-bold mb-4" style={{ color: 'var(--color-warning)' }}>🎁 Early Adopter Bonusar (Första 1000 noder)</h3>
        <div className="grid md:grid-cols-3 gap-4 text-sm">
          <div className="text-center">
            <div className="text-2xl mb-2">🏅</div>
            <p className="font-bold">Top 100</p>
            <p className="text-gray-400">Dubbel GAI-multiplikator 6 månader</p>
          </div>
          <div className="text-center">
            <div className="text-2xl mb-2">🎖️</div>
            <p className="font-bold">Top 500</p>
            <p className="text-gray-400">50% extra GAI 3 månader</p>
          </div>
          <div className="text-center">
            <div className="text-2xl mb-2">📋</div>
            <p className="font-bold">Top 1000</p>
            <p className="text-gray-400">Gratis hårdvara-upgrade kit</p>
          </div>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Community;
