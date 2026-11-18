import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { Dashboard } from '../../../api/generated/types.gen';
import { percent } from '../domain/grid';
import { p } from './styles';

// Sequência de dias e o que já foi feito hoje.
export function StreakCard({ dashboard }: Readonly<{ dashboard: Dashboard }>) {
  const { t } = useTranslation('progress');
  const { streak, studied_today: studiedToday, today_stats: today } = dashboard;
  let hint = t('streak.none');
  if (studiedToday) hint = t('streak.done');
  else if (streak > 0) hint = t('streak.keep');
  return (
    <section aria-labelledby="progress-streak" {...stylex.props(p.card)}>
      <h2 id="progress-streak" {...stylex.props(p.title)}>
        {t('streak.title')}
      </h2>
      <p {...stylex.props(p.big)}>{t('streak.days', { count: streak })}</p>
      <p {...stylex.props(p.text)}>{hint}</p>
      {studiedToday && (
        <p {...stylex.props(p.text)}>
          {t('today.answers', { count: today.reviews })} · {t('today.newWords', { count: today.new_items })}
        </p>
      )}
    </section>
  );
}

// Acertos dos últimos 7 dias e palavras conhecidas, lado a lado.
export function NumbersRow({ dashboard }: Readonly<{ dashboard: Dashboard }>) {
  const { t } = useTranslation('progress');
  const rate = dashboard.accuracy_7d;
  return (
    <div {...stylex.props(p.row)}>
      <section aria-labelledby="progress-accuracy" {...stylex.props(p.card, p.half)}>
        <h2 id="progress-accuracy" {...stylex.props(p.title)}>
          {t('accuracy.title')}
        </h2>
        <p {...stylex.props(p.big)}>{rate === undefined ? '–' : `${percent(rate)}%`}</p>
        {rate === undefined && <p {...stylex.props(p.text)}>{t('accuracy.empty')}</p>}
      </section>
      <section aria-labelledby="progress-known" {...stylex.props(p.card, p.half)}>
        <h2 id="progress-known" {...stylex.props(p.title)}>
          {t('known.title')}
        </h2>
        <p {...stylex.props(p.big)}>{dashboard.known_words}</p>
        <p {...stylex.props(p.text)}>{t('known.hint')}</p>
      </section>
    </div>
  );
}
