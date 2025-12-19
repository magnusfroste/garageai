import { useState } from 'react';
import { motion } from 'framer-motion';

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);



  const scrollToSection = (sectionId) => {
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
      setIsMenuOpen(false);
    }
  };

  return (
    <motion.nav
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.5 }}
      className="fixed top-0 w-full z-50"
      style={{
        background: 'rgba(10, 10, 10, 0.9)',
        backdropFilter: 'blur(20px)',
        borderBottom: '1px solid var(--color-border)'
      }}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex justify-between items-center">
        <motion.div
          className="text-xl font-bold glow-text"
          style={{ color: 'var(--color-primary)' }}
          whileHover={{ scale: 1.05 }}
        >
          � GARAGE AI
        </motion.div>

        {/* Desktop Menu */}
        <div className="hidden md:flex gap-8 text-sm">
          <button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Hem</button>
          <button onClick={() => scrollToSection('comparison')} className="hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Fakta</button>
          <button onClick={() => scrollToSection('roadmap')} className="hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Vision</button>
          <button onClick={() => scrollToSection('community')} className="hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Community</button>
          <button onClick={() => scrollToSection('faq')} className="hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Frågor</button>
        </div>

        {/* Mobile Menu Button */}
        <button
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          className="md:hidden"
          style={{ color: 'var(--color-text-primary)' }}
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={isMenuOpen ? "M6 18L18 6M6 6l12 12" : "M4 6h16M4 12h16M4 18h16"} />
          </svg>
        </button>

        <motion.button
          onClick={() => window.open('https://github.com/magnusfroste/garageai/discussions', '_blank')}
          className="apple-button-primary px-4 py-2 text-sm"
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
        >
          Gå med
        </motion.button>
      </div>

      {/* Mobile Menu */}
      {isMenuOpen && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          style={{
            background: 'rgba(10, 10, 10, 0.95)',
            backdropFilter: 'blur(20px)',
            borderTop: '1px solid var(--color-border)'
          }}
          className="md:hidden"
        >
          <div className="px-4 py-4 space-y-2">
            <button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="block w-full text-left py-2 hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Hem</button>
            <button onClick={() => scrollToSection('comparison')} className="block w-full text-left py-2 hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Fakta</button>
            <button onClick={() => scrollToSection('roadmap')} className="block w-full text-left py-2 hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Vision</button>
            <button onClick={() => scrollToSection('community')} className="block w-full text-left py-2 hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Community</button>
            <button onClick={() => scrollToSection('faq')} className="block w-full text-left py-2 hover:text-cyan-400 transition" style={{ color: 'var(--color-text-secondary)' }}>Frågor</button>
          </div>
        </motion.div>
      )}
    </motion.nav>
  );
};

export default Header;
