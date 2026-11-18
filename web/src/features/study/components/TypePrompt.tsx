import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { StudyCard } from '../../../api/generated/types.gen';
import { s } from './styles';

// O enunciado das três variantes de digitar. O termo em inglês nunca aparece antes da resposta.
export function TypePrompt({ card }: Readonly<{ card: StudyCard }>) {
  const { t } = useTranslation('study');
  if (card.exercise === 'cloze' && card.cloze_text) {
    return (
      <>
        <p {...stylex.props(s.hint)}>{t('cloze.prompt')}</p>
        <h1 lang="en" {...stylex.props(s.term)}>
          {card.cloze_text}
        </h1>
        {card.example?.translation && <p {...stylex.props(s.example)}>{card.example.translation}</p>}
      </>
    );
  }
  if (card.exercise === 'image' && card.image_url) {
    return (
      <>
        <p {...stylex.props(s.hint)}>{t('image.prompt')}</p>
        <img src={card.image_url} alt={t('image.alt')} {...stylex.props(s.picture, s.pictureBig)} />
      </>
    );
  }
  return (
    <>
      <p {...stylex.props(s.hint)}>{t('type.prompt')}</p>
      <h1 {...stylex.props(s.term)}>{card.gloss}</h1>
    </>
  );
}
