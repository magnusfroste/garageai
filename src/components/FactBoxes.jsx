import { motion } from 'framer-motion';

const FactBoxes = () => {
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

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
      className="py-16 px-4 max-w-6xl mx-auto space-y-16"
    >
      {/* Public AI Services vs Garage AI */}
      <motion.section variants={itemVariants} id="comparison">
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-12 text-center">
            Public AI Services & Garage AI
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            Olika verktyg för olika behov – jämför traditionella AI-tjänster med Garage AI
          </p>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <tr className="border-b border-green-500/30">
                <th className="text-left py-3 px-4" style={{ color: 'var(--color-warning)' }}>Public AI Services</th>
                <th className="text-left py-3 px-4" style={{ color: 'var(--color-primary)' }}>Garage AI</th>
              </tr>
              <tr className="border-b border-green-500/30">
                <td className="py-3 px-4">
                  ✅ Perfekt för <span style={{ color: 'var(--color-primary)' }}>snabb prototyping & experiment</span><br />
                  ✅ Enkelt att komma igång utan hårdvara<br />
                  ✅ Stora modeller tillgängliga direkt<br />
                  ✅ Betrodda av miljontals användare<br />
                  ⚠️ Data lämnar din kontroll under bearbetning
                  <a
                    href="https://techcommunity.microsoft.com/blog/azureconfidentialcomputingblog/azure-ai-confidential-inferencing-technical-deep-dive/42"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="read-more-link text-xs"
                  >
                    {' '}[Läs mer Azure]
                  </a>
                </td>
                <td className="py-3 px-4">
                  ✅ <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>100% lokal kontroll & integritet</span><br />
                  ✅ <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>Ingen dataström till externa servrar</span><br />
                  ✅ Miljövänligt – använder befintlig hårdvara<br />
                  ✅ Skalbar genom grannskapsnätverk<br />
                  ✅ Ekonomiskt hållbart långsiktigt
                </td>
              </tr>
              <tr className="border-b border-green-500/30">
                <td className="py-3 px-4">Global infrastruktur – tillgänglig överallt</td>
                <td className="py-3 px-4">Lokal infrastruktur – byggd på svenska förhållanden</td>
              </tr>
              <tr>
                <td className="py-3 px-4">Flexibel prissättning per användning</td>
                <td className="py-3 px-4">Kostnadseffektivt genom delad hårdvara</td>
              </tr>
            </table>
          </div>
        </div>
      </motion.section>

      {/* Sverige's Garage Statistics */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-8 text-center">
            Sveriges Garage-Potential
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            2+ miljoner garage väntar på att bli AI-infrastruktur
          </p>
          <div className="grid md:grid-cols-2 gap-8">
            <div>
              <p className="apple-caption mb-4">
                <strong>📊 SCB-statistik 2024</strong>{' '}
                <a
                  href="https://www.scb.se/hitta-statistik/sverige-i-siffror/manniskorna-i-sverige/boende-i-sverige/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  [Källa]
                </a>
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-3">
                  <span className="text-green-500">✓</span>
                  <span className="apple-caption">2,09 miljoner småhus (42% av hushållen)</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-green-500">✓</span>
                  <span className="apple-caption">~85% har garage/garagerum</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-green-500">✓</span>
                  <span className="apple-caption">= <span style={{ color: 'var(--color-primary)', fontWeight: '600' }}>1,7+ miljoner potentiella AI-noder</span></span>
                </li>
              </ul>
            </div>
            <div>
              <p className="apple-caption mb-4">
                <strong>🚀 Scenario: 5000 garage aktiverade</strong>
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-3">
                  <span className="text-lg">🔌</span>
                  <span className="apple-caption">5000 × 2 RTX = 10,000 GPUs</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-lg">⚡</span>
                  <span className="apple-caption">10,000 × 1.6 kW = <span style={{ color: 'var(--color-accent)', fontWeight: '600' }}>16 MW distribuerad kraft</span></span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-lg">📈</span>
                  <span className="apple-caption">Sveriges behov: ~600 MW → <span style={{ color: 'var(--color-secondary)', fontWeight: '600' }}>2,7% täckning</span></span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </motion.section>

      {/* Fiber & Operatörer */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-8 text-center">
            Sveriges Fiber-Infrastruktur
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            AI-inferens kräver hög bandbredd – Sverige är redo med världsklass fiber
          </p>
          <div className="grid md:grid-cols-2 gap-8">
            <div>
              <p className="apple-caption mb-4">
                <strong>🌐 Fiberutbyggnad 2024</strong>{' '}
                <a
                  href="https://via.tt.se/files/3236104/3870971/165801/sv"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  [PTS Källa]
                </a>
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-3">
                  <span className="text-green-500">✓</span>
                  <span className="apple-caption"><span style={{ color: 'var(--color-primary)', fontWeight: '600' }}>1 Gbit/s+:</span> 123,000 abonnemang (+28% YoY)</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-green-500">✓</span>
                  <span className="apple-caption"><span style={{ color: 'var(--color-primary)', fontWeight: '600' }}>10 Gbit/s:</span> Tillgängligt på fibernät</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-green-500">✓</span>
                  <span className="apple-caption"><span style={{ color: 'var(--color-primary)', fontWeight: '600' }}>98%</span> av befolkningen har fiberåtkomst</span>
                </li>
              </ul>
            </div>
            <div>
              <p className="apple-caption mb-4">
                <strong>🏙️ Operatörer & Stadsnät</strong>
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-3">
                  <span className="text-lg">🔹</span>
                  <span className="apple-caption"><span style={{ color: 'var(--color-accent)', fontWeight: '600' }}>Telia:</span> 29% marknad, 1000 Mbit/s</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-lg">🔹</span>
                  <span className="apple-caption"><span style={{ color: 'var(--color-accent)', fontWeight: '600' }}>Telenor:</span> Stadsnät i 20+ städer</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-lg">🔹</span>
                  <span className="apple-caption"><span style={{ color: 'var(--color-accent)', fontWeight: '600' }}>Tele2, Com Hem:</span> Nationell täckning</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </motion.section>

      {/* Gaming-Riggar */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-8 text-center">
            Sveriges Outnyttjade Kraft
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            500k gaming-datorer står av dagtid – redo att bli AI-infrastruktur
          </p>
          <div className="grid md:grid-cols-1 gap-6">
            <div>
              <p className="apple-caption mb-4">
                <strong>Gaming-PC Marknad Sverige</strong>{' '}
                <a
                  href="https://www.grandviewresearch.com/horizon/outlook/gaming-pc-market/sweden"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  [Källa]
                </a>
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-3">
                  <span className="text-lg">📊</span>
                  <span className="apple-caption">Marknad 2023: ~1,4 Mdr SEK → <span style={{ color: 'var(--color-primary)', fontWeight: '600' }}>500k+ gamingdatorer</span></span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-lg">👥</span>
                  <span className="apple-caption">60% av internetanvändare spelar – 65% på dator</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-lg">⚡</span>
                  <span className="apple-caption">Gaming PC: 250-400W spel, <span style={{ color: 'var(--color-accent)', fontWeight: '600' }}>50-100W idle dagtid</span></span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </motion.section>
    </motion.div>
  );
};

export default FactBoxes;
