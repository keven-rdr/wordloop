import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { HeatmapDay } from '../../../api/generated/types.gen';
import { buildGrid, type Level, studiedDays, WEEKS } from '../domain/grid';
import { h, p } from './styles';

// Opacidade do roxo em cada tom (o tom 0 é só a cor de fundo da célula).
const OPACITY: Record<Level, number> = { 0: 1, 1: 0.35, 2: 0.6, 3: 0.8, 4: 1 };

// Quanto você estudou nas últimas 12 semanas. É uma imagem com resumo falado: 84 células soltas só atrapalhariam
// quem usa leitor de tela.
export function Heatmap({ today, days }: Readonly<{ today: string; days: readonly HeatmapDay[] }>) {
  const { t } = useTranslation('progress');
  const cells = buildGrid(today, days);
  return (
    <section aria-labelledby="progress-heatmap" {...stylex.props(p.card)}>
      <h2 id="progress-heatmap" {...stylex.props(p.title)}>
        {t('heatmap.title', { count: WEEKS })}
      </h2>
      <div
        role="img"
        aria-label={t('heatmap.summary', { count: studiedDays(days), weeks: WEEKS })}
        {...stylex.props(h.grid)}
      >
        {cells.map((cell) => (
          <span
            key={cell.key}
            data-level={cell.date ? cell.level : undefined}
            {...stylex.props(h.cell, !cell.date && h.hidden, cell.level > 0 && h.level(OPACITY[cell.level]))}
          />
        ))}
      </div>
    </section>
  );
}
