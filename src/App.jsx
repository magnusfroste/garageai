import { useEffect } from 'react';
import { motion } from 'framer-motion';
import Header from './components/Header';
import Hero from './components/Hero';
import FactBoxes from './components/FactBoxes';
import Roadmap from './components/Roadmap';
import FAQ from './components/FAQ';
import Community from './components/Community';
import Sponsors from './components/Sponsors';
import CTA from './components/CTA';
import Footer from './components/Footer';

function App() {
  useEffect(() => {
    // Create floating particles
    const createParticles = (container) => {
      for (let i = 0; i < 80; i++) {
        const particle = document.createElement('div');
        particle.className = 'particle-bg';
        particle.style.left = Math.random() * 100 + '%';
        particle.style.top = -10 + 'px';
        particle.style.animationDelay = Math.random() * 10 + 's';
        particle.style.animationDuration = (15 + Math.random() * 20) + 's';
        container.appendChild(particle);
      }
    };

    const heroElement = document.querySelector('.hero-gradient');
    if (heroElement) {
      createParticles(heroElement);
    }
  }, []);

  return (
    <div className="min-h-screen">
      <Header />
      <Hero />
      <FactBoxes />
      <Roadmap />
      <FAQ />
      <Community />
      <Sponsors />
      <CTA />
      <Footer />
    </div>
  );
}

export default App;
