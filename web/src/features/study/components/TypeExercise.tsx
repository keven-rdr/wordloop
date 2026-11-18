import * as stylex from '@stylexjs/stylex';
import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { StudyCard } from '../../../api/generated/types.gen';
import { s } from './styles';

type Props = Readonly<{
  card: StudyCard;
  typed: string;
  locked: boolean;
  hintUsed: boolean;
  onType: (value: string) => void;
  onSubmit: () => void;
  onHint: () => void;
}>;

export function TypeExercise({ card, typed, locked, hintUsed, onType, onSubmit, onHint }: Props) {
  const { t } = useTranslation('study');
  const input = useRef<HTMLInputElement>(null);
  // O teclado abre já na resposta; a área rola para a barra de ação não cobrir o campo.
  useEffect(() => input.current?.focus(), []);
  return (
    <>
      <p {...stylex.props(s.hint)}>{t('type.prompt')}</p>
      <h1 {...stylex.props(s.term)}>{card.gloss}</h1>
      <form
        onSubmit={(event) => {
          event.preventDefault();
          if (!locked && typed.trim() !== '') onSubmit();
        }}
      >
        <label {...stylex.props(s.srOnly)} htmlFor="study-answer">
          {t('type.label')}
        </label>
        <input
          id="study-answer"
          ref={input}
          value={typed}
          readOnly={locked}
          lang="en"
          autoCapitalize="none"
          autoCorrect="off"
          autoComplete="off"
          spellCheck={false}
          enterKeyHint="done"
          onChange={(event) => onType(event.target.value)}
          {...stylex.props(s.input)}
        />
      </form>
      {!locked && !hintUsed && (
        <button type="button" onClick={onHint} {...stylex.props(s.link)}>
          {t('type.hint')}
        </button>
      )}
      {hintUsed && <p {...stylex.props(s.hint)}>{t('type.hintShown', { letter: card.term.charAt(0) })}</p>}
    </>
  );
}
