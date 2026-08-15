import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

// EU Garage data by country (estimates based on housing stock data, Eurostat 2023)
const euGarageData = [
  { country: '🇩🇪 Germany',   homes: '41.5M', garages: '18M',  pct: '43%', evs: '3.5M' },
  { country: '🇫🇷 France',    homes: '36M',   garages: '11M',  pct: '31%', evs: '1.8M' },
  { country: '🇮🇹 Italy',     homes: '26M',   garages: '9M',   pct: '35%', evs: '0.6M' },
  { country: '🇪🇸 Spain',     homes: '22M',   garages: '7M',   pct: '32%', evs: '0.5M' },
  { country: '🇵🇱 Poland',    homes: '14M',   garages: '5M',   pct: '36%', evs: '0.1M' },
  { country: '🇸🇪 Sweden',    homes: '4.8M',  garages: '1.7M', pct: '35%', evs: '1.5M' },
  { country: '🇳🇱 Netherlands', homes: '8M',  garages: '2M',   pct: '25%', evs: '0.5M' },
  { country: '🇧🇪 Belgium',   homes: '5.5M',  garages: '2M',   pct: '36%', evs: '0.3M' },
];

const whySwedenDefault = [
  {
    icon: '🌐',
    title: 'World-Class Fiber',
    stats: [
      '98% fiber coverage',
      '1–10 Gbit/s to homes',
      '+28% YoY in 1Gbit subscriptions',
      'PTS Report 2024'
    ],
    color: 'var(--color-primary)'
  },
  {
    icon: '⚡',
    title: 'EV Leadership',
    stats: [
      '1.5M+ EVs & PHEVs on road',
      '50%+ of new car sales electric',
      'Average 60kWh battery = local storage',
      'V2G-ready grid infrastructure'
    ],
    color: 'var(--color-accent)'
  },
  {
    icon: '☀️',
    title: 'Solar Momentum',
    stats: [
      '300,000+ homes with solar panels',
      '+25% YoY installation growth',
      '5GW+ total installed capacity',
      'Source: Energimyndigheten 2024'
    ],
    color: 'var(--color-warning)'
  }
];

const gridElectricityDefault = [
  'Power plants were the only source of electricity',
  'Solar panels arrived — anyone could produce energy',
  'Excess production feeds back into the grid',
  'EVs become mobile battery storage (V2G)',
  'Today: 600GW solar installed in EU alone',
];

const gridComputingDefault = [
  'Hyperscale datacenters are the only source of AI compute',
  'Consumer hardware (GPU/NPU) now rivals data center performance',
  'Local models run privately, excess capacity shared',
  'A Mac Mini in every home — the "Mac Mini moment"',
  'Tomorrow: your garage feeds the neighbourhood',
];

const gridQuoteDefault = '"The technology is here. The mass scale-up of locally produced electricity is a clear example. Computing — or AI inference, same thing — is following the exact same path."';

const evParkingDefault = [
  { label: 'Car parked (all time)', value: '92–95% of the time¹' },
  { label: 'Of that: parked at home', value: '~50%, mostly evenings & weekends²' },
  { label: 'Of that: at workplace', value: '~25%' },
  { label: 'Solar peak (09:00–15:00 weekdays)', value: 'Car typically away from home' },
  { label: 'Evenings 17:00–08:00', value: 'Car home → V2H fully available' },
  { label: 'Weekends / WFH days', value: 'Full solar + EV stack active' },
];

const energyStackDefault = [
  { label: 'Solar panels (8–12 kW)³', value: '40–60 kWh on a good day' },
  { label: 'Home battery — PRIMARY link', value: 'Powers node 24/7, day & night' },
  { label: 'Powerwall 3 / BYD Battery Box', value: '13.5–15 kWh buffer capacity' },
  { label: 'EV V2H (when home, evenings)', value: '60–100 kWh additional buffer' },
  { label: 'AI inference hardware', value: '300–600W continuous draw' },
  { label: 'Net energy cost (solar)', value: 'Near zero — runs on own production' },
];

const euScaleDefault = [
  { label: '17M EVs in EU today⁴', value: '= 1+ TWh theoretical storage' },
  { label: 'EU target: 100M EVs by 2035⁴', value: '= 6 TWh at 60 kWh avg' },
  { label: 'Sweden: 1.5M EVs⁵', value: '= 90 GWh potential' },
  { label: 'WFH rate Sweden ~35%⁶', value: 'boosts daytime V2H availability' },
];

const evFootnotesDefault = [
  '¹ INRIX Global Traffic Scorecard; RAC Foundation: "Spaced Out" report on vehicle utilisation',
  '² Transport for London / European Environment Agency: vehicle location studies',
  '³ Energimyndigheten: Solar irradiation in Sweden (solenergistatistik 2024)',
  '⁴ ACEA Electric Vehicle Report 2024; IEA Global EV Outlook 2024',
  '⁵ Trafikanalys / BilSweden: Fordonsstatistik 2024',
  '⁶ Eurostat: share of employees working from home, Sweden 2023',
];

const FactBoxes = ({
  title = "Europe's Hidden Data Centers",
  description = '75+ million garages across the EU sit idle — waiting to become the backbone of a sovereign, decentralized AI infrastructure.',
  euData,
  totalRow = { homes: '~225M', garages: '~75M+', pct: '~33%', evs: '~17M' },
  sourceNote = 'Sources: Eurostat Housing Statistics 2023, ACEA Electric Vehicle Report 2024. Garage estimates based on single-family home stock.',
  whySweden,
  gridElectricity,
  gridComputing,
  gridQuote,
  evParking,
  energyStack,
  euScale,
  evFootnotes,
} = {}) => {
  const rows = euData || euGarageData;
  const whySwedenItems = whySweden || whySwedenDefault;
  const gridElectricityItems = gridElectricity || gridElectricityDefault;
  const gridComputingItems = gridComputing || gridComputingDefault;
  const gridQuoteText = gridQuote || gridQuoteDefault;
  const evParkingItems = evParking || evParkingDefault;
  const energyStackItems = energyStack || energyStackDefault;
  const euScaleItems = euScale || euScaleDefault;
  const evFootnoteItems = evFootnotes || evFootnotesDefault;
  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
      className="py-16 px-4 max-w-6xl mx-auto space-y-16"
      id="eu-analysis"
    >

      {/* EU Garage Analysis */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-4 text-center">
            {title}
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            {description}
          </p>

          <div className="overflow-x-auto mb-8">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-green-500/30">
                  <th className="text-left py-3 px-3" style={{ color: 'var(--color-text-muted)' }}>Country</th>
                  <th className="text-right py-3 px-3" style={{ color: 'var(--color-text-muted)' }}>Homes</th>
                  <th className="text-right py-3 px-3" style={{ color: 'var(--color-primary)' }}>Est. Garages</th>
                  <th className="text-right py-3 px-3" style={{ color: 'var(--color-text-muted)' }}>Coverage</th>
                  <th className="text-right py-3 px-3" style={{ color: 'var(--color-accent)' }}>EVs on Road</th>
                </tr>
              </thead>
              <tbody>
                {(rows || []).map((row, i) => (
                  <tr key={i} className="border-b border-slate-800 hover:bg-white/2 transition">
                    <td className="py-3 px-3 font-medium">{row.country}</td>
                    <td className="py-3 px-3 text-right" style={{ color: 'var(--color-text-secondary)' }}>{row.homes}</td>
                    <td className="py-3 px-3 text-right font-bold" style={{ color: 'var(--color-primary)' }}>{row.garages}</td>
                    <td className="py-3 px-3 text-right" style={{ color: 'var(--color-text-muted)' }}>{row.pct}</td>
                    <td className="py-3 px-3 text-right" style={{ color: 'var(--color-accent)' }}>{row.evs}</td>
                  </tr>
                ))}
                <tr className="border-t-2 border-green-500/40">
                  <td className="py-3 px-3 font-black" style={{ color: 'var(--color-primary)' }}>EU-27 TOTAL</td>
                  <td className="py-3 px-3 text-right font-bold">{totalRow.homes}</td>
                  <td className="py-3 px-3 text-right font-black text-lg" style={{ color: 'var(--color-primary)' }}>{totalRow.garages}</td>
                  <td className="py-3 px-3 text-right">{totalRow.pct}</td>
                  <td className="py-3 px-3 text-right font-bold" style={{ color: 'var(--color-accent)' }}>{totalRow.evs}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p className="text-xs text-center" style={{ color: 'var(--color-text-muted)' }}>
            {sourceNote}
          </p>
        </div>
      </motion.section>

      {/* Why Sweden First */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-4 text-center">
            Why Sweden Takes the Lead
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            Sweden doesn't just have the vision — it has every infrastructure advantage needed to make this real.
          </p>
          <div className="grid md:grid-cols-3 gap-6">
            {(whySwedenItems || []).map((item, i) => (
              <div key={i} className="p-5 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
                <div className="text-3xl mb-3">{item.icon}</div>
                <h4 className="font-bold mb-3" style={{ color: item.color }}>{item.title}</h4>
                <ul className="space-y-2">
                  {(item.stats || []).map((s, j) => (
                    <li key={j} className="text-xs flex gap-2" style={{ color: 'var(--color-text-secondary)' }}>
                      <span style={{ color: item.color }}>✓</span> {s}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>
      </motion.section>

      {/* The Electricity Grid Parallel */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-4 text-center">
            We've Seen This Before
          </h3>
          <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
            Centralization has dominated industries for decades. Then came local electricity production.
            Computing is next — and the shift has already begun.
          </p>

          <div className="grid md:grid-cols-2 gap-8 mb-8">
            <div className="p-6 rounded-xl" style={{ background: 'rgba(255,200,0,0.06)', border: '1px solid rgba(255,200,0,0.2)' }}>
              <h4 className="font-bold mb-4 text-lg" style={{ color: 'var(--color-warning)' }}>
                ⚡ Electricity: The Precedent
              </h4>
              <ul className="space-y-3 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
                {(gridElectricityItems || []).map((text, i) => (
                  i === (gridElectricityItems || []).length - 1 ? (
                    <li key={i} className="flex gap-3 font-semibold" style={{ color: 'var(--color-warning)' }}>
                      <span>→</span> {text}
                    </li>
                  ) : (
                    <li key={i} className="flex gap-3"><span style={{ color: 'var(--color-warning)' }}>→</span> {text}</li>
                  )
                ))}
              </ul>
            </div>
            <div className="p-6 rounded-xl" style={{ background: 'rgba(0,255,136,0.06)', border: '1px solid rgba(0,255,136,0.2)' }}>
              <h4 className="font-bold mb-4 text-lg" style={{ color: 'var(--color-primary)' }}>
                🧠 Computing: The Same Shift
              </h4>
              <ul className="space-y-3 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
                {(gridComputingItems || []).map((text, i) => (
                  i === (gridComputingItems || []).length - 1 ? (
                    <li key={i} className="flex gap-3 font-semibold" style={{ color: 'var(--color-primary)' }}>
                      <span>→</span> {text}
                    </li>
                  ) : (
                    <li key={i} className="flex gap-3"><span style={{ color: 'var(--color-primary)' }}>→</span> {text}</li>
                  )
                ))}
              </ul>
            </div>
          </div>

          <div className="p-5 rounded-xl text-center" style={{ background: 'rgba(0,122,255,0.07)', border: '1px solid rgba(0,122,255,0.2)' }}>
            <p className="text-sm font-semibold" style={{ color: 'var(--color-accent)' }}>
              {gridQuoteText}
            </p>
          </div>
        </div>
      </motion.section>

      {/* EV + V2H Power — honest framing */}
      <motion.section variants={itemVariants}>
        <div className="apple-card">
          <h3 className="apple-heading-2 mb-4 text-center">
            EV Batteries: Evening & Weekend Buffer
          </h3>
          <p className="apple-body mb-6 text-center max-w-2xl mx-auto">
            The honest picture: EVs are mostly away during solar peak hours on weekdays.
            But they're home during evenings and weekends — and that's when V2H (Vehicle-to-Home)
            adds real capacity to the stack. The <strong>home battery is the 24/7 backbone</strong>;
            the EV is an important complement.
          </p>

          <div className="grid md:grid-cols-2 gap-8 mb-6">
            <div>
              <p className="font-semibold mb-4" style={{ color: 'var(--color-warning)' }}>🚗 EV Parking Reality (EU average)</p>
              <ul className="space-y-3">
                {(evParkingItems || []).map(({ label, value }, i) => (
                  <li key={i} className="flex justify-between text-sm border-b border-slate-800 pb-2">
                    <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
                    <span className="font-bold text-right" style={{ color: 'var(--color-warning)' }}>{value}</span>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <p className="font-semibold mb-4" style={{ color: 'var(--color-primary)' }}>🏠 Garage Node Energy Stack</p>
              <ul className="space-y-3">
                {(energyStackItems || []).map(({ label, value }, i) => (
                  <li key={i} className="flex justify-between text-sm border-b border-slate-800 pb-2">
                    <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
                    <span className="font-bold text-right" style={{ color: 'var(--color-primary)' }}>{value}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          {/* EU EV scale */}
          <div className="p-5 rounded-xl mb-4" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
            <p className="font-semibold mb-3 text-sm" style={{ color: 'var(--color-accent)' }}>🔋 EU Scale — The Potential is Still Massive</p>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center text-xs">
              {(euScaleItems || []).map(({ label, value }, i) => (
                <div key={i} className="p-3 rounded-lg" style={{ background: 'rgba(0,122,255,0.07)' }}>
                  <div className="font-bold mb-1" style={{ color: 'var(--color-accent)' }}>{label}</div>
                  <div style={{ color: 'var(--color-text-muted)' }}>{value}</div>
                </div>
              ))}
            </div>
          </div>

          <div className="p-4 rounded-lg text-center" style={{ background: 'rgba(52,199,89,0.08)', border: '1px solid rgba(52,199,89,0.2)' }}>
            <p className="text-sm" style={{ color: 'var(--color-primary)' }}>
              <strong>The stack works.</strong> Solar + home battery powers the AI node around the clock.
              The EV adds meaningful evening and weekend capacity via V2H.
              <strong> Locally produced. Locally consumed. Locally owned.</strong>
            </p>
          </div>

          {/* Footnotes */}
          <div className="mt-6 pt-4 border-t border-slate-800 space-y-1">
            {(evFootnoteItems || []).map((note, i) => (
              <p key={i} className="text-xs" style={{ color: 'var(--color-text-muted)', opacity: 0.55 }}>{note}</p>
            ))}
          </div>
        </div>
      </motion.section>

    </motion.div>
  );
};

export default FactBoxes;
