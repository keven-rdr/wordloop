import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { sectionsOf } from '../domain/legal';
import { l } from './styles';

type Props = Readonly<{ kind: 'privacy' | 'terms' }>;

const keys = {
  privacy: { title: 'privacy.title', updated: 'privacy.updated', sections: 'privacy.sections' },
  terms: { title: 'terms.title', updated: 'terms.updated', sections: 'terms.sections' },
} as const;

// Página pública de texto (política de privacidade ou termos de uso): abre sem entrar na conta.
export function LegalPage({ kind }: Props) {
  const { t, ready } = useTranslation('legal');
  if (!ready) return <div aria-busy="true" />;
  const key = keys[kind];
  return (
    <BrandScreen title={t(key.title)} subtitle={t('draft.badge')}>
      <article {...stylex.props(l.article)}>
        <p role="note" {...stylex.props(l.notice)}>
          {t('draft.notice')}
        </p>
        <p {...stylex.props(l.updated)}>{t(key.updated)}</p>
        {sectionsOf(t(key.sections, { returnObjects: true })).map((section) => (
          <section key={section.title} {...stylex.props(l.section)}>
            <h2 {...stylex.props(l.heading)}>{section.title}</h2>
            {section.body.map((line) => (
              <p key={line} {...stylex.props(l.paragraph)}>
                {line}
              </p>
            ))}
          </section>
        ))}
        <nav aria-label={t('links.label')} {...stylex.props(l.nav)}>
          <Link to={kind === 'privacy' ? '/terms' : '/privacy'} {...stylex.props(l.link)}>
            {t(kind === 'privacy' ? 'links.terms' : 'links.privacy')}
          </Link>
          <Link to="/" {...stylex.props(l.link)}>
            {t('links.home')}
          </Link>
        </nav>
      </article>
    </BrandScreen>
  );
}
