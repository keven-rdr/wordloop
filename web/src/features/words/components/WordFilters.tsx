import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { useUnits } from '../hooks/useUnits';
import { type Cefr, LEVELS, type SearchFilters } from '../hooks/useWordSearch';
import { w } from './styles';

type Props = Readonly<{ filters: SearchFilters; onChange: (next: SearchFilters) => void }>;

// Filtros da busca: unidade e nível (CEFR). Vazio = todos.
export function WordFilters({ filters, onChange }: Props) {
  const { t } = useTranslation('words');
  const { units } = useUnits();
  const pickLevel = (value: string) => onChange({ ...filters, cefr: LEVELS.find((l: Cefr) => l === value) });
  const pickUnit = (value: string) => onChange({ ...filters, unit: value || undefined });
  return (
    <div {...stylex.props(w.filters)}>
      <select
        aria-label={t('filter.unit')}
        value={filters.unit ?? ''}
        onChange={(e) => pickUnit(e.target.value)}
        {...stylex.props(w.select)}
      >
        <option value="" {...stylex.props(w.option)}>
          {t('filter.allUnits')}
        </option>
        {units.map((unit) => (
          <option key={unit.id} value={unit.id} {...stylex.props(w.option)}>
            {t(`units.theme.${unit.theme_key}`, { defaultValue: unit.theme_key })}
          </option>
        ))}
      </select>
      <select
        aria-label={t('filter.level')}
        value={filters.cefr ?? ''}
        onChange={(e) => pickLevel(e.target.value)}
        {...stylex.props(w.select)}
      >
        <option value="" {...stylex.props(w.option)}>
          {t('filter.allLevels')}
        </option>
        {LEVELS.map((level) => (
          <option key={level} value={level} {...stylex.props(w.option)}>
            {level}
          </option>
        ))}
      </select>
    </div>
  );
}
