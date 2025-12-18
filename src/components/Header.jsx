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
      className="fixed top-0 w-full bg-opacity-80 backdrop-blur-md bg-slate-950 z-50 border-b border-green-500/20"
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
          <button onClick={() => scrollToSection('problem')} className="hover:text-green-400 transition">Problem</button>
          <button onClick={() => scrollToSection('solution')} className="hover:text-green-400 transition">Lösning</button>
          <button onClick={() => scrollToSection('infrastructure')} className="hover:text-green-400 transition">Infrastruktur</button>
          <button onClick={() => scrollToSection('ai-explained')} className="hover:text-green-400 transition">AI-Inferens</button>
          <button onClick={() => scrollToSection('community')} className="hover:text-green-400 transition">Community</button>
          <button onClick={() => scrollToSection('sponsors')} className="hover:text-green-400 transition">Sponsors</button>
        </div>

        {/* Mobile Menu Button */}
        <button
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          className="md:hidden text-white"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={isMenuOpen ? "M6 18L18 6M6 6l12 12" : "M4 6h16M4 12h16M4 18h16"} />
          </svg>
        </button>

        <motion.button
          onClick={showComingSoon}
          className="px-6 py-2 bg-green-500/20 border border-green-500 rounded-lg hover:bg-green-500/30 transition text-sm coming-soon-tooltip"
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
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
          className="md:hidden bg-slate-950/95 backdrop-blur-md border-t border-green-500/20"
        >
          <div className="px-4 py-4 space-y-2">
            <button onClick={() => scrollToSection('problem')} className="block w-full text-left py-2 hover:text-green-400 transition">Problem</button>
            <button onClick={() => scrollToSection('solution')} className="block w-full text-left py-2 hover:text-green-400 transition">Lösning</button>
            <button onClick={() => scrollToSection('infrastructure')} className="block w-full text-left py-2 hover:text-green-400 transition">Infrastruktur</button>
            <button onClick={() => scrollToSection('ai-explained')} className="block w-full text-left py-2 hover:text-green-400 transition">AI-Inferens</button>
            <button onClick={() => scrollToSection('community')} className="block w-full text-left py-2 hover:text-green-400 transition">Community</button>
            <button onClick={() => scrollToSection('sponsors')} className="block w-full text-left py-2 hover:text-green-400 transition">Sponsors</button>
          </div>
        </motion.div>
      )}
    </motion.nav>
  );
};

export default Header;
