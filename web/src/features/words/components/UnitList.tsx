import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useUnits } from '../hooks/useUnits';
import { LoadFailed } from './LoadFailed';
import { w } from './styles';

export function UnitList() {
  const { t } = useTranslation('words');
  const { units, loading, failed, retry } = useUnits();
  if (failed) return <LoadFailed onRetry={() => void retry()} />;
  if (loading) return <div aria-busy="true" />;
  if (units.length === 0) return <p {...stylex.props(w.muted)}>{t('units.empty')}</p>;
  return (
    <ul {...stylex.props(w.list)}>
      {units.map((unit) => (
        <li key={unit.id}>
          <Link to="/words/$unit" params={{ unit: unit.id }} {...stylex.props(w.row)}>
            <span {...stylex.props(w.rowTitle)}>
              {t(`units.theme.${unit.theme_key}`, { defaultValue: unit.theme_key })}
            </span>
            <span {...stylex.props(w.count)}>{t('units.count', { count: unit.item_count })}</span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
