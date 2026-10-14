import { RouterProvider } from '@tanstack/react-router';
import type { AppRouter } from './core/router';

export function App({ router }: Readonly<{ router: AppRouter }>) {
  return <RouterProvider router={router} />;
}
