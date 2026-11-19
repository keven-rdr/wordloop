import { useState } from 'react';
import type { StudyCard } from '../../../api/generated/types.gen';

// Como a pessoa responde: escolhe uma opção, monta a frase com peças ou digita (digitar, lacuna e imagem).
export type Kind = 'choose' | 'assemble' | 'type';

export function kindOf(exercise: StudyCard['exercise']): Kind {
  if (exercise === 'choose') return 'choose';
  return exercise === 'assemble' ? 'assemble' : 'type';
}

// O que a pessoa já respondeu nesta pergunta (opção escolhida, texto digitado ou peças colocadas) e o que isso permite.
export function useAnswerState(card: StudyCard) {
  const [chosen, setChosen] = useState<string>();
  const [typed, setTyped] = useState('');
  const [placed, setPlaced] = useState<readonly number[]>([]); // posições das peças no cartão, na ordem colocada
  const kind = kindOf(card.exercise);
  const tiles = card.tiles ?? [];
  const remaining = tiles.map((_, index) => index).filter((index) => !placed.includes(index));
  const complete = tiles.length > 0 && remaining.length === 0;
  const assembled = placed.map((index) => tiles[index]).join(' ');
  return {
    kind,
    chosen,
    setChosen,
    typed,
    setTyped,
    placed,
    assemble: { remaining: remaining.length, complete },
    canCheck: { choose: chosen !== undefined, assemble: complete, type: typed.trim() !== '' }[kind],
    answer: { choose: { chosenItemId: chosen }, assemble: { answer: assembled }, type: { answer: typed } }[kind],
    placeTile: (index: number) => setPlaced((list) => [...list, index]),
    placeAt: (position: number) => {
      const index = remaining[position];
      if (index !== undefined) setPlaced((list) => [...list, index]);
    },
    removeAt: (position: number) => setPlaced((list) => list.filter((_, at) => at !== position)),
    undo: () => setPlaced((list) => list.slice(0, -1)),
    clear: () => setPlaced([]),
  };
}
