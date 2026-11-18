import * as stylex from '@stylexjs/stylex';
import { p } from './styles';

type Option<T extends string> = Readonly<{ value: T; label: string }>;

type Props<T extends string> = Readonly<{
  legend: string;
  name: string;
  value: T;
  options: readonly Option<T>[];
  onChange: (value: T) => void;
}>;

// Um grupo de opções: cada uma é um cartão com um radio de verdade dentro (teclado e leitor de tela de graça).
export function ChoiceGroup<T extends string>({ legend, name, value, options, onChange }: Props<T>) {
  return (
    <fieldset {...stylex.props(p.section)}>
      <legend {...stylex.props(p.legend)}>{legend}</legend>
      {options.map((option) => (
        <label key={option.value} {...stylex.props(p.option, value === option.value && p.chosen)}>
          <input
            type="radio"
            name={name}
            checked={value === option.value}
            onChange={() => onChange(option.value)}
            {...stylex.props(p.control)}
          />
          <span {...stylex.props(p.name)}>{option.label}</span>
        </label>
      ))}
    </fieldset>
  );
}
