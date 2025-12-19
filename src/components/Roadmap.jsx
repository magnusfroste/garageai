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
      quarter: "2025",
      title: "Prototyping",
      items: [
        "Teknisk research och Privat AI-tester med validering",
        "Lokala kluster körs idag",
        "Proof-of-concept för garage-noder",
        "Säkerhets- och prestandatester"
      ],
      status: "current"
    },
    {
      quarter: "2026",
      title: "Launch",
      items: [
        "Första 1000 noder online",
        "AI-modeller för svenska användare",
        "Grundläggande token-system",
        "Community-styrd utveckling"
      ],
      status: "upcoming"
    },
    {
      quarter: "2027",
      title: "Scale",
      items: [
        "10,000+ aktiva noder",
        "Avancerade AI-modeller",
        "Integration med företag",
        "Europeisk expansion"
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
          className="apple-heading-1 mb-16 text-center"
        >
          Vår Vision
        </motion.h2>

      <motion.div
        variants={itemVariants}
        className="grid md:grid-cols-1 lg:grid-cols-3 gap-8 mb-16"
      >
        {roadmapItems.map((phase, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="apple-card relative"
          >
            {phase.status === 'current' && (
              <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                <span className="text-white text-xs font-bold px-3 py-1 rounded-full" style={{ backgroundColor: 'var(--color-secondary)' }}>
                  🔴 LIVE
                </span>
              </div>
            )}

            <div className="text-center mb-6">
              <h3 className="apple-heading-2 mb-2" style={{ color: getStatusColor(phase.status) }}>
                {phase.quarter}
              </h3>
              <h4 className="apple-body font-semibold">
                {phase.title}
              </h4>
            </div>

            <ul className="space-y-3">
              {phase.items.map((item, itemIndex) => (
                <li key={itemIndex} className="flex items-start gap-3">
                  <span className="text-lg mt-1" style={{ color: getStatusColor(phase.status) }}>•</span>
                  <span className="apple-caption">{item}</span>
                </li>
              ))}
            </ul>
          </motion.div>
        ))}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="apple-card max-w-4xl mx-auto"
      >
        <h3 className="apple-heading-2 mb-8 text-center">
          Varför delta nu?
        </h3>

        <div className="grid md:grid-cols-3 gap-8 mb-12">
          <div className="text-center">
            <div className="text-4xl mb-4">🎯</div>
            <h4 className="apple-heading-3 mb-3">Early Adopter-fördelar</h4>
            <p className="apple-caption">Dubbla belöningar och early access till alla funktioner</p>
          </div>
          <div className="text-center">
            <div className="text-4xl mb-4">🌱</div>
            <h4 className="apple-heading-3 mb-3">Forma framtiden</h4>
            <p className="apple-caption">Din feedback formar hur Garage AI utvecklas</p>
          </div>
          <div className="text-center">
            <div className="text-4xl mb-4">🤝</div>
            <h4 className="apple-heading-3 mb-3">Community Impact</h4>
            <p className="apple-caption">Var med och bygg Europas mest hållbara AI-infrastruktur</p>
          </div>
        </div>

        <div className="text-center mb-8">
          <p className="apple-body mb-6 max-w-2xl mx-auto">
            Vi är precis i början av denna revolution. <span style={{ color: 'var(--color-primary)', fontWeight: '600' }}>
            Var med från dag ett och forma framtiden för AI.
            </span>
          </p>
        </div>

        <div className="text-center">
          <motion.button
            onClick={() => window.open('https://github.com/garageai/garageai/discussions', '_blank')}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            🚀 Bli medlem i communityt
          </motion.button>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Roadmap;
