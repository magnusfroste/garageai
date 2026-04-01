import { motion } from 'framer-motion';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 }
};

const useCases = [
  {
    icon: '🏥',
    sector: 'Healthcare',
    example: 'Patient records & clinical notes analysed locally. No data leaves the building. GDPR compliant by architecture, not policy.',
  },
  {
    icon: '⚖️',
    sector: 'Legal & Finance',
    example: 'Contract review, due diligence, and financial analysis on private fine-tuned models. Zero exposure to third-party cloud logs.',
  },
  {
    icon: '🏭',
    sector: 'Manufacturing',
    example: 'Quality control, anomaly detection, and predictive maintenance — running on edge nodes inside the facility, powered by on-site solar.',
  },
  {
    icon: '🏫',
    sector: 'Education',
    example: 'AI tutoring and content generation with student data fully local. No tracking. No profiling. No Big Tech dependency in the classroom.',
  },
  {
    icon: '🏘️',
    sector: 'Municipal Services',
    example: 'Citizen services, document processing, and planning tools on sovereign regional infrastructure. Publicly owned, publicly accountable.',
  },
  {
    icon: '🛒',
    sector: 'Retail & E-commerce',
    example: 'Customer analytics and recommendation systems on proprietary models — with customer data that never crosses a national border.',
  },
];

const B2B = () => {
  return (
    <motion.section
      id="for-business"
      className="py-20 px-4 max-w-6xl mx-auto"
      variants={containerVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <motion.h2 variants={itemVariants} className="apple-heading-1 mb-4 text-center">
        For Business
      </motion.h2>
      <motion.p variants={itemVariants} className="apple-body mb-16 text-center max-w-2xl mx-auto">
        When collective inference capacity exists locally, businesses gain something hyperscalers
        can never offer: <span style={{ color: 'var(--color-primary)', fontWeight: 600 }}>sovereignty,
        privacy, and a provably smaller carbon footprint</span> — all on shared infrastructure.
      </motion.p>

      {/* VPS analogy card */}
      <motion.div variants={itemVariants} className="apple-card mb-10">
        <div className="flex flex-col md:flex-row gap-8 items-start">
          <div className="flex-1">
            <h3 className="text-xl font-black mb-3" style={{ color: 'var(--color-accent)' }}>
              The VPS Model — Applied to AI Inference
            </h3>
            <p className="text-sm leading-relaxed mb-4" style={{ color: 'var(--color-text-secondary)' }}>
              A Virtual Private Server gives you isolated, dedicated compute on shared physical hardware.
              Your workload is yours — no other tenant sees your data, your processes, or your models.
            </p>
            <p className="text-sm leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
              Garage AI applies the same principle to AI inference. Your business gets a{' '}
              <strong style={{ color: 'var(--color-primary)' }}>dedicated inference slice</strong> on
              community-owned, locally produced compute — running whatever model you choose,
              with no shared context, no cross-tenant leakage, and no data leaving your region.
            </p>
          </div>
          <div className="shrink-0 md:w-64">
            <div className="p-5 rounded-xl" style={{ background: 'rgba(0,122,255,0.07)', border: '1px solid rgba(0,122,255,0.2)' }}>
              <p className="text-xs font-black mb-4" style={{ color: 'var(--color-accent)' }}>VPS → AI Inference Mapping</p>
              {[
                ['Virtual machine', 'Inference container'],
                ['Shared CPU/RAM', 'Shared GPU cluster'],
                ['Isolated OS', 'Isolated model runtime'],
                ['Your data = yours', 'Your prompts = yours'],
                ['Any OS/software', 'Any model (open or private)'],
              ].map(([from, to], i) => (
                <div key={i} className="flex items-center gap-2 text-xs mb-2" style={{ color: 'var(--color-text-secondary)' }}>
                  <span className="flex-1 text-right" style={{ color: 'var(--color-text-muted)' }}>{from}</span>
                  <span style={{ color: 'var(--color-accent)' }}>→</span>
                  <span className="flex-1" style={{ color: 'var(--color-primary)' }}>{to}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </motion.div>

      {/* Three business benefits */}
      <motion.div variants={itemVariants} className="grid md:grid-cols-3 gap-6 mb-10">
        {[
          {
            icon: '🌱',
            title: 'Reduced Carbon Footprint',
            color: 'var(--color-primary)',
            bg: 'rgba(0,255,136,0.05)',
            border: 'rgba(0,255,136,0.2)',
            points: [
              'Inference runs on solar + V2G — renewable by default',
              'No long-haul data transmission, no cooling overhead',
              '85%+ less energy loss vs centralised data centers',
              'Reportable in ESG disclosures (Scope 2 & 3 reduction)',
              'Certified locally produced energy source',
            ],
          },
          {
            icon: '🔒',
            title: 'Private Models on Shared Infrastructure',
            color: 'var(--color-accent)',
            bg: 'rgba(0,122,255,0.05)',
            border: 'rgba(0,122,255,0.2)',
            points: [
              'Deploy proprietary fine-tuned models — not shared',
              'Run open-weight models (Llama, Mistral, Qwen, etc.)',
              'Isolated runtime: no cross-tenant model contamination',
              'Model weights stay on your dedicated slice',
              'Audit logs you control, not your vendor',
            ],
          },
          {
            icon: '🛡️',
            title: 'Compliance by Design',
            color: 'var(--color-warning)',
            bg: 'rgba(255,214,10,0.05)',
            border: 'rgba(255,214,10,0.2)',
            points: [
              'Data never crosses national or EU borders',
              'No US Cloud Act exposure — no American jurisdiction',
              'EU AI Act compliant architecture from day one',
              'GDPR: data minimisation enforced at infrastructure level',
              'No vendor lock-in, no pricing surprises',
            ],
          },
        ].map((col, i) => (
          <div key={i} className="p-6 rounded-xl" style={{ background: col.bg, border: `1px solid ${col.border}` }}>
            <div className="text-3xl mb-3">{col.icon}</div>
            <h4 className="font-bold mb-4" style={{ color: col.color }}>{col.title}</h4>
            <ul className="space-y-2">
              {col.points.map((pt, j) => (
                <li key={j} className="text-xs flex gap-2" style={{ color: 'var(--color-text-secondary)' }}>
                  <span style={{ color: col.color }}>✓</span>{pt}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </motion.div>

      {/* Use cases */}
      <motion.div variants={itemVariants} className="apple-card mb-8">
        <h3 className="apple-heading-2 mb-4 text-center">Industry Use Cases</h3>
        <p className="apple-body mb-8 text-center max-w-xl mx-auto">
          Every sector dealing with sensitive data, compliance requirements,
          or sustainability targets is a natural fit for locally produced AI inference.
        </p>
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {useCases.map((uc, i) => (
            <div
              key={i}
              className="p-5 rounded-xl"
              style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
            >
              <div className="flex items-center gap-2 mb-2">
                <span className="text-2xl">{uc.icon}</span>
                <span className="font-bold text-sm" style={{ color: 'var(--color-accent)' }}>{uc.sector}</span>
              </div>
              <p className="text-xs leading-relaxed" style={{ color: 'var(--color-text-secondary)' }}>
                {uc.example}
              </p>
            </div>
          ))}
        </div>
      </motion.div>

      {/* Cost model callout */}
      <motion.div
        variants={itemVariants}
        className="p-6 rounded-xl text-center"
        style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}
      >
        <p className="text-sm font-semibold mb-2" style={{ color: 'var(--color-primary)' }}>
          Predictable costs. No per-token pricing. No surprise invoices.
        </p>
        <p className="text-xs max-w-2xl mx-auto" style={{ color: 'var(--color-text-muted)' }}>
          Cloud AI charges per API call. A single autonomous agent workflow can generate thousands of calls per day.
          Garage AI's shared infrastructure model means capacity is reserved — not metered.
          You pay for compute, not for tokens. Your costs scale with your infrastructure, not your curiosity.
        </p>
      </motion.div>

    </motion.section>
  );
};

export default B2B;
