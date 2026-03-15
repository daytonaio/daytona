import React, { useState } from 'react';
import { categoryColors } from '../data/agents';

/**
 * AgentAvatar — renders a unique robot avatar image from DiceBear API.
 * size: 'sm' | 'md' | 'lg'
 */
export default function AgentAvatar({ agent, size = 'md' }) {
  const color = categoryColors[agent.category] || '#00D4FF';
  const [imgError, setImgError] = useState(false);

  // Use DiceBear bottts-neutral style — unique per agent name seed
  const avatarUrl = agent.avatarUrl || 
    `https://api.dicebear.com/9.x/bottts-neutral/svg?seed=${encodeURIComponent(agent.name)}&backgroundColor=transparent&radius=12`;

  const dims = {
    sm: 'w-12 h-12',
    md: 'w-16 h-16',
    lg: 'w-24 h-24',
  };

  const outer = dims[size] || dims.md;

  return (
    <div
      className={`${outer} rounded-2xl flex items-center justify-center shrink-0 overflow-hidden`}
      style={{
        background: `radial-gradient(circle at 30% 30%, ${color}25, ${color}06)`,
        border: `1.5px solid ${color}45`,
        boxShadow: `0 0 16px ${color}20, inset 0 0 12px ${color}08`,
        padding: '6px',
      }}
    >
      {!imgError ? (
        <img
          src={avatarUrl}
          alt={agent.name}
          className="w-full h-full object-contain"
          onError={() => setImgError(true)}
          loading="lazy"
        />
      ) : (
        // Fallback: robot emoji if image fails to load
        <span className="text-3xl select-none" role="img" aria-label={agent.name}>🤖</span>
      )}
    </div>
  );
}
