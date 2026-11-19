import type { ReviewInput, StudyCard } from '../../../api/generated/types.gen';

export type Difficulty = NonNullable<ReviewInput['difficulty']>;
export type ErrorKind = NonNullable<ReviewInput['error_kind']>;

// Ordem dos botões = ordem das teclas 1, 2 e 3.
export const DIFFICULTIES: readonly Difficulty[] = ['hard', 'medium', 'easy'];
export const ERROR_KINDS: readonly ErrorKind[] = ['no_idea', 'almost', 'slip'];

// Resultado provisório, só para dar feedback na hora. Quem corrige de verdade é o servidor.
export type Outcome = 'right' | 'typo' | 'wrong';

// Só palavras com 5 letras ou mais toleram um erro de digitação (mesma regra do servidor).
const MIN_TYPO_LENGTH = 5;

// Limiares da dificuldade sugerida (ms). No digitar soma-se um tempo extra, porque digitar demora mais.
export const EASY_WITHIN_MS = 3000;
export const MEDIUM_WITHIN_MS = 8000;
export const TYPING_EXTRA_MS = 4000;

export function normalize(text: string): string {
  return text.normalize('NFD').replaceAll(/\p{M}/gu, '').toLowerCase().split(/\s+/u).filter(Boolean).join(' ');
}

// Uma troca, uma inserção, uma remoção ou duas letras vizinhas invertidas.
function oneEditApart(a: string, b: string): boolean {
  if (Math.abs(a.length - b.length) > 1) return false;
  let i = 0;
  while (i < a.length && i < b.length && a[i] === b[i]) i++;
  if (a.length === b.length) {
    const swapped = a[i] === b[i + 1] && a[i + 1] === b[i] && a.slice(i + 2) === b.slice(i + 2);
    return a.slice(i + 1) === b.slice(i + 1) || swapped;
  }
  const [long, short] = a.length > b.length ? [a, b] : [b, a];
  return long.slice(i + 1) === short.slice(i);
}

export function checkTyped(expected: string, typed: string): Outcome {
  const want = normalize(expected);
  const got = normalize(typed);
  if (got === want && got !== '') return 'right';
  if (got !== '' && want.length >= MIN_TYPO_LENGTH && oneEditApart(want, got)) return 'typo';
  return 'wrong';
}

// Palavras da frase sem a pontuação das pontas (mesma regra do servidor: peças não têm vírgula nem ponto final).
const EDGE_PUNCTUATION = /^[.,?!;:"()]+|[.,?!;:"()]+$/gu;

function wordsOf(text: string): string[] {
  return text
    .split(/\s+/u)
    .map((word) => normalize(word.replaceAll(EDGE_PUNCTUATION, '')))
    .filter(Boolean);
}

// Montar a frase: as mesmas palavras na mesma ordem. Nunca é "quase": uma peça fora do lugar é erro.
export function checkAssembled(expected: string, assembled: string): Outcome {
  const want = wordsOf(expected);
  const got = wordsOf(assembled);
  const same = got.length === want.length && want.every((word, index) => word === got[index]);
  return same && got.length > 0 ? 'right' : 'wrong';
}

export function checkChoice(card: StudyCard, chosenItemId: string): Outcome {
  return chosenItemId === card.item_id ? 'right' : 'wrong';
}

export function suggestDifficulty(responseMs: number, exercise: StudyCard['exercise']): Difficulty {
  const extra = exercise === 'choose' ? 0 : TYPING_EXTRA_MS;
  if (responseMs <= EASY_WITHIN_MS + extra) return 'easy';
  if (responseMs <= MEDIUM_WITHIN_MS + extra) return 'medium';
  return 'hard';
}

// O cartão que errou volta daqui a 3 cartões (ou ao fim, se a fila for menor).
export const RETRY_AFTER = 3;

export function requeue(queue: readonly StudyCard[], card: StudyCard): StudyCard[] {
  const at = Math.min(RETRY_AFTER, queue.length);
  return [...queue.slice(0, at), card, ...queue.slice(at)];
}
