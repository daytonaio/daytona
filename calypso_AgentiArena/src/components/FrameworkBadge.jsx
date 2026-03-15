import React from 'react';
import { frameworkColors } from '../data/agents';

export default function FrameworkBadge({ framework }) {
  const color = frameworkColors[framework] || '#8888AA';

  return (
    <span
      className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold"
      style={{
        backgroundColor: `${color}20`,
        color: color,
        border: `1px solid ${color}40`,
      }}
    >
      {framework}
    </span>
  );
}
