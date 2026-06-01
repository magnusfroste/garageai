import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.1 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 }
};

const FAQ = ({ title, description, faqs }) => {
  return (
    <motion.section
      id="faq"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="apple-heading-1 mb-4 text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        {title}
      </motion.h2>
      <motion.p
        variants={itemVariants}
        className="apple-body mb-12 text-center max-w-xl mx-auto"
      >
        {description}
      </motion.p>

      <motion.div variants={itemVariants} className="space-y-4">
        {faqs && faqs.map((faq, index) => (
          <motion.div key={index} variants={itemVariants} className="fact-box">
            <h3 className="text-base font-bold mb-3" style={{ color: 'var(--color-accent)' }}>
              {faq.question}
            </h3>
            <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
              {faq.answer}
            </p>
          </motion.div>
        ))}
      </motion.div>
    </motion.section>
  );
};

export default FAQ;
