import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { LeaderboardProfile } from '../../../api/generated/types.gen';
import { panel } from '../../../shared/ui/panelStyles';
import { type SaveResult, useSaveProfile } from '../hooks/useLeaderboard';
import { l } from './styles';

// Mesmo formato do servidor (que é quem decide): 3 a 20 letras sem acento, números ou _.
const NICKNAME = /^\w{3,20}$/;

const messages: Record<Exclude<SaveResult, 'saved'>, string> = {
  taken: 'join.taken',
  invalid: 'join.invalid',
  failed: 'join.failed',
};

// Participar do ranking: apelido (o nome do perfil nunca aparece) e se aparece para os outros.
export function JoinRanking({ profile }: Readonly<{ profile: LeaderboardProfile }>) {
  const { t } = useTranslation('leaderboard');
  const [nickname, setNickname] = useState(profile.nickname);
  const [visible, setVisible] = useState(profile.nickname === '' ? true : profile.visible);
  const { save, saving, result } = useSaveProfile();
  const valid = NICKNAME.test(nickname);
  return (
    <section aria-labelledby="ranking-join" {...stylex.props(panel.card)}>
      <h2 id="ranking-join" {...stylex.props(panel.heading)}>
        {t('join.title')}
      </h2>
      <p {...stylex.props(panel.text)}>{t('join.text')}</p>
      <div {...stylex.props(panel.field)}>
        <label {...stylex.props(panel.field)}>
          <span {...stylex.props(panel.label)}>{t('join.nickname')}</span>
          <input
            type="text"
            autoComplete="nickname"
            maxLength={20}
            value={nickname}
            aria-invalid={nickname !== '' && !valid}
            aria-describedby="ranking-nickname-hint"
            onChange={(event) => setNickname(event.target.value.trim())}
            {...stylex.props(panel.input)}
          />
        </label>
        <span id="ranking-nickname-hint" {...stylex.props(panel.text)}>
          {t('join.hint')}
        </span>
      </div>
      <label {...stylex.props(l.check)}>
        <input
          type="checkbox"
          checked={visible}
          onChange={(event) => setVisible(event.target.checked)}
          {...stylex.props(l.box)}
        />
        {t('join.visible')}
      </label>
      <output {...stylex.props(panel.status)}>
        {result === 'saved' && <p {...stylex.props(panel.good)}>{t('join.saved')}</p>}
        {result && result !== 'saved' && <p {...stylex.props(panel.error)}>{t(messages[result])}</p>}
      </output>
      <button
        type="button"
        disabled={!valid || saving}
        onClick={() => save({ nickname, visible })}
        {...stylex.props(panel.action)}
      >
        {saving ? t('join.saving') : t('join.save')}
      </button>
    </section>
  );
}
