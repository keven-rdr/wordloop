import * as stylex from '@stylexjs/stylex';
import { color, radius, size, space } from './tokens/color.stylex';

// Foco visível do teclado (o mesmo em botões, links de ação e campos).
const focus = {
  outlineStyle: { default: 'none', ':focus-visible': 'solid' },
  outlineWidth: 3,
  outlineColor: color.link,
  outlineOffset: 2,
} as const;

// Peças comuns das telas: cartão, textos, botão de ação, campo e barra de progresso nativa. Cada feature usa estas e
// guarda nos próprios estilos só o que é dela.
export const panel = stylex.create({
  card: {
    display: 'flex',
    flexDirection: 'column',
    gap: space.sm,
    padding: space.md,
    borderRadius: radius.md,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.surface,
  },
  section: { display: 'flex', flexDirection: 'column', gap: space.md },
  title: { margin: 0, fontSize: '0.95rem', fontWeight: 700, color: color.textMuted },
  heading: { margin: 0, fontSize: '1.2rem' },
  big: { margin: 0, fontSize: '1.4rem', fontWeight: 800, lineHeight: 1.2 },
  text: { margin: 0, color: color.textMuted, fontSize: '0.95rem' },
  good: { margin: 0, color: color.success, fontWeight: 700 },
  error: { margin: 0, color: color.danger, fontWeight: 600 },
  // Mensagem de status (<output>): ocupa a linha inteira, como um parágrafo.
  status: { display: 'block' },
  // <progress> nativo: o navegador desenha a barra e o leitor de tela lê o valor.
  progress: { width: '100%', height: space.sm, accentColor: color.primary },
  action: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: size.touch,
    paddingInline: space.lg,
    borderRadius: radius.md,
    borderWidth: 0,
    backgroundColor: { default: color.primary, ':hover': color.primaryEdge, ':disabled': color.border },
    color: { default: color.onPrimary, ':disabled': color.textMuted },
    fontSize: '1rem',
    fontWeight: 700,
    textDecoration: 'none',
    cursor: { default: 'pointer', ':disabled': 'not-allowed' },
    ...focus,
  },
  secondary: {
    backgroundColor: { default: color.surface, ':hover': color.border },
    color: color.text,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
  },
  field: { display: 'flex', flexDirection: 'column', gap: space.xs },
  label: { fontSize: '0.9rem', fontWeight: 600 },
  input: {
    minHeight: size.touch,
    paddingInline: space.md,
    borderRadius: radius.md,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.bg,
    color: color.text,
    fontSize: '1rem',
    boxSizing: 'border-box',
    width: '100%',
    ...focus,
  },
});
