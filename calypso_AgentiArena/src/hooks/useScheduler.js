import { useState, useEffect, useCallback, useRef } from 'react';

const STORAGE_KEY = 'agentarena_schedules';

function loadSchedules() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
  } catch {
    return [];
  }
}

function saveSchedules(schedules) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(schedules));
}

function getIntervalMs(interval) {
  switch (interval) {
    case 'every5min':  return 5 * 60 * 1000;
    case 'hourly':     return 60 * 60 * 1000;
    case 'every6h':    return 6 * 60 * 60 * 1000;
    case 'daily':      return 24 * 60 * 60 * 1000;
    case 'weekly':     return 7 * 24 * 60 * 60 * 1000;
    default:           return 60 * 60 * 1000;
  }
}

export function useScheduler({ onFireSchedule } = {}) {
  const [schedules, setSchedules] = useState(loadSchedules);
  const onFireRef = useRef(onFireSchedule);
  onFireRef.current = onFireSchedule;

  // Persist whenever schedules change
  useEffect(() => {
    saveSchedules(schedules);
  }, [schedules]);

  // Poll every 30s to check if any schedule should fire
  useEffect(() => {
    const tick = () => {
      const now = Date.now();
      setSchedules(prev => prev.map(sched => {
        if (!sched.enabled) return sched;

        const intervalMs = getIntervalMs(sched.interval);
        const lastRun = sched.lastRunAt || 0;
        const isDue = now - lastRun >= intervalMs;

        if (isDue) {
          // Fire the external callback (to trigger payment + execution)
          if (onFireRef.current) {
            onFireRef.current(sched);
          }
          return { ...sched, lastRunAt: now, runCount: (sched.runCount || 0) + 1 };
        }
        return sched;
      }));
    };

    // Run once immediately on mount to catch any overdue tasks
    tick();

    const timer = setInterval(tick, 30_000);
    return () => clearInterval(timer);
  }, []);

  const addSchedule = useCallback((schedule) => {
    const newSchedule = {
      id: `sched_${Date.now()}`,
      createdAt: Date.now(),
      lastRunAt: null,
      runCount: 0,
      enabled: true,
      ...schedule,
    };
    setSchedules(prev => [newSchedule, ...prev]);
    return newSchedule;
  }, []);

  const toggleSchedule = useCallback((id) => {
    setSchedules(prev =>
      prev.map(s => s.id === id ? { ...s, enabled: !s.enabled } : s)
    );
  }, []);

  const removeSchedule = useCallback((id) => {
    setSchedules(prev => prev.filter(s => s.id !== id));
  }, []);

  const getNextRunTime = useCallback((sched) => {
    if (!sched.enabled || !sched.lastRunAt) return 'Pending first run';
    const next = sched.lastRunAt + getIntervalMs(sched.interval);
    const diff = next - Date.now();
    if (diff <= 0) return 'Due now';
    const h = Math.floor(diff / 3_600_000);
    const m = Math.floor((diff % 3_600_000) / 60_000);
    const s = Math.floor((diff % 60_000) / 1000);
    if (h > 0) return `in ${h}h ${m}m`;
    if (m > 0) return `in ${m}m ${s}s`;
    return `in ${s}s`;
  }, []);

  return { schedules, addSchedule, toggleSchedule, removeSchedule, getNextRunTime };
}

export const INTERVALS = [
  { value: 'every5min', label: 'Every 5 minutes (demo)' },
  { value: 'hourly',    label: 'Every hour' },
  { value: 'every6h',   label: 'Every 6 hours' },
  { value: 'daily',     label: 'Every day' },
  { value: 'weekly',    label: 'Every week' },
];
