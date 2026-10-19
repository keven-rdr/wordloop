import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { ItemSummary } from '../../../api/generated/types.gen';
import { type SearchFilters, useWordSearch } from '../hooks/useWordSearch';
import { w } from './styles';

function Row({ item }: Readonly<{ item: ItemSummary }>) {
  const { t } = useTranslation('words');
  return (
    <li {...stylex.props(w.row)}>
      <span {...stylex.props(w.rowText)}>
        <span {...stylex.props(w.rowTitle)}>{item.term}</span>
        <span {...stylex.props(w.muted)}>{item.gloss}</span>
      </span>
      <span {...stylex.props(w.rowText)}>
        {item.cefr && <span {...stylex.props(w.chip)}>{item.cefr}</span>}
        {item.status === 'draft' && <span {...stylex.props(w.chip)}>{t('draft')}</span>}
      </span>
    </li>
  );
}

export function SearchResults({ text, filters }: Readonly<{ text: string; filters: SearchFilters }>) {
  const { t } = useTranslation('words');
  const search = useWordSearch(text, filters);
  if (!search.active) return <p {...stylex.props(w.muted)}>{t('search.short')}</p>;
  if (search.failed)
    return (
      <p role="alert" {...stylex.props(w.error)}>
        {t('error.load')}
      </p>
    );
  if (search.items.length === 0 && !search.loading) return <p {...stylex.props(w.muted)}>{t('search.empty')}</p>;
  return (
    <>
      <p aria-live="polite" {...stylex.props(w.muted)}>
        {t('search.results', { count: search.total })}
      </p>
      <ul {...stylex.props(w.list)}>
        {search.items.map((item) => (
          <Row key={item.id} item={item} />
        ))}
      </ul>
    </>
  );
}
