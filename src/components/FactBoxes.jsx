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
        <div className="fact-box">
          <h3 className="text-2xl font-bold mb-8 glow-text" style={{ color: 'var(--color-primary)' }}>
            🔒 Public AI Services & Garage AI – Olika Verktyg för Olika Behov
          </h3>
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
        <div className="fact-box">
          <h3 className="text-2xl font-bold mb-8 glow-text" style={{ color: 'var(--color-primary)' }}>
            🏠 Sverige har 2+ miljoner Garage – Din GPU-kraftcentral
          </h3>
          <div className="grid md:grid-cols-2 gap-8">
            <div>
              <p className="text-sm text-gray-400 mb-4">
                <strong>📊 SCB-statistik 2024</strong>{' '}
                <a
                  href="https://www.scb.se/hitta-statistik/sverige-i-siffror/manniskorna-i-sverige/boende-i-sverige/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="read-more-link"
                >
                  [Läs mer SCB]
                </a>
              </p>
              <ul className="space-y-2 text-sm">
                <li>✓ 2,09 miljoner småhus (42% av hushållen)</li>
                <li>✓ ~85% har garage/garagerum</li>
                <li>✓ = <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>1,7+ miljoner potentiella AI-noder</span></li>
              </ul>
            </div>
            <div>
              <p className="text-sm text-gray-400 mb-4">
                <strong>🚀 Scenario: 5000 garage aktiverade</strong>
              </p>
              <ul className="space-y-2 text-sm">
                <li>🔌 5000 × 2 RTX = 10,000 GPUs</li>
                <li>⚡ 10,000 × 1.6 kW = <span style={{ color: 'var(--color-accent)' }}>16 MW distribuerad kraft</span></li>
                <li>📈 Sverige dagligt behov: ~600 MW → <span style={{ color: 'var(--color-warning)' }}>2,7% av nationellt behov</span></li>
              </ul>
            </div>
          </div>
        </div>
      </motion.section>

      {/* Fiber & Operatörer */}
      <motion.section variants={itemVariants}>
        <div className="fact-box">
          <h3 className="text-2xl font-bold mb-8 glow-text" style={{ color: 'var(--color-primary)' }}>
            📡 Fiber för AI-inferens: Sverige är Ready
          </h3>
          <div className="grid md:grid-cols-2 gap-8">
            <div>
              <p className="text-sm text-gray-400 mb-4">
                <strong>🌐 Fiberutbyggnad 2024</strong>{' '}
                <a
                  href="https://via.tt.se/files/3236104/3870971/165801/sv"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="read-more-link"
                >
                  [Läs mer PTS]
                </a>
              </p>
              <ul className="space-y-2 text-sm">
                <li>✓ <span style={{ color: 'var(--color-primary)' }}>1 Gbit/s+:</span> 123,000 abonnemang (+28% YoY)</li>
                <li>✓ <span style={{ color: 'var(--color-primary)' }}>10 Gbit/s:</span> Tillgängligt på fibernät – perfekt för GPU-kluster!</li>
                <li>✓ <span style={{ color: 'var(--color-primary)' }}>98%</span> av befolkningen inom fiber-räckhåll</li>
              </ul>
            </div>
            <div>
              <p className="text-sm text-gray-400 mb-4">
                <strong>🏙️ Operatörer & Stadsnät</strong>
              </p>
              <ul className="space-y-2 text-sm">
                <li>🔹 <span style={{ color: 'var(--color-accent)' }}>Telia:</span> 29% marknad, 1000 Mbit/s</li>
                <li>🔹 <span style={{ color: 'var(--color-accent)' }}>Telenor:</span> Stadsnät (Trollhättan, Umeå, Sundbyberg...)</li>
                <li>🔹 <span style={{ color: 'var(--color-accent)' }}>Tele2, Com Hem, Lokala stadsnät:</span> 30+ städer</li>
              </ul>
            </div>
          </div>
        </div>
      </motion.section>

      {/* Gaming-Riggar */}
      <motion.section variants={itemVariants}>
        <div className="fact-box">
          <h3 className="text-2xl font-bold mb-8 glow-text" style={{ color: 'var(--color-primary)' }}>
            🎮 Sverige's Outnyttjade Kraft: 500k Gaming-Riggar Står Av
          </h3>
          <div className="space-y-6">
            <div>
              <p className="text-sm text-gray-400 mb-4">
                <strong>Gaming-PC Marknad Sverige</strong>{' '}
                <a
                  href="https://www.grandviewresearch.com/horizon/outlook/gaming-pc-market/sweden"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="read-more-link"
                >
                  [Läs mer]
                </a>
              </p>
              <ul className="space-y-1 text-sm">
                <li>📊 Marknad 2023: ~1,4 Mdr SEK → <span style={{ color: 'var(--color-primary)', fontWeight: 'bold' }}>500k+ gamingdatorer</span></li>
                <li>👥 60% av internetanvändare spelar – varav 65% på dator</li>
                <li>⚡ Gaming PC: 250-400W under spel, <span style={{ color: 'var(--color-accent)' }}>50-100W idle dagtid</span></li>
              </ul>
            </div>
            <div className="bg-slate-950/50 p-4 rounded border border-green-500/20">
              <p className="text-sm font-bold text-green-400 mb-3">☀️ Solenergi-Kopplingen:</p>
              <ul className="space-y-1 text-sm">
                <li>🌞 Sverige: 300 soldagar/år, 10-12h peak (11:00-15:00)</li>
                <li>🔋 Dagtid: 25 MW sol + 500k idle gaming-riggar = PERFEKT MATCH</li>
                <li>💡 Varje garage-rigg: 300W × 8h = 2.4 kWh × 300 dagar = <span style={{ color: 'var(--color-accent)' }}>720 kWh/år</span></li>
                <li>🏠 Spillvärme värmer hus vinter – ingen ombyggnad behövs</li>
              </ul>
            </div>
          </div>
        </div>
      </motion.section>
    </motion.div>
  );
};

export default FactBoxes;
