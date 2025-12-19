import { motion } from 'framer-motion';

const Infrastructure = () => {
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

  const showComingSoon = () => {
    alert('🚀 Community-portalen lanseras snart!\n\nI mellan tiden:\n\n📞 Autoversio → autoversio.com\n⚡ Liteit → liteit.se');
  };

  return (
    <motion.section
      id="infrastructure"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="apple-heading-1 mb-6 glow-text text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        ⚡ Sverige's Infrastruktur Kan Göra Det – Vi Bara Behöver Tro
      </motion.h2>

      <motion.div
        variants={itemVariants}
        className="max-w-3xl mx-auto mb-12 text-center"
      >
        <p className="apple-body mb-4 leading-relaxed">
          Medan andra länder kämpar med elbrist – vi exporterar el till Tyskland & Danmark. Medan globala AI-tjänster bygger enorma datacenter – vi har 1.7 miljoner garage som kan skapa en distribuerad, miljövänlig infrastruktur.
        </p>
        <p className="apple-body mb-4 leading-relaxed">
          Men här är möjligheten: <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Vår infrastruktur kan skapa nya värden lokalt.</span>
        </p>
        <p className="apple-body mb-6 leading-relaxed">
          Svenska solpaneler kan driva svenska AI-modeller. Svensk fiber kan transportera svensk data säkert. Och svenska garage kan bli grunden för Europas mest hållbara AI-infrastruktur.
        </p>
        <p className="apple-body mb-8 leading-relaxed">
          Vad om vi kompletterade de globala tjänsterna? Vad om <span style={{ color: 'var(--color-accent)', fontWeight: 'bold' }}>din garage blev en motor i Europas AI-infrastruktur – OCH du tjänade på det?</span>
        </p>
        <p className="apple-body-large font-semibold italic text-gray-500">
          Det är inte teknisk omöjlighet. Det är en valsituation.
        </p>
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="text-center mb-16"
      >
        <h3 className="text-xl font-bold mb-4" style={{ color: 'var(--color-accent)' }}>🚀 Det Börjar Med Ett Enkelt Val</h3>
        <motion.button
          onClick={() => document.getElementById('facts-grid').scrollIntoView({ behavior: 'smooth' })}
          className="px-8 py-3 bg-green-500/20 border border-green-500 hover:bg-green-500/30 text-green-400 font-bold rounded-lg transition"
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
        >
          Se Hur Det Fungerar ↓
        </motion.button>
      </motion.div>

      <motion.div
        variants={itemVariants}
        id="facts-grid"
        className="grid md:grid-cols-2 gap-8 mb-16"
      >
        <motion.div
          variants={itemVariants}
          className="card-hover fact-box"
          style={{ background: 'rgba(255, 0, 170, 0.08)', border: '1px solid rgba(255, 0, 170, 0.2)' }}
        >
          <h3 className="text-2xl font-bold mb-6" style={{ color: 'var(--color-accent)' }}>
            🇸🇪 Sverige's Paradox
          </h3>

          <div className="mb-6">
            <p className="text-sm font-bold text-gray-300 mb-3">Hur Det Är Idag:</p>
            <ul className="space-y-2 text-sm text-gray-300">
              <li>✓ Sverige: Billigaste el i Europa (€0.04/kWh)</li>
              <li>✓ Sverige: Världsklass fiber (98% coverage)</li>
              <li>✓ Sverige: 1.7M garage, kallklimat</li>
              <li>❌ Men: Microsoft Azure använder denna för GLOBALA kunder</li>
              <li>❌ Resultat: Sverige är DATA-EXPORT-NATION</li>
            </ul>
          </div>

          <div className="bg-slate-950/70 p-4 rounded border-l-4" style={{ borderColor: 'var(--color-accent)' }}>
            <p className="text-xs font-bold mb-3" style={{ color: 'var(--color-accent)' }}>📊 Siffror:</p>
            <ul className="space-y-2 text-xs text-gray-300">
              <li>• Microsoft får €0.01-1.00/request → Du får: 0</li>
              <li>• Datacenter kräver 87% MER elnätskapacitet</li>
              <li>• Centraliserad el → 8-9.5% förlust → 30 TWh/år slöseri</li>
            </ul>
          </div>
        </motion.div>

        <motion.div
          variants={itemVariants}
          className="card-hover fact-box"
          style={{ background: 'rgba(0, 255, 136, 0.08)', border: '1px solid rgba(0, 255, 136, 0.2)' }}
        >
          <h3 className="text-2xl font-bold mb-6" style={{ color: 'var(--color-primary)' }}>
            🏠 Garage AI: Kontrollen Tillbaka
          </h3>

          <div className="mb-6">
            <p className="text-sm font-bold text-gray-300 mb-3">En Helt Annorlunda Väg:</p>
            <ul className="space-y-2 text-sm text-gray-300">
              <li>✓ 5,000 garage-noder (du & grannar)</li>
              <li>✓ Lokal processing (noll långväga överföring)</li>
              <li>✓ GAI-tokens som belöning (DU tjänar)</li>
              <li>✓ 85-90% mindre elförlust</li>
              <li>✓ Värde stannar i Sverige</li>
            </ul>
          </div>

          <div className="bg-slate-950/70 p-4 rounded border-l-4" style={{ borderColor: 'var(--color-primary)' }}>
            <p className="text-xs font-bold mb-3" style={{ color: 'var(--color-primary)' }}>📈 Förändring:</p>
            <ul className="space-y-2 text-xs text-gray-300">
              <li>• 16 MW distribuerad vs 500 MW centraliserad</li>
              <li>• 0 transmission upgrades (fiber redan finns)</li>
              <li>• 3,650 GAI/år per nod (värde stannar hos dig)</li>
            </ul>
          </div>
        </motion.div>
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="bg-slate-950/70 border border-green-500/30 p-8 rounded text-center"
      >
        <h2 className="text-3xl font-bold mb-6 glow-text" style={{ color: 'var(--color-primary)' }}>
          🌍 Det Här Är Inte Ett Tech-Problem – Det Är Ett Valsituation
        </h2>

        <div className="max-w-3xl mx-auto mb-8">
          <div className="flex flex-col md:flex-row gap-6 justify-center mb-12">
            <div className="flex-1 bg-slate-950/50 p-6 rounded border-l-4 border-red-400">
              <p className="font-bold text-lg mb-4">A) Fortsätta Som Nu</p>
              <p className="text-sm text-gray-400">Sverige exporterar infrastruktur. USA-bolag tjänar pengarna. Vi får jobb och skatter.</p>
            </div>

            <div className="flex-1 bg-slate-950/50 p-6 rounded border-l-4" style={{ borderColor: 'var(--color-primary)' }}>
              <p className="font-bold text-lg mb-4">B) Ta Kontrollen</p>
              <p className="text-sm text-gray-400">Garage AI + Autoversio + Liteit. Sverige behåller värde, data, framtid.</p>
            </div>
          </div>

          <p className="text-2xl font-bold glow-text mb-6" style={{ color: 'var(--color-accent)' }}>
            Du bestämmer om ditt garage blir en motor i Sveriges AI-revolution.
          </p>
        </div>

        <div className="flex gap-4 justify-center flex-wrap">
          <motion.button
            onClick={showComingSoon}
            className="px-8 py-4 bg-green-500 hover:bg-green-600 text-black font-bold rounded-lg transition transform hover:scale-105 glow-neon"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            🚀 Starta Din AI-Nod
          </motion.button>
          <motion.button
            onClick={() => document.getElementById('ai-explained').scrollIntoView({ behavior: 'smooth' })}
            className="px-8 py-4 border-2 border-green-500 hover:bg-green-500/10 text-green-400 font-bold rounded-lg transition"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            📊 Se Tekniska Detaljer
          </motion.button>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Infrastructure;
