import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import './global.css';

if (import.meta.env.DEV) {
  // Em dev o CSS do StyleX vem de um módulo virtual; em produção ele é anexado ao CSS emitido pelo Vite.
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = '/virtual:stylex.css';
  document.head.append(link);
  void import('virtual:stylex:runtime');
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
