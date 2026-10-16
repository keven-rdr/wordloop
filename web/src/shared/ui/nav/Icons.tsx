import type { ReactNode, SVGProps } from 'react';

type IconProps = Readonly<SVGProps<SVGSVGElement>>;

function Icon({ children, ...rest }: IconProps & Readonly<{ children: ReactNode }>) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      {...rest}
    >
      {children}
    </svg>
  );
}

export function HomeIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M3 11l9-8 9 8" />
      <path d="M5 10v10h5v-6h4v6h5V10" />
    </Icon>
  );
}

export function CardsIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="4" y="6" width="12" height="15" rx="2" />
      <path d="M8 3h10a2 2 0 0 1 2 2v12" />
    </Icon>
  );
}
