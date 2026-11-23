import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { Goal, GoalSuggestion, GoalSuggestions } from '../../../api/generated/types.gen';
import { type Choice, type Custom, keepRequest, MAX_MINUTES, minutesOf, requestOf } from '../domain/choice';
import { useStartGoal } from '../hooks/useGoals';
import { c, g } from './styles';

const names = { light: 'preset.light', steady: 'preset.steady', intense: 'preset.intense' } as const;

function CustomFields({ custom, onChange }: Readonly<{ custom: Custom; onChange: (next: Custom) => void }>) {
  const { t } = useTranslation('goals');
  const minutesOk = minutesOf(custom) !== undefined;
  return (
    <div {...stylex.props(c.fields)}>
      <label {...stylex.props(c.field)}>
        <span {...stylex.props(c.label)}>{t('choose.minutes')}</span>
        <input
          type="text"
          inputMode="numeric"
          value={custom.minutes}
          aria-invalid={!minutesOk}
          onChange={(event) => onChange({ ...custom, minutes: event.target.value })}
          {...stylex.props(c.input)}
        />
        {!minutesOk && <span {...stylex.props(g.error)}>{t('choose.errorMinutes', { max: MAX_MINUTES })}</span>}
      </label>
    </div>
  );
}

type OptionsProps = Readonly<{
  suggestions: readonly GoalSuggestion[];
  choice: Choice;
  onChoose: (next: Choice) => void;
}>;

// As 3 sugestões e "personalizar", cada uma um cartão com um radio de verdade.
function OptionList({ suggestions, choice, onChoose }: OptionsProps) {
  const { t } = useTranslation('goals');
  return (
    <fieldset {...stylex.props(c.options)}>
      <legend {...stylex.props(c.legend)}>{t('choose.legend')}</legend>
      {suggestions.map((item) => (
        <label key={item.preset} {...stylex.props(c.option, choice === item.preset && c.chosen)}>
          <input
            type="radio"
            name="goal"
            checked={choice === item.preset}
            onChange={() => onChoose(item.preset)}
            {...stylex.props(c.radio)}
          />
          <span {...stylex.props(c.optionText)}>
            <span {...stylex.props(c.optionName)}>{t(names[item.preset])}</span>
            <span {...stylex.props(g.text)}>{t('choose.detail', { minutes: item.daily_minutes })}</span>
          </span>
        </label>
      ))}
      <label {...stylex.props(c.option, choice === 'custom' && c.chosen)}>
        <input
          type="radio"
          name="goal"
          checked={choice === 'custom'}
          onChange={() => onChoose('custom')}
          {...stylex.props(c.radio)}
        />
        <span {...stylex.props(c.optionName)}>{t('preset.custom')}</span>
      </label>
    </fieldset>
  );
}

type Props = Readonly<{ suggestions: GoalSuggestions; goal?: Goal; ended: boolean }>;

// As 3 sugestões e "personalizar". Começar abre o ciclo; com o ciclo anterior acabado há também "Manter a mesma".
export function GoalChooser({ suggestions, goal, ended }: Props) {
  const { t } = useTranslation('goals');
  const [choice, setChoice] = useState<Choice>('steady');
  const [custom, setCustom] = useState<Custom>({ minutes: '30' });
  const { start, saving, failed } = useStartGoal();
  const request = requestOf(choice, suggestions.suggestions, custom);
  return (
    <div {...stylex.props(g.section)}>
      <p {...stylex.props(g.text)}>{suggestions.based_on_history ? t('choose.fromHistory') : t('choose.initial')}</p>
      <OptionList suggestions={suggestions.suggestions} choice={choice} onChoose={setChoice} />
      {choice === 'custom' && <CustomFields custom={custom} onChange={setCustom} />}
      {failed && (
        <p role="alert" {...stylex.props(g.error)}>
          {t('choose.failed')}
        </p>
      )}
      <button
        type="button"
        disabled={!request || saving}
        onClick={() => request && start(request)}
        {...stylex.props(c.action, c.block)}
      >
        {saving ? t('choose.saving') : t('choose.start')}
      </button>
      {ended && goal && (
        <button
          type="button"
          disabled={saving}
          onClick={() => start(keepRequest(goal))}
          {...stylex.props(c.action, c.secondary, c.block)}
        >
          {t('choose.keep')}
        </button>
      )}
    </div>
  );
}
