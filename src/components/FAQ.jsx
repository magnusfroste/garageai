import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.1 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 }
};

const faqs = [
  {
    question: 'What exactly is a Garage AI node?',
    answer: 'A Garage AI node is a computer — a Mac Mini, a gaming PC with a GPU, or a dedicated inference board — sitting in your garage, running local AI models. It connects to your home network via fiber and serves your household and potentially your neighbourhood with private, fast AI inference. No data leaves your property.',
  },
  {
    question: 'Why does local inference matter for autonomous agents?',
    answer: 'Autonomous agents make hundreds of model calls per task — planning, tool use, reflection, decision-making. At €0.01–0.10 per API call to cloud providers, the cost explodes fast. More critically, agents handling private data (home security, health, finances) cannot safely rely on third-party cloud servers. Local inference is the only answer for true autonomy.',
  },
  {
    question: 'How does V2G (Vehicle-to-Grid) connect to this?',
    answer: 'Your electric car\'s battery is a 60–100 kWh energy store parked in your garage for ~22 hours a day. V2G technology lets it supply power back to your home (V2H) or the grid. Combined with solar panels, this means your AI inference hardware can run on near-zero-cost renewable energy — day or night.',
  },
  {
    question: 'What hardware do I need to get started?',
    answer: 'If you already have a gaming PC with an RTX 30/40/50-series GPU, you can start today — free. The software is open source on GitHub. For new builds, a Mac Mini M4 Pro (~€1,400) or a mini-PC with integrated GPU is enough to run 7B–70B parameter models. No special hardware needed.',
  },
  {
    question: 'Is this just for tech enthusiasts?',
    answer: 'Not at all. The concept scales from a single node running in a home to a municipal initiative deploying nodes across a region. Many enthusiasts already run private AI servers at home — Garage AI gives this a structure, a protocol, and a European mission. The installation process is being designed for non-technical users.',
  },
  {
    question: 'How does this relate to European AI sovereignty?',
    answer: 'Currently, almost all AI inference for European users happens on US-owned infrastructure — exposed to the US Cloud Act, foreign jurisdiction, and commercial decisions by American companies. Garage AI builds inference capacity on European soil, owned by European individuals and communities. It\'s the fiber broadband model for AI: distributed, resilient, local.',
  },
  {
    question: 'What about security and privacy?',
    answer: '100% local processing. Your data never leaves your garage. Models run on-device. The Garage AI protocol is open source and auditable. For the neighbourhood-sharing mode, requests are processed locally and only anonymised metadata (not content) is logged for load balancing.',
  },
  {
    question: 'Where does Sweden\'s fiber advantage come in?',
    answer: 'AI inference nodes need to communicate — for coordination, for load distribution, for software updates. Sweden\'s 1–10 Gbit/s fiber coverage (98% of the population) means this coordination happens at near-zero latency without strain on the network. This is why Sweden is the ideal first market.',
  },
  {
    question: 'Is the €100M announcement real?',
    answer: 'The initiative and the vision are 100% real. The prototype is running on GitHub right now. The exact funding figure? We published this on April 1st — draw your own conclusions. But a sovereign AI infrastructure initiative backed at that scale? It should happen, and we\'re making the case for it.',
  },
];

const FAQ = () => {
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
        Frequently Asked Questions
      </motion.h2>
      <motion.p
        variants={itemVariants}
        className="apple-body mb-12 text-center max-w-xl mx-auto"
      >
        Everything you need to know about Garage AI, local inference, and Europe's distributed AI future.
      </motion.p>

      <motion.div variants={itemVariants} className="space-y-4">
        {faqs.map((faq, index) => (
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
