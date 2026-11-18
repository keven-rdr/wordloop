import * as stylex from '@stylexjs/stylex';
import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { StudyCard } from '../../../api/generated/types.gen';
import type { Outcome } from '../domain/answer';
import { s } from './styles';

type Props = Readonly<{
  card: StudyCard;
  chosen?: string;
  outcome?: Outcome;
  onChoose: (itemId: string) => void;
}>;

export const OPTION_KEYS = ['A', 'B', 'C', 'D'] as const;

export function ChooseExercise({ card, chosen, outcome, onChoose }: Props) {
  const { t } = useTranslation('study');
  const prompt = useRef<HTMLParagraphElement>(null);
  useEffect(() => prompt.current?.focus(), []);
  const locked = outcome !== undefined;
  return (
    <>
      <p ref={prompt} tabIndex={-1} {...stylex.props(s.hint)}>
        {t('choose.prompt')}
      </p>
      {card.image_url && <img src={card.image_url} alt="" {...stylex.props(s.picture)} />}
      <h1 {...stylex.props(s.term)} lang="en">
        {card.term}
      </h1>
      {card.example && (
        <p lang="en" {...stylex.props(s.example)}>
          {card.example.text}
        </p>
      )}
      <fieldset {...stylex.props(s.options)}>
        <legend {...stylex.props(s.srOnly)}>{t('choose.options')}</legend>
        {(card.options ?? []).map((option, index) => {
          const isChosen = chosen === option.item_id;
          const isAnswer = option.item_id === card.item_id;
          return (
            <button
              key={option.item_id}
              type="button"
              aria-pressed={isChosen}
              disabled={locked}
              onClick={() => onChoose(option.item_id)}
              {...stylex.props(
                s.option,
                isChosen && s.optionChosen,
                locked && isAnswer && s.optionRight,
                locked && isChosen && !isAnswer && s.optionWrong,
              )}
            >
              <span aria-hidden="true" {...stylex.props(s.optionKey)}>
                {OPTION_KEYS[index]}
              </span>
              {option.gloss}
            </button>
          );
        })}
      </fieldset>
    </>
  );
}
