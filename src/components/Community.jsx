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

  const testimonials = [
    {
      quote: "Som tidig användare har jag sett hur Garage AI transformerar idle gaming-hårdvara till värdefulla AI-resurser. Tekniken fungerar verkligen!",
      author: "Garage AI Core Team",
      location: "Stockholm",
      avatar: "🚀"
    },
    {
      quote: "Att kunna köra avancerade AI-modeller lokalt utan att offra gaming-prestanda är revolutionerande. Detta löser verkligen ett stort problem.",
      author: "Early Adopter",
      location: "Göteborg",
      avatar: "🎯"
    },
    {
      quote: "Distributed AI med gaming-PCs som infrastruktur - det här är framtiden för hållbar AI-utveckling i Europa.",
      author: "AI Researcher",
      location: "Malmö",
      avatar: "🧠"
    }
  ];

  const stats = [
    { label: "Aktiva Noder", value: "500+", icon: "🖥️" },
    { label: "AI Tokens Skapade", value: "2.1M", icon: "🪙" },
    { label: "CO₂ Sparad", value: "85%", icon: "🌱" },
    { label: "Genomströmning", value: "45 t/s", icon: "⚡" }
  ];

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
        className="apple-heading-1 mb-12 glow-text text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        🚀 Vår Community
      </motion.h2>

      <motion.p
        variants={itemVariants}
        className="apple-body max-w-2xl mx-auto mb-16 text-center"
      >
        Var med i den växande communityn av gaming-entusiaster som bygger framtiden för AI.
        Varje nod gör skillnad.
      </motion.p>

      {/* Stats Grid */}
      <motion.div
        variants={itemVariants}
        className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-16"
      >
        {stats.map((stat, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="text-center p-6 bg-slate-950/30 rounded-xl border border-slate-700/50"
            whileHover={{ scale: 1.05 }}
          >
            <div className="text-3xl mb-2">{stat.icon}</div>
            <div className="text-2xl font-bold text-white mb-1">{stat.value}</div>
            <div className="text-sm text-gray-400">{stat.label}</div>
          </motion.div>
        ))}
      </motion.div>

      {/* Testimonials */}
      <motion.div
        variants={itemVariants}
        className="mb-16"
      >
        <motion.h3
          variants={itemVariants}
          className="apple-heading-2 text-center mb-8"
          style={{ color: 'var(--color-accent)' }}
        >
          Vad säger våra medlemmar?
        </motion.h3>
        <div className="grid md:grid-cols-3 gap-6">
          {testimonials.map((testimonial, index) => (
            <motion.div
              key={index}
              variants={itemVariants}
              className="fact-box p-6"
              whileHover={{ y: -5 }}
            >
              <div className="text-4xl mb-4 text-center">{testimonial.avatar}</div>
              <blockquote className="text-sm italic text-gray-300 mb-4 leading-relaxed">
                "{testimonial.quote}"
              </blockquote>
              <div className="text-right">
                <div className="font-semibold" style={{ color: 'var(--color-primary)' }}>
                  {testimonial.author}
                </div>
                <div className="text-xs text-gray-500">{testimonial.location}</div>
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>

      {/* Community Features */}
      <motion.div
        variants={itemVariants}
        className="grid md:grid-cols-2 gap-8 mb-16"
      >
        <motion.div
          variants={itemVariants}
          className="fact-box p-8"
        >
          <motion.h4
            variants={itemVariants}
            className="apple-heading-2 mb-4"
            style={{ color: 'var(--color-primary)' }}
          >
            🎯 Top 100 Leaderboard
          </motion.h4>
          <ul className="space-y-3 text-sm">
            <li className="flex items-center gap-2">
              <span className="text-yellow-400">🥇</span>
              <span>Dubbel GAI-multiplikator 6 månader</span>
            </li>
            <li className="flex items-center gap-2">
              <span className="text-gray-400">🥈</span>
              <span>50% extra GAI 3 månader</span>
            </li>
            <li className="flex items-center gap-2">
              <span className="text-amber-600">🥉</span>
              <span>Gratis hårdvara-upgrade kit</span>
            </li>
          </ul>
        </motion.div>

        <motion.div
          variants={itemVariants}
          className="fact-box p-8"
        >
          <motion.h4
            variants={itemVariants}
            className="apple-heading-2 mb-4"
            style={{ color: 'var(--color-accent)' }}
          >
            🤝 Community Events
          </motion.h4>
          <ul className="space-y-3 text-sm">
            <li className="flex items-center gap-2">
              <span>🎮</span>
              <span>Månatliga gaming-turneringar</span>
            </li>
            <li className="flex items-center gap-2">
              <span>💻</span>
              <span>Tekniska workshops</span>
            </li>
            <li className="flex items-center gap-2">
              <span>🌍</span>
              <span>Globala meetups</span>
            </li>
          </ul>
        </motion.div>
      </motion.div>

      {/* Call to Action */}
      <motion.div
        variants={itemVariants}
        className="text-center"
      >
        <div className="bg-slate-950/50 border border-slate-700/50 rounded-2xl p-8 max-w-2xl mx-auto">
          <motion.h3
            variants={itemVariants}
            className="apple-heading-2 mb-4"
            style={{ color: 'var(--color-primary)' }}
          >
            Bli en del av rörelsen! 🚀
          </motion.h3>
          <p className="text-gray-300 mb-6">
            Gå med i 500+ gaming-entusiaster som redan bidrar till Europas AI-framtid.
            Varje RTX räknas.
          </p>
          <div className="flex gap-4 justify-center flex-wrap">
            <motion.button
              onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
              className="apple-button-primary"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              💬 GitHub Diskussioner
            </motion.button>
            <motion.button
              onClick={() => window.open('https://github.com/magnusfroste/garageai', '_blank')}
              className="apple-button-secondary"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              🔓 GitHub
            </motion.button>
          </div>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default Community;
