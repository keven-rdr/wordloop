import type { ReviewInput, StudyCard } from '../../../api/generated/types.gen';
import {
  checkChoice,
  checkTyped,
  type Difficulty,
  type ErrorKind,
  type Outcome,
  requeue,
  suggestDifficulty,
} from '../domain/answer';

// O que foi respondido e ainda espera a pergunta de dificuldade.
export type Checked = Readonly<{
  outcome: Outcome;
  answer?: string;
  chosenItemId?: string;
  hintUsed: boolean;
  answeredAt: number;
  suggested: Difficulty;
}>;

export type SessionState = Readonly<{
  sessionId: string;
  queue: readonly StudyCard[];
  shownAt: number;
  hintUsed: boolean;
  checked?: Checked;
  reviews: readonly ReviewInput[];
}>;

export type SessionAction =
  | Readonly<{ type: 'hint' }>
  | Readonly<{ type: 'check'; now: number; answer?: string; chosenItemId?: string }>
  | Readonly<{ type: 'grade'; now: number; id: string; difficulty?: Difficulty; errorKind?: ErrorKind }>;

export function startSession(sessionId: string, cards: readonly StudyCard[], now: number): SessionState {
  return { sessionId, queue: cards, shownAt: now, hintUsed: false, reviews: [] };
}

function check(state: SessionState, action: Extract<SessionAction, { type: 'check' }>): SessionState {
  const card = state.queue[0];
  if (!card || state.checked) return state;
  const outcome =
    card.exercise === 'type'
      ? checkTyped(card.term, action.answer ?? '')
      : checkChoice(card, action.chosenItemId ?? '');
  const suggested = suggestDifficulty(action.now - state.shownAt, card.exercise);
  const checked: Checked = {
    outcome,
    answer: action.answer,
    chosenItemId: action.chosenItemId,
    hintUsed: state.hintUsed,
    answeredAt: action.now,
    suggested,
  };
  return { ...state, checked };
}

type GradeAction = Extract<SessionAction, { type: 'grade' }>;

// Só vale quando acertou; com dica a nota fica limitada a Difícil.
function difficultyOf(checked: Checked, action: GradeAction): Difficulty | undefined {
  if (checked.outcome !== 'right') return undefined;
  return checked.hintUsed ? 'hard' : action.difficulty;
}

function review(state: SessionState, card: StudyCard, checked: Checked, action: GradeAction): ReviewInput {
  const right = checked.outcome === 'right';
  return {
    id: action.id,
    card_id: card.id,
    session_id: state.sessionId,
    exercise: card.exercise,
    source: 'session',
    shown_at: new Date(state.shownAt).toISOString(),
    answered_at: new Date(checked.answeredAt).toISOString(),
    response_ms: Math.max(0, checked.answeredAt - state.shownAt),
    answer: checked.answer,
    chosen_item_id: checked.chosenItemId,
    client_correct: right,
    hint_used: checked.hintUsed,
    difficulty_suggested: checked.suggested,
    difficulty: difficultyOf(checked, action),
    error_kind: checked.outcome === 'wrong' ? action.errorKind : undefined,
  };
}

function grade(state: SessionState, action: GradeAction): SessionState {
  const [card, ...rest] = state.queue;
  const { checked } = state;
  if (!card || !checked) return state;
  return {
    ...state,
    queue: checked.outcome === 'wrong' ? requeue(rest, card) : rest,
    shownAt: action.now,
    hintUsed: false,
    checked: undefined,
    reviews: [...state.reviews, review(state, card, checked, action)],
  };
}

export function sessionReducer(state: SessionState, action: SessionAction): SessionState {
  switch (action.type) {
    case 'hint':
      return { ...state, hintUsed: true };
    case 'check':
      return check(state, action);
    case 'grade':
      return grade(state, action);
  }
}

// Há pergunta de dificuldade? Acerto sem dica pergunta "fácil, médio ou difícil"; erro pergunta "como foi?".
// Erro de digitação e acerto com dica não perguntam nada.
export function asksDifficulty(checked: Checked): 'difficulty' | 'errorKind' | undefined {
  if (checked.outcome === 'wrong') return 'errorKind';
  if (checked.outcome === 'right' && !checked.hintUsed) return 'difficulty';
  return undefined;
}
