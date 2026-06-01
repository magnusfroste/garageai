import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 50 },
  visible: { opacity: 1, y: 0 }
};

const stackLayers = [
  {
    step: '01',
    icon: '☀️',
    title: 'Solar Panels',
    subtitle: 'Energy Production',
    details: 'Roof-mounted solar (8–12 kW) generates 40–60 kWh on a good summer day in Sweden. This is the primary energy source — powering the home, charging the battery, and running the AI node directly during daylight hours.',
    color: 'var(--color-warning)',
    footnote: '¹ Energimyndigheten: Swedish solar irradiation data 2024',
  },
  {
    step: '02',
    icon: '🔋',
    title: 'Home Battery Storage',
    subtitle: 'The Critical Link — Powers the Node 24/7',
    details: 'Home batteries (Tesla Powerwall: 13.5 kWh, BYD: 10–15 kWh) are the backbone of the energy stack. They absorb solar surplus during the day and release it at night — decoupling the AI node from solar timing entirely. The node runs around the clock regardless of weather or time of day.',
    color: 'var(--color-accent)',
    footnote: '² Tesla Powerwall 3 specs; BYD Battery Box specs',
  },
  {
    step: '03',
    icon: '🚗',
    title: 'EV Battery — Evening & Weekend Buffer',
    subtitle: 'V2H (Vehicle-to-Home), not primarily V2G',
    details: 'The honest picture: cars are away during solar peak hours on weekdays. But they\'re home during evenings and weekends — exactly when V2H (Vehicle-to-Home) adds value. A 60–100 kWh EV battery parked at home from 17:00 can supply the house and AI node overnight. Studies show cars are parked ~92% of the time; roughly half of that is at home, concentrated in evenings and weekends.',
    color: 'rgba(255,59,48,0.9)',
    footnote: '³ INRIX / IEA: average car parked 92–95% of time. IEA Global EV Outlook 2024 on V2H deployment.',
  },
  {
    step: '04',
    icon: '🖥️',
    title: 'AI Inference Node',
    subtitle: 'The Computing Layer',
    details: 'A Mac Mini M4 Pro, RTX 4090 rig, or purpose-built inference server in your garage. Runs local models (Llama, Mistral, Qwen) at 15–120 tokens/sec. Consumes 300–600W.',
    color: 'var(--color-primary)',
  },
  {
    step: '05',
    icon: '🌐',
    title: 'Fiber Connection',
    subtitle: 'Distribution',
    details: 'Sweden\'s 1–10 Gbit/s fiber network connects your node to neighbours, local businesses, and the broader Garage AI mesh. Inference stays local — bandwidth is just for coordination.',
    color: 'rgba(0,122,255,0.9)',
  },
];

const Infrastructure = () => {
  return (
    <motion.section
      id="infrastructure"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2
        variants={itemVariants}
        className="apple-heading-1 mb-4 text-center"
      >
        The Garage Node
      </motion.h2>
      <motion.p
        variants={itemVariants}
        className="apple-body mb-16 text-center max-w-2xl mx-auto"
      >
        Five layers. One garage. A complete, self-sustaining AI infrastructure node
        that powers your home, your neighbourhood, and Europe's sovereign AI future.
      </motion.p>

      {/* Stack diagram */}
      <motion.div variants={itemVariants} className="mb-16">
        <div className="space-y-3 max-w-4xl mx-auto">
          {stackLayers.map((layer, i) => (
            <div
              key={i}
              className="flex items-start gap-5 p-5 rounded-xl transition-all"
              style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid rgba(255,255,255,0.07)` }}
            >
              <div className="text-center shrink-0 w-12">
                <div className="text-3xl mb-1">{layer.icon}</div>
                <div className="text-xs font-black" style={{ color: layer.color, opacity: 0.5 }}>{layer.step}</div>
              </div>
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-2 mb-1">
                  <h4 className="font-bold" style={{ color: layer.color }}>{layer.title}</h4>
                  <span className="text-xs px-2 py-0.5 rounded" style={{ background: `${layer.color}20`, color: layer.color }}>
                    {layer.subtitle}
                  </span>
                </div>
                <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
                  {layer.details}
                </p>
                {layer.footnote && (
                  <p className="text-xs mt-2" style={{ color: 'var(--color-text-muted)', opacity: 0.6 }}>
                    {layer.footnote}
                  </p>
                )}
              </div>
            </div>
          ))}
        </div>
      </motion.div>

      {/* Energy flow — honest diagram */}
      <motion.div variants={itemVariants} className="apple-card mb-8">
        <h3 className="apple-heading-2 mb-4 text-center">How the Energy Actually Flows</h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          The Powerwall is the critical link — it decouples solar production from consumption.
          The EV adds evening and weekend capacity via V2H.
        </p>
        <div className="overflow-x-auto">
          <div className="min-w-[560px] space-y-3 max-w-2xl mx-auto text-sm">
            {[
              { time: 'Daytime (solar peak)', flow: 'Solar → Home use + Powerwall charge + AI node direct', note: 'Car typically away — EV not in the loop', ok: true },
              { time: 'Daytime (overcast)', flow: 'Powerwall → AI node', note: 'Node runs uninterrupted from stored energy', ok: true },
              { time: 'Evening (car home)', flow: 'EV V2H → Home use + AI node buffer', note: 'EV discharges to home via V2H overnight', ok: true },
              { time: 'Weekends / WFH days', flow: 'Solar → Powerwall + EV charge + AI node', note: 'Car at home: full stack active simultaneously', ok: true },
              { time: 'Deep winter nights', flow: 'Powerwall + EV V2H + grid backup', note: 'Grid as fallback — node stays online', ok: true },
            ].map((row, i) => (
              <div key={i} className="flex flex-col sm:flex-row gap-2 p-3 rounded-lg" style={{ background: 'rgba(255,255,255,0.03)' }}>
                <span className="shrink-0 text-xs font-bold w-44" style={{ color: 'var(--color-warning)' }}>{row.time}</span>
                <span className="flex-1 font-medium" style={{ color: 'var(--color-primary)' }}>{row.flow}</span>
                <span className="text-xs italic" style={{ color: 'var(--color-text-muted)' }}>{row.note}</span>
              </div>
            ))}
          </div>
        </div>
        <p className="text-xs text-center mt-6" style={{ color: 'var(--color-text-muted)', opacity: 0.6 }}>
          V2G (Vehicle-to-Grid) = selling back to the public grid — still limited deployment in Sweden.
          V2H (Vehicle-to-Home) = powering your own home from the EV battery — available today on most major EVs.
          Sources: IEA Global EV Outlook 2024; Energimyndigheten; Tesla Powerwall specs.
        </p>
      </motion.div>

      {/* Locally produced, locally consumed */}
      <motion.div variants={itemVariants} className="apple-card mb-8">
        <h3 className="apple-heading-2 mb-4 text-center">
          Locally Produced. Locally Consumed.
        </h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          Nearshoring compute has the same advantages as nearshoring food. Lower latency, lower
          carbon, stronger community resilience — and the value stays in the region.
        </p>
        <div className="grid md:grid-cols-3 gap-6">
          {[
            {
              icon: '⚡',
              title: 'Latency',
              points: [
                'Local inference: &lt;1ms roundtrip',
                'Cloud inference: 50–300ms',
                'For real-time agents: local wins',
                'No transatlantic bottleneck',
              ],
              color: 'var(--color-warning)'
            },
            {
              icon: '🌱',
              title: 'Carbon & Cost',
              points: [
                'Runs on own solar energy',
                'No data center cooling overhead',
                '85%+ less transmission loss',
                'Near-zero marginal cost',
              ],
              color: 'var(--color-primary)'
            },
            {
              icon: '🛡️',
              title: 'Sovereignty',
              points: [
                'Data never leaves the garage',
                'No US Cloud Act exposure',
                'GDPR compliance by design',
                'Operational during outages',
              ],
              color: 'var(--color-accent)'
            },
          ].map((col, i) => (
            <div key={i} className="p-5 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}>
              <div className="text-3xl mb-2">{col.icon}</div>
              <h4 className="font-bold mb-3" style={{ color: col.color }}>{col.title}</h4>
              <ul className="space-y-2">
                {col.points.map((pt, j) => (
                  <li key={j} className="text-xs flex gap-2" style={{ color: 'var(--color-text-secondary)' }}>
                    <span style={{ color: col.color }}>✓</span>
                    <span dangerouslySetInnerHTML={{ __html: pt }} />
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </motion.div>

      {/* From home to neighbourhood */}
      <motion.div variants={itemVariants} className="apple-card">
        <h3 className="apple-heading-2 mb-4 text-center">From Your Garage to the Neighbourhood</h3>
        <p className="apple-body mb-8 text-center max-w-2xl mx-auto">
          Once your home's own AI needs are covered — the surplus is a gift to the community.
          Small businesses, schools, neighbours: all get access to private, fast, local AI.
        </p>
        <div className="grid md:grid-cols-2 gap-8">
          <div className="p-6 rounded-xl" style={{ background: 'rgba(0,255,136,0.05)', border: '1px solid rgba(0,255,136,0.2)' }}>
            <p className="font-semibold mb-4" style={{ color: 'var(--color-primary)' }}>🏠 Home First</p>
            <ul className="space-y-2 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
              {['Security cameras (local face detection)', 'Smart heating (AI-optimized schedules)', 'Garden automation (weather + soil sensors)', 'Kitchen assistant (offline voice)', 'Energy management (solar + EV + grid)'].map((item, i) => (
                <li key={i} className="flex gap-2"><span style={{ color: 'var(--color-primary)' }}>→</span>{item}</li>
              ))}
            </ul>
          </div>
          <div className="p-6 rounded-xl" style={{ background: 'rgba(0,122,255,0.05)', border: '1px solid rgba(0,122,255,0.2)' }}>
            <p className="font-semibold mb-4" style={{ color: 'var(--color-accent)' }}>🏘️ Neighbourhood Surplus</p>
            <ul className="space-y-2 text-sm" style={{ color: 'var(--color-text-secondary)' }}>
              {['Local bakery: AI for inventory & ordering', 'Dentist clinic: private document analysis', 'School: AI tutor with no tracking', 'Neighbours: shared private assistant', 'Local business: AI without Big Tech dependency'].map((item, i) => (
                <li key={i} className="flex gap-2"><span style={{ color: 'var(--color-accent)' }}>→</span>{item}</li>
              ))}
            </ul>
          </div>
        </div>
        <div className="mt-6 p-4 rounded-xl text-center" style={{ background: 'rgba(255,214,10,0.07)', border: '1px solid rgba(255,214,10,0.2)' }}>
          <p className="text-sm font-semibold" style={{ color: 'var(--color-warning)' }}>
            The garage is the new server room. Decentralized. Sovereign. Community-owned.
          </p>
        </div>
      </motion.div>

    </motion.section>
  );
};

export default Infrastructure;
