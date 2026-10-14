import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { art } from './brandStyles';

export function BrandPanel() {
  const { t, ready } = useTranslation('auth');
  const { t: tCommon } = useTranslation('common');
  return (
    <aside {...stylex.props(art.panel)}>
      <div aria-hidden="true" {...stylex.props(art.orb, art.orbPink)} />
      <div aria-hidden="true" {...stylex.props(art.orb, art.orbYellow)} />
      <div aria-hidden="true" {...stylex.props(art.orb, art.orbViolet)} />
      <div {...stylex.props(art.logo)}>
        <img src="/favicon.svg" alt="" {...stylex.props(art.mark)} />
        <span {...stylex.props(art.name)}>{tCommon('app.name')}</span>
      </div>
      <div aria-hidden="true" {...stylex.props(art.stage)}>
        <div {...stylex.props(art.float)}>
          <div {...stylex.props(art.card, art.cardYellow)} />
        </div>
        <div {...stylex.props(art.float, art.floatB)}>
          <div {...stylex.props(art.card, art.cardPink)} />
        </div>
        <div {...stylex.props(art.float, art.floatC)}>
          <div {...stylex.props(art.card, art.cardPaper)}>
            <svg
              aria-hidden="true"
              viewBox="0 0 100 100"
              fill="none"
              stroke="currentColor"
              strokeWidth="7"
              strokeLinecap="round"
              strokeLinejoin="round"
              {...stylex.props(art.glyph)}
            >
              <path d="M24 34 L36 72 L50 44 L64 72 L76 34" />
            </svg>
          </div>
        </div>
      </div>
      <div {...stylex.props(art.copy)}>
        {ready && (
          <>
            <h2 {...stylex.props(art.headline)}>
              {t('brand.headline')} <span {...stylex.props(art.accent)}>{t('brand.headlineAccent')}</span>
            </h2>
            <p {...stylex.props(art.text)}>{t('brand.text')}</p>
          </>
        )}
      </div>
    </aside>
  );
}
