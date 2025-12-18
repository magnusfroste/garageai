import { motion } from 'framer-motion';

const Roadmap = () => {
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

  const roadmapItems = [
    {
      quarter: "Q4 2024",
      title: "Beta Launch",
      items: [
        "Första 500 noder online",
        "Grundläggande AI-modeller (Llama, Mistral)",
        "GAI-token ekonomi",
        "Discord community"
      ],
      status: "current"
    },
    {
      quarter: "Q1 2025",
      title: "Expansion",
      items: [
        "10,000+ noder målet",
        "Avancerade modeller (GPT-4 nivå)",
        "Enterprise integration",
        "Mobil app för nod-hantering"
      ],
      status: "upcoming"
    },
    {
      quarter: "Q2 2025",
      title: "Ecosystem Growth",
      items: [
        "AI-marknadplats för tjänster",
        "Partner-integrationer",
        "DeFi features för GAI-tokens",
        "Internationell expansion"
      ],
      status: "future"
    },
    {
      quarter: "Q3 2025",
      title: "Scale & Sustainability",
      items: [
        "100,000+ noder",
        "Klimatkompensering genom solenergi",
        "Reglerad token ekonomi",
        "Global datacenter konkurrens"
      ],
      status: "future"
    }
  ];

  const getStatusColor = (status) => {
    switch (status) {
      case 'current': return 'var(--color-primary)';
      case 'upcoming': return 'var(--color-accent)';
      default: return 'var(--color-warning)';
    }
  };

  return (
    <motion.section
      id="roadmap"
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
        🗺️ Roadmap: Från Garage till Global Kraft
      </motion.h2>

      <motion.div
        variants={itemVariants}
        className="grid md:grid-cols-2 lg:grid-cols-4 gap-8 mb-12"
      >
        {roadmapItems.map((phase, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="fact-box relative"
            style={{
              borderColor: getStatusColor(phase.status),
              borderWidth: phase.status === 'current' ? '2px' : '1px'
            }}
          >
            {phase.status === 'current' && (
              <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                <span className="bg-green-500 text-black text-xs font-bold px-3 py-1 rounded-full">
                  🔴 LIVE
                </span>
              </div>
            )}

            <h3 className="text-xl font-bold mb-4" style={{ color: getStatusColor(phase.status) }}>
              {phase.quarter}
            </h3>
            <h4 className="text-lg font-semibold mb-4 text-white">
              {phase.title}
            </h4>

            <ul className="space-y-2 text-sm text-gray-300">
              {phase.items.map((item, itemIndex) => (
                <li key={itemIndex} className="flex items-start gap-2">
                  <span style={{ color: getStatusColor(phase.status) }}>•</span>
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </motion.div>
        ))}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="bg-gradient-to-r from-slate-950/50 to-slate-950/70 border border-green-500/30 p-8 rounded text-center"
      >
        <h3 className="text-2xl font-bold mb-6 glow-text" style={{ color: 'var(--color-primary)' }}>
          🌟 Varför delta nu?
        </h3>

        <div className="grid md:grid-cols-3 gap-6 mb-8">
          <div className="text-center">
            <div className="text-3xl mb-3">🎯</div>
            <h4 className="font-bold mb-2">Early Adopter Advantage</h4>
            <p className="text-sm text-gray-400">Dubbla GAI-belöningar och early access till alla features</p>
          </div>
          <div className="text-center">
            <div className="text-3xl mb-3">🌱</div>
            <h4 className="font-bold mb-2">Forma framtiden</h4>
            <p className="text-sm text-gray-400">Din feedback formar hur Garage AI utvecklas</p>
          </div>
          <div className="text-center">
            <div className="text-3xl mb-3">🤝</div>
            <h4 className="font-bold mb-2">Community Impact</h4>
            <p className="text-sm text-gray-400">Var med och bygg Europas mest hållbara AI-infrastruktur</p>
          </div>
        </div>

        <p className="text-lg text-gray-300 mb-6">
          Vi är precis i början av denna revolution. <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>
          Var med från dag ett och forma framtiden för AI.
          </span>
        </p>

        <motion.button
          className="px-8 py-4 bg-green-500 hover:bg-green-600 text-black font-bold rounded-lg transition transform hover:scale-105 glow-neon"
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
        >
          🚀 Bli en av de första 500
        </motion.button>
      </motion.div>
    </motion.section>
  );
};

export default Roadmap;
