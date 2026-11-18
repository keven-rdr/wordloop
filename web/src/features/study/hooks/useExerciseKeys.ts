import { useEffect } from 'react';
import type { StudyCard } from '../../../api/generated/types.gen';
import { DIFFICULTIES, type Difficulty, ERROR_KINDS, type ErrorKind } from '../domain/answer';
import { asksDifficulty, type Checked } from './sessionReducer';

const LETTERS = ['a', 'b', 'c', 'd'];

export type Grade = Readonly<{ difficulty?: Difficulty; errorKind?: ErrorKind }>;

export type KeyAction =
  | Readonly<{ do: 'choose'; itemId: string }>
  | Readonly<{ do: 'verify' }>
  | Readonly<{ do: 'grade'; grade: Grade }>;

type Context = Readonly<{ card: StudyCard; checked?: Checked; chosen?: string }>;

// Posição 0–3 de uma tecla de opção (A–D ou 1–4), ou -1.
const optionIndex = (key: string) => (LETTERS.includes(key) ? LETTERS.indexOf(key) : Number(key) - 1);

function beforeCheck(key: string, { card, chosen }: Context): KeyAction | undefined {
  const option = card.exercise === 'choose' ? card.options?.[optionIndex(key)] : undefined;
  if (option) return { do: 'choose', itemId: option.item_id };
  return key === 'enter' && chosen ? { do: 'verify' } : undefined;
}

function afterCheck(key: string, checked: Checked): KeyAction | undefined {
  const ask = asksDifficulty(checked);
  const slot = Number(key) - 1;
  if (!ask) return key === 'enter' ? { do: 'grade', grade: {} } : undefined;
  if (ask === 'difficulty') {
    if (key === 'enter') return { do: 'grade', grade: { difficulty: checked.suggested } };
    const difficulty = DIFFICULTIES[slot];
    return difficulty ? { do: 'grade', grade: { difficulty } } : undefined;
  }
  const errorKind = ERROR_KINDS[slot];
  return errorKind ? { do: 'grade', grade: { errorKind } } : undefined;
}

// Atalhos (desktop): A–D ou 1–4 escolhem, Enter verifica/continua, 1–3 respondem a dificuldade.
export function keyAction(key: string, context: Context): KeyAction | undefined {
  const lower = key.toLowerCase();
  return context.checked ? afterCheck(lower, context.checked) : beforeCheck(lower, context);
}

// Num campo de texto quem recebe as teclas é o campo. Num botão focado, Enter e Espaço são do botão (ativam-no);
// as demais teclas (A–D, 1–3) continuam valendo, porque o foco cai nos botões de dificuldade.
const ownsKeys = (event: KeyboardEvent) =>
  event.target instanceof HTMLInputElement ||
  event.target instanceof HTMLTextAreaElement ||
  (event.target instanceof HTMLButtonElement && (event.key === 'Enter' || event.key === ' '));

type Handlers = Readonly<{
  onChoose: (itemId: string) => void;
  onVerify: () => void;
  onGrade: (grade: Grade) => void;
}>;

function run(action: KeyAction | undefined, { onChoose, onVerify, onGrade }: Handlers) {
  if (action?.do === 'choose') onChoose(action.itemId);
  else if (action?.do === 'verify') onVerify();
  else if (action?.do === 'grade') onGrade(action.grade);
}

const hasModifier = (event: KeyboardEvent) => event.metaKey || event.ctrlKey || event.altKey;

export function useExerciseKeys(context: Context, handlers: Handlers) {
  const { card, checked, chosen } = context;
  const { onChoose, onVerify, onGrade } = handlers;
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (ownsKeys(event) || hasModifier(event)) return;
      run(keyAction(event.key, { card, checked, chosen }), { onChoose, onVerify, onGrade });
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [card, checked, chosen, onChoose, onVerify, onGrade]);
}
