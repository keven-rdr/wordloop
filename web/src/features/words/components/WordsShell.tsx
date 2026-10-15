import * as stylex from '@stylexjs/stylex';
import type { ReactNode } from 'react';
import { w } from './styles';

type Props = Readonly<{
  title: string;
  subtitle?: string;
  top?: ReactNode;
  extra?: ReactNode;
  plain?: boolean;
  children: ReactNode;
}>;

export function WordsShell({ title, subtitle, top, extra, plain, children }: Props) {
  return (
    <div {...stylex.props(w.screen)}>
      <div {...stylex.props(w.inner)}>
        <header {...stylex.props(w.header)}>
          {top}
          <h1 {...stylex.props(w.title)}>{title}</h1>
          {subtitle && <p {...stylex.props(w.subtitle)}>{subtitle}</p>}
          {extra}
        </header>
      </div>
      {plain ? (
        <div {...stylex.props(w.inner, w.deckBody)}>{children}</div>
      ) : (
        <div {...stylex.props(w.sheet)}>
          <div {...stylex.props(w.inner)}>{children}</div>
        </div>
      )}
    </div>
  );
}
