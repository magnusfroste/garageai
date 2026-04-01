import { useEffect, useRef, useState } from 'react';
import { useInView } from 'framer-motion';

/**
 * Counts up to `value` (numeric) when scrolled into view.
 * Pass `prefix`/`suffix` for symbols like "$", "B", "%", "×".
 * `duration` in ms.
 */
const AnimatedCounter = ({ value, prefix = '', suffix = '', duration = 1800, className = '', style = {} }) => {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const [display, setDisplay] = useState('0');

  useEffect(() => {
    if (!isInView) return;
    const numeric = parseFloat(String(value).replace(/[^0-9.]/g, ''));
    if (isNaN(numeric)) { setDisplay(value); return; }

    const startTime = performance.now();
    const tick = (now) => {
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / duration, 1);
      // ease-out cubic
      const eased = 1 - Math.pow(1 - progress, 3);
      const current = numeric * eased;
      // format nicely: integers stay integers, decimals stay decimals
      const formatted = Number.isInteger(numeric)
        ? Math.round(current).toLocaleString()
        : current.toFixed(1);
      setDisplay(formatted);
      if (progress < 1) requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  }, [isInView, value, duration]);

  return (
    <span ref={ref} className={className} style={style}>
      {prefix}{display}{suffix}
    </span>
  );
};

export default AnimatedCounter;
