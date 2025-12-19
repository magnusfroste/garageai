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
      className="py-12 px-4 max-w-4xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.div
        variants={itemVariants}
        className="apple-card text-center"
      >
        <p className="text-sm text-gray-300 mb-4">
          <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Tekniska partners:</span>{' '}
          <a
            href="https://www.autoversio.com"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-primary"
            style={{ color: 'var(--color-accent)' }}
          >
            Autoversio
          </a>{' '}
          &{' '}
          <a
            href="https://www.liteit.se"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-primary"
            style={{ color: 'var(--color-accent)' }}
          >
            Liteit
          </a>
        </p>
        <p className="text-xs text-gray-400">
          Säker AI-infrastruktur och AI-driven utveckling
        </p>
      </motion.div>
    </motion.section>
  );
};

export default Sponsors;
