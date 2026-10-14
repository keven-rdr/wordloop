import { useSyncExternalStore } from 'react';

export type ThemeChoice = 'system' | 'light' | 'dark';

const storageKey = 'wl-theme';
const listeners = new Set<() => void>();

function stored(): ThemeChoice {
  try {
    const value = window.localStorage.getItem(storageKey);
    return value === 'light' || value === 'dark' ? value : 'system';
  } catch {
    return 'system';
  }
}

let choice: ThemeChoice = stored();

function persist(next: ThemeChoice) {
  try {
    if (next === 'system') window.localStorage.removeItem(storageKey);
    else window.localStorage.setItem(storageKey, next);
  } catch {
    return;
  }
}

export function setThemeChoice(next: ThemeChoice) {
  choice = next;
  persist(next);
  for (const notify of listeners) notify();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function useThemeChoice(): ThemeChoice {
  return useSyncExternalStore(
    subscribe,
    () => choice,
    () => 'system',
  );
}

export function systemPrefersDark(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-color-scheme: dark)').matches;
}

export function effectiveTheme(current: ThemeChoice): 'light' | 'dark' {
  if (current !== 'system') return current;
  return systemPrefersDark() ? 'dark' : 'light';
}
