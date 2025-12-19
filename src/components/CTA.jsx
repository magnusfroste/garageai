import { motion } from 'framer-motion';

const CTA = () => {
  const showComingSoon = () => {
    alert('🚀 Community-portalen lanseras snart!\n\nI mellan tiden:\n\n📞 Autoversio → autoversio.com\n⚡ Liteit → liteit.se');
  };

  return (
    <motion.section
      className="py-20 px-4 max-w-6xl mx-auto text-center"
      initial={{ opacity: 0, y: 50 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
    >
      <div className="fact-box">
        <motion.h2
          className="apple-heading-1 mb-6 glow-text"
          style={{ color: 'var(--color-primary)' }}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          Var med och forma framtiden
        </motion.h2>
        <motion.p
          className="apple-body-large mb-8 max-w-2xl mx-auto"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          Var med från början och hjälp oss bygga Europas mest hållbara AI-infrastruktur.
          Varje garage blir en viktig del av nätverket.
        </motion.p>
        <motion.div
          className="flex gap-4 justify-center flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.6 }}
        >
          <motion.button
            onClick={showComingSoon}
            className="px-8 py-4 font-bold rounded-lg transition transform hover:scale-105 glow-neon"
            style={{ backgroundColor: 'var(--color-primary)', color: 'black' }}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            🚀 Starta Min Nod Nu
          </motion.button>
          <motion.button
            onClick={() => window.open('https://github.com/garage-ai/platform', '_blank')}
            className="px-8 py-4 border-2 font-bold rounded-lg transition hover:bg-opacity-10"
            style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)', backgroundColor: 'transparent' }}
            whileHover={{ scale: 1.05, backgroundColor: 'rgba(6, 182, 212, 0.1)' }}
            whileTap={{ scale: 0.95 }}
          >
            🔓 GitHub (Open Source)
          </motion.button>

        </motion.div>
      </div>
    </motion.section>
  );
};

export default CTA;
