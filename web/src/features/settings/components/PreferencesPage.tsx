import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { Settings } from '../../../api/generated/types.gen';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { LOCALES, THEMES } from '../domain/prefs';
import { useSaveSettings, useSettings } from '../hooks/useSettings';
import { ChoiceGroup } from './ChoiceGroup';
import { p } from './styles';
import { TimeZoneSection } from './TimeZoneSection';

// Cada mudança é salva na hora (as preferências inteiras) e já vale na tela.
function Form({ settings }: Readonly<{ settings: Settings }>) {
  const { t } = useTranslation('settings');
  const { save, saving, failed, saved } = useSaveSettings();
  let message = '';
  if (saving) message = t('saving');
  else if (saved) message = t('saved');
  return (
    <div {...stylex.props(p.stack)}>
      <ChoiceGroup
        legend={t('locale.legend')}
        name="ui_locale"
        value={settings.ui_locale}
        options={LOCALES.map((value) => ({ value, label: t(`locale.${value}`) }))}
        onChange={(ui_locale) => save({ ...settings, ui_locale })}
      />
      <ChoiceGroup
        legend={t('theme.legend')}
        name="theme"
        value={settings.theme}
        options={THEMES.map((value) => ({ value, label: t(`theme.${value}`) }))}
        onChange={(theme) => save({ ...settings, theme })}
      />
      <TimeZoneSection settings={settings} onChange={save} />
      <p role="status" {...stylex.props(p.status)}>
        {message}
      </p>
      {failed && (
        <p role="alert" {...stylex.props(p.error)}>
          {t('failed')}
        </p>
      )}
    </div>
  );
}

// A tela de preferências: idioma, tema e fuso horário.
export function PreferencesPage() {
  const { t, ready } = useTranslation('settings');
  const settings = useSettings();
  if (!ready || settings.isPending) return <div aria-busy="true" />;
  return (
    <BrandScreen title={t('title')} subtitle={t('subtitle')}>
      {settings.isError ? (
        <p role="status" {...stylex.props(p.error)}>
          {t('loadFailed')}
        </p>
      ) : (
        <Form settings={settings.data} />
      )}
    </BrandScreen>
  );
}
