import React, { useEffect, useState } from 'react';
import { getReputationColor } from '../data/agents';

export default function ReputationGauge({ score, size = 100 }) {
  const [animatedScore, setAnimatedScore] = useState(0);
  const color = getReputationColor(score);
  const radius = (size - 10) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (animatedScore / 100) * circumference;

  useEffect(() => {
    const timer = setTimeout(() => setAnimatedScore(score), 100);
    return () => clearTimeout(timer);
  }, [score]);

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="#1E1E2E"
          strokeWidth="5"
          fill="none"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke={color}
          strokeWidth="5"
          fill="none"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          className="gauge-circle"
          style={{ filter: `drop-shadow(0 0 6px ${color}60)` }}
        />
      </svg>
      <div className="absolute flex flex-col items-center">
        <span className="text-lg font-bold" style={{ color, fontSize: size * 0.22 }}>{score}</span>
        {size >= 80 && <span className="text-muted" style={{ fontSize: size * 0.1 }}>REP</span>}
      </div>
    </div>
  );
}
