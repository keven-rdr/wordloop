import { type createMemoryHistory, createRouter } from '@tanstack/react-router';
import { routeTree } from '../routeTree.gen';

export function createAppRouter(history?: ReturnType<typeof createMemoryHistory>) {
  return createRouter({ routeTree, history, defaultPreload: 'intent' });
}

export type AppRouter = ReturnType<typeof createAppRouter>;

declare module '@tanstack/react-router' {
  interface Register {
    router: AppRouter;
  }
}
