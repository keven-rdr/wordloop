import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useWho } from '../hooks/useWho';
import { hasFilters, type SearchFilters } from '../hooks/useWordSearch';
import { SearchResults } from './SearchResults';
import { w } from './styles';
import { UnitList } from './UnitList';
import { WordFilters } from './WordFilters';
import { WordsShell } from './WordsShell';

export function WordsPage() {
  const { t, ready } = useTranslation('words');
  const who = useWho();
  const [text, setText] = useState('');
  const [filters, setFilters] = useState<SearchFilters>({});
  if (who === 'anonymous') return <Navigate to="/login" replace />;
  if (!ready || who === 'loading') return <div aria-busy="true" />;
  const search = (
    <>
      <input
        type="search"
        value={text}
        onChange={(e) => setText(e.target.value)}
        maxLength={100}
        aria-label={t('search.label')}
        placeholder={t('search.placeholder')}
        {...stylex.props(w.search)}
      />
      <WordFilters filters={filters} onChange={setFilters} />
    </>
  );
  const searching = text.trim() !== '' || hasFilters(filters);
  return (
    <WordsShell title={t('title')} subtitle={t('subtitle')} extra={search}>
      {searching ? <SearchResults text={text} filters={filters} /> : <UnitList />}
      <Link to="/words/credits" {...stylex.props(w.creditsLink)}>
        {t('credits.link')}
      </Link>
    </WordsShell>
  );
}
