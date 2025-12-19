import { useState } from 'react';
import { motion } from 'framer-motion';

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const showComingSoon = () => {
    alert('🚀 Community-portalen lanseras snart!\n\nI mellan tiden:\n\n📞 Autoversio → autoversio.com\n⚡ Liteit → liteit.se');
  };

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
        background: 'rgba(255, 255, 255, 0.8)',
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
          <button onClick={() => scrollToSection('comparison')} className="hover:text-primary transition">Fakta</button>
          <button onClick={() => scrollToSection('roadmap')} className="hover:text-primary transition">Vision</button>
          <button onClick={() => scrollToSection('faq')} className="hover:text-primary transition">Frågor</button>
          <button onClick={() => scrollToSection('sponsors')} className="hover:text-primary transition">Partners</button>
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
          onClick={showComingSoon}
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
            background: 'rgba(255, 255, 255, 0.95)',
            backdropFilter: 'blur(20px)',
            borderTop: '1px solid var(--color-border)'
          }}
          className="md:hidden"
        >
          <div className="px-4 py-4 space-y-2">
            <button onClick={() => scrollToSection('comparison')} className="block w-full text-left py-2 hover:text-primary transition">Fakta</button>
            <button onClick={() => scrollToSection('roadmap')} className="block w-full text-left py-2 hover:text-primary transition">Vision</button>
            <button onClick={() => scrollToSection('faq')} className="block w-full text-left py-2 hover:text-primary transition">Frågor</button>
            <button onClick={() => scrollToSection('sponsors')} className="block w-full text-left py-2 hover:text-primary transition">Partners</button>
          </div>
        </motion.div>
      )}
    </motion.nav>
  );
};

export default Header;
