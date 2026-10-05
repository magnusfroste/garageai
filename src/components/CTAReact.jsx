import { motion } from 'framer-motion';

const CTA = ({
  icon = '🏠',
  title = 'Your Garage Is Ready.',
  subtitle = 'Is Europe?',
  description = "Europe's AI capacity doesn't have to be only a billion-euro data centre somewhere. Part of it can be the GPUs people and companies already own, connected, local and paid.",
  secondaryText = 'The mesh, the gateway and the portal are live, and the first garages are serving models.',
  buttons = [
    {
      text: '🚀 Offer Your GPU',
      url: 'https://app.garageai.eu/auth?intent=operator',
      variant: 'primary',
    },
    {
      text: '⚡ Use AI',
      url: 'https://app.garageai.eu/auth?intent=buyer',
      variant: 'secondary',
    },
    {
      text: '🔓 Open Source on GitHub',
      url: 'https://github.com/magnusfroste/garageai',
      variant: 'secondary',
    },
  ],
} = {}) => {
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
          {icon}
        </motion.div>
        <motion.h2
          className="apple-heading-1 mb-6 glow-text"
          style={{ color: 'var(--color-primary)' }}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          {title}<br />{subtitle}
        </motion.h2>
        <motion.p
          className="apple-body-large mb-4 max-w-2xl mx-auto"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.3 }}
        >
          {description}
        </motion.p>
        <motion.p
          className="text-sm mb-10 max-w-xl mx-auto"
          style={{ color: 'var(--color-text-muted)' }}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          {secondaryText}
        </motion.p>
        <motion.div
          className="flex gap-4 justify-center flex-wrap"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.5 }}
        >
          {(buttons || []).map((button, index) =>
            button.variant === 'primary' ? (
              <motion.button
                key={index}
                onClick={() => window.open(button.url, '_blank')}
                className="px-8 py-4 font-bold rounded-lg transition transform glow-neon"
                style={{ backgroundColor: 'var(--color-primary)', color: 'black' }}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                {button.text}
              </motion.button>
            ) : (
              <motion.button
                key={index}
                onClick={() => window.open(button.url, '_blank')}
                className="px-8 py-4 border-2 font-bold rounded-lg transition"
                style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)', backgroundColor: 'transparent' }}
                whileHover={{ scale: 1.05, backgroundColor: 'rgba(6, 182, 212, 0.08)' }}
                whileTap={{ scale: 0.95 }}
              >
                {button.text}
              </motion.button>
            )
          )}
        </motion.div>
      </div>
    </motion.section>
  );
};

export default CTA;
