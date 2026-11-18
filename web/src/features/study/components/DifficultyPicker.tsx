import * as stylex from '@stylexjs/stylex';
import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { p } from './panelStyles';

type Props<T extends string> = Readonly<{
  kind: 'difficulty' | 'errorKind';
  values: readonly T[];
  suggested?: T;
  onPick: (value: T) => void;
}>;

// Pergunta de 1 toque: três botões iguais; no acerto, o sugerido pelo tempo ganha marca e o foco inicial.
export function DifficultyPicker<T extends string>({ kind, values, suggested, onPick }: Props<T>) {
  const { t } = useTranslation('study');
  const initial = useRef<HTMLButtonElement>(null);
  useEffect(() => initial.current?.focus(), []);
  const focused = suggested ?? values[0];
  return (
    <section>
      <h2 {...stylex.props(p.question)}>{t(`${kind}.question`)}</h2>
      <div {...stylex.props(p.choices)}>
        {values.map((value) => (
          <button
            key={value}
            ref={value === focused ? initial : undefined}
            type="button"
            onClick={() => onPick(value)}
            {...stylex.props(p.choice, value === suggested && p.suggested)}
          >
            {t(`${kind}.${value}`)}
            {value === suggested && <span {...stylex.props(p.tag)}>{t('difficulty.suggested')}</span>}
          </button>
        ))}
      </div>
    </section>
  );
}
