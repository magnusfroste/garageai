import { motion } from 'framer-motion';

const CTA = () => {
  return (
    <motion.section
      className="py-20 px-4 max-w-6xl mx-auto text-center"
      initial={{ opacity: 0, y: 50 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
    >
      <div className="fact-box">
        <motion.div
          className="text-5xl mb-6"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
        >
          🏠
        </motion.div>
        <motion.h2
          className="apple-heading-1 mb-6 glow-text"
          style={{ color: 'var(--color-primary)' }}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          Your Garage Is Ready.<br />Is Europe?
        </motion.h2>
        <motion.p
          className="apple-body-large mb-4 max-w-2xl mx-auto"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.3 }}
        >
          The infrastructure for Europe's sovereign AI future isn't a billion-euro data center
          somewhere in the Nordics. It's 75 million garages — already built, already powered,
          already connected.
        </motion.p>
        <motion.p
          className="text-sm mb-10 max-w-xl mx-auto"
          style={{ color: 'var(--color-text-muted)' }}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          The prototype is running on GitHub. The nodes are online. Sweden takes the lead.
          The only thing missing is you.
        </motion.p>
        <motion.div
          className="flex gap-4 justify-center flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.5 }}
        >
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/blob/main/docs/GET_STARTED.md', '_blank')}
            className="px-8 py-4 font-bold rounded-lg transition transform glow-neon"
            style={{ backgroundColor: 'var(--color-primary)', color: 'black' }}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            🚀 Start My Node
          </motion.button>
          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai', '_blank')}
            className="px-8 py-4 border-2 font-bold rounded-lg transition"
            style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)', backgroundColor: 'transparent' }}
            whileHover={{ scale: 1.05, backgroundColor: 'rgba(6, 182, 212, 0.08)' }}
            whileTap={{ scale: 0.95 }}
          >
            🔓 Open Source on GitHub
          </motion.button>
        </motion.div>
      </div>
    </motion.section>
  );
};

export default CTA;
