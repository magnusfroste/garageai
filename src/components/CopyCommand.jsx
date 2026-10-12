import { useState } from 'react';

// A shell command with a copy button; long lines scroll inside the box, not the page.
const CopyCommand = ({ command, note, color }) => {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  };
  return (
    <div className="mt-3">
      <div
        className="flex items-stretch rounded-lg overflow-hidden"
        style={{ background: 'rgba(0,0,0,0.35)', border: '1px solid rgba(255,255,255,0.12)' }}
      >
        <pre className="flex-1 min-w-0 overflow-x-auto px-3 py-2 text-xs m-0" style={{ whiteSpace: 'pre', overflowWrap: 'normal' }}>
          <code style={{ color: 'var(--color-text-primary)' }}>{command}</code>
        </pre>
        <button
          type="button"
          onClick={copy}
          className="shrink-0 px-3 text-xs font-bold"
          style={{ background: color, color: '#000' }}
          aria-label="Copy command"
        >
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
      {note && <p className="mt-1 text-xs" style={{ color: 'var(--color-text-muted)' }}>{note}</p>}
    </div>
  );
};

export default CopyCommand;
