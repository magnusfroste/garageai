import { useScroll, useSpring, motion } from 'framer-motion';
import Header from './components/Header';
import Hero from './components/Hero';
import FactBoxes from './components/FactBoxes';
import AIExplained from './components/AIExplained';
import Infrastructure from './components/Infrastructure';
import B2B from './components/B2B';
import TokenEconomy from './components/TokenEconomy';
import Roadmap from './components/Roadmap';
import FAQ from './components/FAQ';
import Community from './components/Community';
import Sponsors from './components/Sponsors';
import CTA from './components/CTA';
import Footer from './components/Footer';

function ScrollProgress() {
  const { scrollYProgress } = useScroll();
  const scaleX = useSpring(scrollYProgress, { stiffness: 200, damping: 30 });
  return (
    <motion.div
      style={{
        scaleX,
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        height: '2px',
        background: 'linear-gradient(90deg, #00D4FF, #00FF88, #FF00AA)',
        transformOrigin: '0%',
        zIndex: 9999,
      }}
    />
  );
}

function App() {
  return (
    <div className="min-h-screen">
      <ScrollProgress />
      <Header />
      <Hero />
      <FactBoxes />
      <AIExplained />
      <Infrastructure />
      <B2B />
      <TokenEconomy />
      <Roadmap />
      <Community />
      <FAQ />
      <Sponsors />
      <CTA />
      <Footer />
    </div>
  );
}

export default App;
