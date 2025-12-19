import { motion } from 'framer-motion';
import { useState } from 'react';

const Hero = () => {
  const [gpuInfo, setGpuInfo] = useState(null);
  const [isChecking, setIsChecking] = useState(false);

  const checkGPUCapabilities = async () => {
    setIsChecking(true);

    try {
      // Check WebGPU support
      if ('gpu' in navigator) {
        const adapter = await navigator.gpu.requestAdapter();
        if (adapter) {
          const device = await adapter.requestDevice();
          const info = adapter.info || {};

          setGpuInfo({
            webgpu: true,
            vendor: info.vendor || 'Unknown',
            architecture: info.architecture || 'Unknown',
            device: info.device || 'Unknown',
            compatible: true,
            message: '🎉 Din webbläsare stödjer WebGPU! Du kan köra AI-modeller lokalt.'
          });
        } else {
          setGpuInfo({
            webgpu: false,
            compatible: false,
            message: '❌ Ingen GPU detekterad via WebGPU. Prova med en modern GPU.'
          });
        }
      } else if (typeof WebGLRenderingContext !== 'undefined') {
        // Fallback to WebGL detection
        const canvas = document.createElement('canvas');
        const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');

        if (gl) {
          const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
          const renderer = debugInfo ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : 'Unknown';

          setGpuInfo({
            webgpu: false,
            webgl: true,
            renderer: renderer,
            compatible: true,
            message: '✅ WebGL detekterad! Grundläggande GPU-stöd finns. För bästa prestanda rekommenderar vi WebGPU-kompatibla enheter.'
          });
        } else {
          setGpuInfo({
            webgpu: false,
            webgl: false,
            compatible: false,
            message: '❌ Varken WebGPU eller WebGL stödjs. Uppdatera din webbläsare eller använd en modern enhet.'
          });
        }
      } else {
        setGpuInfo({
          compatible: false,
          message: '❌ Ingen GPU-detektion möjlig. Använd en modern webbläsare med GPU-stöd.'
        });
      }
    } catch (error) {
      setGpuInfo({
        compatible: false,
        message: `❌ Fel vid GPU-detektion: ${error.message}`
      });
    }

    setIsChecking(false);
  };

  const startNodeOnboarding = () => {
    if (!gpuInfo) {
      checkGPUCapabilities();
    } else {
      // Show next steps based on GPU compatibility
      const message = gpuInfo.compatible
        ? `🚀 Bra! Din enhet har GPU-stöd.\n\nNästa steg:\n1. Gå med i vårt GitHub-community\n2. Ladda ner Garage AI-node mjukvara\n3. Konfigurera din första AI-nod\n\nVill du öppna GitHub Discussions nu?`
        : `⚠️ Din enhet behöver bättre GPU-stöd för optimal prestanda.\n\nAlternativ:\n1. Uppgradera till RTX 30/40-serie GPU\n2. Använd en annan enhet med bättre GPU\n3. Bidra till community-utveckling istället\n\nÖppna GitHub Discussions för mer info?`;

      if (confirm(message)) {
        window.open('https://github.com/magnusfroste/garageai/discussions', '_blank');
      }
    }
  };

  return (
    <section className="hero-gradient min-h-screen flex items-center justify-center pt-20 relative overflow-hidden" style={{ background: 'var(--color-bg-primary)' }}>
      {/* Subtle Apple-style background */}
      <div className="absolute inset-0 opacity-30">
        <div
          style={{
            position: 'absolute',
            top: '20%',
            left: '10%',
            width: '200px',
            height: '200px',
            background: 'radial-gradient(circle, rgba(0, 122, 255, 0.1) 0%, transparent 70%)',
            borderRadius: '50%'
          }}
        />
        <div
          style={{
            position: 'absolute',
            top: '60%',
            right: '15%',
            width: '150px',
            height: '150px',
            background: 'radial-gradient(circle, rgba(52, 199, 89, 0.1) 0%, transparent 70%)',
            borderRadius: '50%'
          }}
        />
      </div>

      <div className="relative z-10 text-center max-w-4xl mx-auto px-4">
        <motion.h1
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          className="apple-heading-1 mb-8"
        >
          GARAGE AI
        </motion.h1>

        <motion.h2
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="apple-heading-2 mb-8"
        >
          Din garage → Europas AI-motor
        </motion.h2>

        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.4 }}
          className="apple-body mb-6 max-w-2xl mx-auto"
        >
          Demokratisera AI genom decentraliserad infrastruktur. Var med och bygg Europas mest hållbara AI-nätverk -
          drivet av gaming-datorer och förnybar energi från svenska hem.
        </motion.p>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.5 }}
          className="mb-8"
        >
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto">
            <div className="text-center">
              <div className="text-2xl mb-1">🚀</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>1.7M+</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>Potentiella noder</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">⚡</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>120 tokens/sek</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>per RTX 4090</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">🏆</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>98%</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>fiber-täckning</div>
            </div>
            <div className="text-center">
              <div className="text-2xl mb-1">🌱</div>
              <div className="text-sm font-medium" style={{ color: 'var(--color-text-primary)' }}>100%</div>
              <div className="text-xs" style={{ color: 'var(--color-text-muted)' }}>lokal integritet</div>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="mb-12"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-slate-950 border border-slate-700 rounded-full text-green-400 text-sm font-medium">
            <span>🔓</span>
            <span>100% Open Source & MIT Licensed</span>
          </div>
        </motion.div>

        {/* GPU Detection Status */}
        {gpuInfo && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className={`mb-6 px-4 py-2 rounded-lg text-sm ${
              gpuInfo.compatible
                ? 'bg-green-500/10 border border-green-500/30 text-green-400'
                : 'bg-yellow-500/10 border border-yellow-500/30 text-yellow-400'
            }`}
          >
            {gpuInfo.message}
          </motion.div>
        )}

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.6 }}
          className="flex gap-4 justify-center flex-wrap"
        >
          <motion.button
            onClick={startNodeOnboarding}
            disabled={isChecking}
            className="apple-button-primary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            {isChecking ? '🔍 Kontrollerar GPU...' : '🚀 Starta min nod'}
          </motion.button>

          <motion.button
            onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            💬 Community
          </motion.button>

          <motion.button
            className="apple-button-secondary"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            📚 Teknisk info
          </motion.button>
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
