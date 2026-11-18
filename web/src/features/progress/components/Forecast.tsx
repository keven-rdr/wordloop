import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { ForecastDay, MissedItem } from '../../../api/generated/types.gen';
import { barHeight, totalDue } from '../domain/grid';
import { f, m, p } from './styles';

const weekday = (language: string, date: string) =>
  new Intl.DateTimeFormat(language, { weekday: 'narrow', timeZone: 'UTC' }).format(new Date(`${date}T00:00:00Z`));

// Quantas revisões vencem em cada um dos próximos 7 dias (atrasadas contam em hoje).
export function Forecast({ forecast }: Readonly<{ forecast: readonly ForecastDay[] }>) {
  const { t, i18n } = useTranslation('progress');
  const total = totalDue(forecast);
  return (
    <section aria-labelledby="progress-forecast" {...stylex.props(p.card)}>
      <h2 id="progress-forecast" {...stylex.props(p.title)}>
        {t('forecast.title')}
      </h2>
      <p {...stylex.props(p.text)}>{total === 0 ? t('forecast.none') : t('forecast.total', { count: total })}</p>
      <div aria-hidden="true" {...stylex.props(f.bars)}>
        {forecast.map((day) => (
          <div key={day.date} {...stylex.props(f.col)}>
            <div {...stylex.props(f.bar(barHeight(day.due, forecast)))} />
          </div>
        ))}
      </div>
      <div aria-hidden="true" {...stylex.props(f.labels)}>
        {forecast.map((day, at) => (
          <span key={day.date} {...stylex.props(f.label)}>
            {at === 0 ? t('forecast.today') : weekday(i18n.language, day.date)}
          </span>
        ))}
      </div>
    </section>
  );
}

// As palavras em que você mais errou no último mês.
export function MostMissed({ items }: Readonly<{ items: readonly MissedItem[] }>) {
  const { t } = useTranslation('progress');
  return (
    <section aria-labelledby="progress-missed" {...stylex.props(p.card)}>
      <h2 id="progress-missed" {...stylex.props(p.title)}>
        {t('missed.title')}
      </h2>
      {items.length === 0 ? (
        <p {...stylex.props(p.text)}>{t('missed.empty')}</p>
      ) : (
        <ul {...stylex.props(m.list)}>
          {items.map((item) => (
            <li key={item.item_id} {...stylex.props(m.item)}>
              <span>
                <span {...stylex.props(m.term)}>{item.term}</span> <span {...stylex.props(m.gloss)}>{item.gloss}</span>
              </span>
              <span {...stylex.props(m.misses)}>{t('missed.misses', { count: item.misses })}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
