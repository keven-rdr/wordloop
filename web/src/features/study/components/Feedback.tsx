import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { StudyCard } from '../../../api/generated/types.gen';
import type { Outcome } from '../domain/answer';
import { CheckIcon, CrossIcon } from './Icons';
import { p } from './panelStyles';

type Props = Readonly<{ card: StudyCard; outcome: Outcome }>;

// Cor nunca sozinha: ícone e texto dizem se acertou. Quem corrige de verdade é o servidor.
export function Feedback({ card, outcome }: Props) {
  const { t } = useTranslation('study');
  const right = outcome === 'right';
  return (
    <section {...stylex.props(p.feedback, right ? p.feedbackRight : p.feedbackWrong)}>
      <h2 {...stylex.props(p.feedbackHead, right ? p.right : p.wrong)}>
        {right ? <CheckIcon {...stylex.props(p.feedbackIcon)} /> : <CrossIcon {...stylex.props(p.feedbackIcon)} />}
        {t(`feedback.${outcome}`)}
      </h2>
      {!right && (
        <p {...stylex.props(p.answer)}>
          {t('feedback.answer')} <strong lang="en">{card.term}</strong>
        </p>
      )}
      {card.example && (
        <p lang="en" {...stylex.props(p.example)}>
          {card.example.text}
        </p>
      )}
    </section>
  );
}
