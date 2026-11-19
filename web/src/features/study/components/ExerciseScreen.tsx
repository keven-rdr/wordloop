import * as stylex from '@stylexjs/stylex';
import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { StudyCard } from '../../../api/generated/types.gen';
import { DIFFICULTIES, ERROR_KINDS } from '../domain/answer';
import { asksDifficulty, type Checked } from '../hooks/sessionReducer';
import { useAnswerState } from '../hooks/useAnswerState';
import { type Grade, useExerciseKeys } from '../hooks/useExerciseKeys';
import { AssembleExercise } from './AssembleExercise';
import { ChooseExercise } from './ChooseExercise';
import { DifficultyPicker } from './DifficultyPicker';
import { Feedback } from './Feedback';
import { p } from './panelStyles';
import { s } from './styles';
import { TypeExercise } from './TypeExercise';

export type Answer = Readonly<{ answer?: string; chosenItemId?: string }>;

type Props = Readonly<{
  card: StudyCard;
  checked?: Checked;
  hintUsed: boolean;
  onHint: () => void;
  onCheck: (answer: Answer) => void;
  onGrade: (grade: Grade) => void;
}>;

type BarProps = Readonly<{
  checked?: Checked;
  canCheck: boolean;
  isType: boolean;
  verify: () => void;
  skip: () => void;
  next: () => void;
}>;

function ActionBar({ checked, canCheck, isType, verify, skip, next }: BarProps) {
  const { t } = useTranslation('study');
  const nextButton = useRef<HTMLButtonElement>(null);
  const showNext = checked !== undefined && !asksDifficulty(checked);
  useEffect(() => {
    if (showNext) nextButton.current?.focus();
  }, [showNext]);
  if (checked && asksDifficulty(checked)) return null;
  return (
    <footer {...stylex.props(s.bar)}>
      <div {...stylex.props(s.column, s.top)}>
        {checked ? (
          <button ref={nextButton} type="button" onClick={next} {...stylex.props(s.action)}>
            {t('next')}
          </button>
        ) : (
          <>
            {isType && (
              <button type="button" onClick={skip} {...stylex.props(s.action, s.secondary)}>
                {t('type.unknown')}
              </button>
            )}
            <button type="button" disabled={!canCheck} onClick={verify} {...stylex.props(s.action)}>
              {t('verify')}
            </button>
          </>
        )}
      </div>
    </footer>
  );
}

// O resultado da resposta: o feedback e, conforme o caso, a pergunta de dificuldade (acertou) ou do tipo de erro (errou).
function Verdict({
  card,
  checked,
  onGrade,
}: Readonly<{ card: StudyCard; checked?: Checked; onGrade: (grade: Grade) => void }>) {
  const ask = checked ? asksDifficulty(checked) : undefined;
  return (
    <div aria-live="polite" {...stylex.props(p.live)}>
      {checked && <Feedback card={card} outcome={checked.outcome} />}
      {ask === 'difficulty' && checked && (
        <DifficultyPicker
          kind="difficulty"
          values={DIFFICULTIES}
          suggested={checked.suggested}
          onPick={(difficulty) => onGrade({ difficulty })}
        />
      )}
      {ask === 'errorKind' && (
        <DifficultyPicker kind="errorKind" values={ERROR_KINDS} onPick={(errorKind) => onGrade({ errorKind })} />
      )}
    </div>
  );
}

// Uma pergunta: responder → verificar → feedback → dificuldade (ou continuar). A mesma moldura serve a todo exercício.
export function ExerciseScreen({ card, checked, hintUsed, onHint, onCheck, onGrade }: Props) {
  const a = useAnswerState(card);
  const verify = () => onCheck(a.answer);
  useExerciseKeys(
    { card, checked, chosen: a.chosen, assemble: a.assemble },
    { onChoose: a.setChosen, onVerify: verify, onGrade, onPlace: a.placeAt, onUndo: a.undo },
  );
  return (
    <>
      <main {...stylex.props(s.body)}>
        <div {...stylex.props(s.column)}>
          {a.kind === 'type' && (
            <TypeExercise
              card={card}
              typed={a.typed}
              locked={checked !== undefined}
              hintUsed={hintUsed}
              onType={a.setTyped}
              onSubmit={verify}
              onHint={onHint}
            />
          )}
          {a.kind === 'assemble' && (
            <AssembleExercise
              card={card}
              placed={a.placed}
              outcome={checked?.outcome}
              onPlace={a.placeTile}
              onRemove={a.removeAt}
              onClear={a.clear}
            />
          )}
          {a.kind === 'choose' && (
            <ChooseExercise card={card} chosen={a.chosen} outcome={checked?.outcome} onChoose={a.setChosen} />
          )}
          <Verdict card={card} checked={checked} onGrade={onGrade} />
        </div>
      </main>
      <ActionBar
        checked={checked}
        canCheck={a.canCheck}
        isType={a.kind !== 'choose'}
        verify={verify}
        skip={() => onCheck({ answer: '' })}
        next={() => onGrade({})}
      />
    </>
  );
}
