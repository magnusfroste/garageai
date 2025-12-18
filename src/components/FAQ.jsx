import { motion } from 'framer-motion';

const FAQ = () => {
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1
      }
    }
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 }
  };

  const faqs = [
    {
      question: "Hur mycket kan jag tjäna med min gaming-rigg?",
      answer: "Med en RTX 4090 kan du tjäna 8-12 GAI-tokens per dag genom att köra AI-inferens under lågtrafik. Det motsvarar ~€5-8 per dag efter el-kostnader. Med solpaneler kopplade får du dessutom 20% bonus."
    },
    {
      question: "Är det säkert? Vad händer med min data?",
      answer: "100% lokal bearbetning. Din data lämnar aldrig din garage. Vi använder samma kryptering som banker. Inga loggar, ingen spårning, ingen tredje part har åtkomst."
    },
    {
      question: "Vad kostar det att komma igång?",
      answer: "Inget! Om du redan har en gaming-PC med RTX 30/40-serie kan du börja direkt. Vi tillhandahåller all programvara gratis. Enda kostnaden är din elförbrukning."
    },
    {
      question: "Hur fungerar GAI-token ekonomin?",
      answer: "Du tjänar GAI-tokens genom att bidra till nätverket. Tokens kan växlas mot tjänster inom ekosystemet: AI-modeller, hårdvara-upgrades, eller senare mot SEK/EUR när vi når scale."
    },
    {
      question: "Vad händer om jag stänger av min dator?",
      answer: "Din nod blir temporärt offline. Du kan när som helst pausa och återuppta utan att förlora din position i nätverket. Vi rekommenderar minst 16h/dag för optimal avkastning."
    },
    {
      question: "Kan företag använda Garage AI?",
      answer: "Absolut! Företag kan sätta upp dedikerade noder för privat AI-inferens. Perfekt för GDPR-känsliga processer som dokumentanalys, kundsupport, eller interna AI-assistenter."
    },
    {
      question: "Vad är skillnaden mot ChatGPT eller Claude?",
      answer: "De är fantastiska för prototyping och allmänna frågor. Garage AI är för när du behöver maximal integritet, lokal kontroll, och vill bidra till ett hållbart AI-ekosystem. Det kompletterar, ersätter inte."
    },
    {
      question: "Hur miljövänligt är det egentligen?",
      answer: "Mycket! Vi använder befintliga gaming-riggar som annars skulle stå idle 20h/dag. Solpaneler ger dessutom extra incitament för förnybar energi. 85% mindre elförlust än centraliserade datacenter."
    }
  ];

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
        className="text-4xl md:text-5xl font-bold mb-12 glow-text text-center"
        style={{ color: 'var(--color-primary)' }}
      >
        ❓ Vanliga Frågor
      </motion.h2>

      <motion.div variants={itemVariants} className="space-y-4">
        {faqs.map((faq, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="fact-box"
          >
            <h3 className="text-lg font-bold mb-3" style={{ color: 'var(--color-accent)' }}>
              {faq.question}
            </h3>
            <p className="text-sm text-gray-300 leading-relaxed">
              {faq.answer}
            </p>
          </motion.div>
        ))}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="mt-12 text-center bg-slate-950/50 border border-green-500/30 p-8 rounded"
      >
        <h3 className="text-xl font-bold mb-4" style={{ color: 'var(--color-primary)' }}>
          🚀Redo att börja?
        </h3>
        <p className="text-sm text-gray-300 mb-6 max-w-2xl mx-auto">
          Har du fler frågor? Kontakta oss på Discord eller maila hello@garageai.eu.
          Vi hjälper dig gärna att komma igång!
        </p>
        <div className="flex gap-4 justify-center flex-wrap">
          <button className="px-6 py-3 bg-green-500 hover:bg-green-600 text-black font-bold rounded-lg transition">
            💬 Gå med Discord
          </button>
          <button className="px-6 py-3 border-2 border-green-500 hover:bg-green-500/10 text-green-400 font-bold rounded-lg transition">
            📧 Kontakta oss
          </button>
        </div>
      </motion.div>
    </motion.section>
  );
};

export default FAQ;
