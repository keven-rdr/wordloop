import type { SVGProps } from 'react';

// Troféu do item "Ranking" no menu (mesmo traço dos ícones de shared/ui/nav).
export function TrophyIcon(props: Readonly<SVGProps<SVGSVGElement>>) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      {...props}
    >
      <path d="M8 21h8" />
      <path d="M12 17v4" />
      <path d="M7 4h10v5a5 5 0 0 1-10 0V4z" />
      <path d="M17 6h3v2a3 3 0 0 1-3 3" />
      <path d="M7 6H4v2a3 3 0 0 0 3 3" />
    </svg>
  );
}
