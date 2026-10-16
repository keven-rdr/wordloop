import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { brand } from '../tokens/brand.stylex';
import { radius } from '../tokens/color.stylex';

const styles = stylex.create({
  base: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
    borderRadius: radius.full,
    backgroundColor: brand.violet,
    color: brand.onBrand,
    fontWeight: 800,
    overflow: 'hidden',
    objectFit: 'cover',
  },
  small: { width: '2.25rem', height: '2.25rem', fontSize: '1.1rem' },
  large: { width: '5rem', height: '5rem', fontSize: '2.2rem' },
});

type Props = Readonly<{ name: string; pictureUrl?: string; large?: boolean }>;

export function initialOf(name: string): string {
  const first = Array.from(name.trim())[0];
  return first ? first.toUpperCase() : '?';
}

// Foto do login social; sem foto (ou se ela falhar ao carregar) mostra a inicial do nome.
export function Avatar({ name, pictureUrl, large }: Props) {
  const [failedUrl, setFailedUrl] = useState('');
  const look = stylex.props(styles.base, large ? styles.large : styles.small);
  if (pictureUrl && pictureUrl !== failedUrl) {
    return (
      <img src={pictureUrl} alt="" referrerPolicy="no-referrer" onError={() => setFailedUrl(pictureUrl)} {...look} />
    );
  }
  return (
    <span aria-hidden="true" {...look}>
      {initialOf(name)}
    </span>
  );
}
