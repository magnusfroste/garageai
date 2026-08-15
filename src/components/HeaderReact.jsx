import { useState } from 'react';
import { motion } from 'framer-motion';

const Header = ({ logo = '🏠 GARAGE AI', ctaText = 'Join the Movement', ctaUrl = 'https://github.com/magnusfroste/garageai/discussions', navLinks = [] } = {}) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const scrollToSection = (sectionId) => {
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
      setIsMenuOpen(false);
    }
  };

  const defaultNavLinks = [
    { label: 'Home', sectionId: null },
    { label: 'EU Analysis', sectionId: 'eu-analysis' },
    { label: 'The Vision', sectionId: 'three-waves' },
    { label: 'The Garage Node', sectionId: 'infrastructure' },
    { label: 'For Business', sectionId: 'for-business' },
    { label: 'Economy', sectionId: 'token-economy' },
    { label: 'Roadmap', sectionId: 'roadmap' },
    { label: 'FAQ', sectionId: 'faq' },
  ];

  const links = navLinks.length > 0 ? navLinks : defaultNavLinks;

  const handleNavClick = (sectionId) => {
    if (sectionId === null) {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else {
      scrollToSection(sectionId);
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
          className="text-xl font-bold glow-text cursor-pointer"
          style={{ color: 'var(--color-primary)' }}
          whileHover={{ scale: 1.05 }}
          onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
        >
          {logo}
        </motion.div>

        {/* Desktop Menu */}
        <div className="hidden md:flex gap-6 text-sm">
          {links.map((link) => (
            <button
              key={link.label}
              onClick={() => handleNavClick(link.sectionId)}
              className="hover:text-cyan-400 transition"
              style={{ color: 'var(--color-text-secondary)' }}
            >
              {link.label}
            </button>
          ))}
        </div>

        {/* Mobile toggle */}
        <button
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          className="md:hidden"
          style={{ color: 'var(--color-text-primary)' }}
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d={isMenuOpen ? 'M6 18L18 6M6 6l12 12' : 'M4 6h16M4 12h16M4 18h16'} />
          </svg>
        </button>

        <motion.button
          onClick={() => window.open(ctaUrl, '_blank')}
          className="apple-button-primary px-4 py-2 text-sm hidden md:block"
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
        >
          {ctaText}
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
        >
          <div className="px-4 py-4 space-y-2">
            {links.map((link) => (
              <button
                key={link.label}
                onClick={() => handleNavClick(link.sectionId)}
                className="block w-full text-left py-2 hover:text-cyan-400 transition"
                style={{ color: 'var(--color-text-secondary)' }}
              >
                {link.label}
              </button>
            ))}
          </div>
        </motion.div>
      )}
    </motion.nav>
  );
};

export default Header;
