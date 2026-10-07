import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { configureApi } from './core/api';
import { setupI18n } from './core/i18n';
import './global.css';

if (import.meta.env.DEV) {
  // Em dev o CSS do StyleX vem de um modulo virtual; em producao ele e anexado ao CSS emitido pelo Vite.
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = '/virtual:stylex.css';
  document.head.append(link);
  void import('virtual:stylex:runtime');
}

const queryClient = new QueryClient();

async function bootstrap() {
  configureApi();
  await setupI18n();
  const root = document.getElementById('root');
  if (!root) return;
  createRoot(root).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </StrictMode>,
  );
}

void bootstrap();
