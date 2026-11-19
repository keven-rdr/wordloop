import { useEffect } from 'react';
import type { StudyCard } from '../../../api/generated/types.gen';
import { DIFFICULTIES, type Difficulty, ERROR_KINDS, type ErrorKind } from '../domain/answer';
import { asksDifficulty, type Checked } from './sessionReducer';

const LETTERS = ['a', 'b', 'c', 'd'];

export type Grade = Readonly<{ difficulty?: Difficulty; errorKind?: ErrorKind }>;

export type KeyAction =
  | Readonly<{ do: 'choose'; itemId: string }>
  | Readonly<{ do: 'place'; position: number }>
  | Readonly<{ do: 'undo' }>
  | Readonly<{ do: 'verify' }>
  | Readonly<{ do: 'grade'; grade: Grade }>;

// Montar frase: quantas peças ainda estão no banco e se a frase já está completa.
export type AssembleState = Readonly<{ remaining: number; complete: boolean }>;

type Context = Readonly<{ card: StudyCard; checked?: Checked; chosen?: string; assemble?: AssembleState }>;

// Posição 0–3 de uma tecla de opção (A–D ou 1–4), ou -1.
const optionIndex = (key: string) => (LETTERS.includes(key) ? LETTERS.indexOf(key) : Number(key) - 1);

// Montar frase: 1–9 colocam a peça daquela posição do banco, Backspace desfaz a última, Enter verifica.
function assembleKey(key: string, { remaining, complete }: AssembleState): KeyAction | undefined {
  if (key === 'enter') return complete ? { do: 'verify' } : undefined;
  if (key === 'backspace') return { do: 'undo' };
  const position = Number(key) - 1;
  return Number.isInteger(position) && position >= 0 && position < remaining ? { do: 'place', position } : undefined;
}

// Escolher: A–D ou 1–4 escolhem a opção; digitar (lacuna, imagem...) só tem o Enter, que verifica com algo respondido.
function answerKey(key: string, { card, chosen }: Context): KeyAction | undefined {
  const options = card.exercise === 'choose' ? card.options : undefined;
  const option = options?.[optionIndex(key)];
  if (option) return { do: 'choose', itemId: option.item_id };
  return key === 'enter' && chosen ? { do: 'verify' } : undefined;
}

function beforeCheck(key: string, context: Context): KeyAction | undefined {
  const { card, assemble } = context;
  if (card.exercise !== 'assemble') return answerKey(key, context);
  return assemble ? assembleKey(key, assemble) : undefined;
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

// Atalhos (desktop): A–D ou 1–4 escolhem, Enter verifica/continua, 1–3 respondem a dificuldade; na montagem da frase,
// 1–9 colocam peças e Backspace desfaz.
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
  onPlace?: (position: number) => void;
  onUndo?: () => void;
}>;

const noop = () => undefined;

function run(action: KeyAction | undefined, { onChoose, onVerify, onGrade, onPlace = noop, onUndo = noop }: Handlers) {
  if (!action) return;
  switch (action.do) {
    case 'choose':
      return onChoose(action.itemId);
    case 'place':
      return onPlace(action.position);
    case 'undo':
      return onUndo();
    case 'verify':
      return onVerify();
    case 'grade':
      return onGrade(action.grade);
  }
}

const hasModifier = (event: KeyboardEvent) => event.metaKey || event.ctrlKey || event.altKey;

export function useExerciseKeys(context: Context, handlers: Handlers) {
  const { card, checked, chosen, assemble } = context;
  const { onChoose, onVerify, onGrade, onPlace, onUndo } = handlers;
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (ownsKeys(event) || hasModifier(event)) return;
      const action = keyAction(event.key, { card, checked, chosen, assemble });
      // Tecla tratada aqui não pode ainda "apertar" o botão que ganhou o foco com a resposta (o Enter que verifica
      // não pode escolher a dificuldade sugerida logo em seguida).
      if (action) event.preventDefault();
      run(action, { onChoose, onVerify, onGrade, onPlace, onUndo });
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [card, checked, chosen, assemble, onChoose, onVerify, onGrade, onPlace, onUndo]);
}
