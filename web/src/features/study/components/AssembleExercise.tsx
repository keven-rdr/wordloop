import * as stylex from '@stylexjs/stylex';
import { useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { StudyCard } from '../../../api/generated/types.gen';
import type { Outcome } from '../domain/answer';
import { s } from './styles';
import { tile } from './tileStyles';

type RowProps = Readonly<{
  tiles: readonly string[];
  placed: readonly number[];
  outcome?: Outcome;
  onRemove: (position: number) => void;
}>;

// A frase como a pessoa a montou: tocar numa peça a devolve ao banco.
function AnswerRow({ tiles, placed, outcome, onRemove }: RowProps) {
  const { t } = useTranslation('study');
  const locked = outcome !== undefined;
  return (
    <section aria-label={t('assemble.answer')}>
      <ol lang="en" {...stylex.props(tile.row)}>
        {placed.length === 0 && <li {...stylex.props(tile.empty)}>{t('assemble.empty')}</li>}
        {placed.map((index, position) => (
          <li key={index}>
            <button
              type="button"
              disabled={locked}
              aria-label={`${t('assemble.remove')} ${tiles[index]}`}
              onClick={() => onRemove(position)}
              {...stylex.props(
                tile.piece,
                tile.placed,
                outcome === 'right' && tile.right,
                outcome === 'wrong' && tile.wrong,
              )}
            >
              {tiles[index]}
            </button>
          </li>
        ))}
      </ol>
      <p aria-live="polite" {...stylex.props(s.srOnly)}>
        {placed.map((index) => tiles[index]).join(' ')}
      </p>
    </section>
  );
}

type Props = Readonly<{
  card: StudyCard;
  // Posições (no banco de peças do cartão) na ordem em que a pessoa as colocou.
  placed: readonly number[];
  outcome?: Outcome;
  onPlace: (index: number) => void;
  onRemove: (position: number) => void;
  onClear: () => void;
}>;

// Montar a frase: a pergunta é a tradução; a resposta, as peças em inglês tocadas na ordem certa.
// Tocar numa peça do banco a coloca na linha; tocar numa peça da linha a devolve. Teclado: 1–9, Backspace e Enter.
export function AssembleExercise({ card, placed, outcome, onPlace, onRemove, onClear }: Props) {
  const { t } = useTranslation('study');
  const bank = useRef<HTMLUListElement>(null);
  const tiles = card.tiles ?? [];
  const locked = outcome !== undefined;
  const remaining = tiles.map((word, index) => ({ word, index })).filter(({ index }) => !placed.includes(index));
  // A peça tocada sai do banco (ou volta para ele): o foco vai para a primeira do banco, para o teclado não se perder.
  const refocus = () => requestAnimationFrame(() => bank.current?.querySelector('button')?.focus());
  const act = (change: () => void) => {
    change();
    refocus();
  };
  return (
    <>
      <p {...stylex.props(s.hint)}>{t('assemble.prompt')}</p>
      <h1 lang="pt-BR" {...stylex.props(s.term)}>
        {card.gloss}
      </h1>
      <AnswerRow
        tiles={tiles}
        placed={placed}
        outcome={outcome}
        onRemove={(position) => act(() => onRemove(position))}
      />
      <ul ref={bank} lang="en" aria-label={t('assemble.bank')} {...stylex.props(tile.bank)}>
        {remaining.map(({ word, index }, position) => (
          <li key={index}>
            <button
              type="button"
              disabled={locked}
              onClick={() => act(() => onPlace(index))}
              {...stylex.props(tile.piece)}
            >
              <span aria-hidden="true" {...stylex.props(tile.key)}>
                {position + 1}
              </span>
              {word}
            </button>
          </li>
        ))}
      </ul>
      {!locked && placed.length > 0 && (
        <button type="button" onClick={onClear} {...stylex.props(s.link)}>
          {t('assemble.clear')}
        </button>
      )}
    </>
  );
}
