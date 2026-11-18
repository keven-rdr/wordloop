import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { Settings } from '../../../api/generated/types.gen';
import { detectTimeZone, followDevice, pinZone, timeZones } from '../domain/prefs';
import { p } from './styles';

// O fuso define a virada do dia de estudo. Por padrão acompanha o aparelho; escolher um o deixa fixo.
export function TimeZoneSection({
  settings,
  onChange,
}: Readonly<{ settings: Settings; onChange: (next: Settings) => void }>) {
  const { t } = useTranslation('settings');
  const device = detectTimeZone();
  return (
    <fieldset {...stylex.props(p.section)}>
      <legend {...stylex.props(p.legend)}>{t('zone.legend')}</legend>
      <label {...stylex.props(p.option, settings.tz_auto && p.chosen)}>
        <input
          type="checkbox"
          checked={settings.tz_auto}
          onChange={(event) =>
            onChange(event.target.checked ? followDevice(settings, device) : pinZone(settings, settings.tz))
          }
          {...stylex.props(p.control)}
        />
        <span {...stylex.props(p.texts)}>
          <span {...stylex.props(p.name)}>{t('zone.auto')}</span>
          <span {...stylex.props(p.hint)}>{t('zone.detected', { zone: device })}</span>
        </span>
      </label>
      {!settings.tz_auto && (
        <label {...stylex.props(p.field)}>
          <span {...stylex.props(p.label)}>{t('zone.choose')}</span>
          <select
            value={settings.tz}
            onChange={(event) => onChange(pinZone(settings, event.target.value))}
            {...stylex.props(p.select)}
          >
            {timeZones(settings.tz, device).map((zone) => (
              <option key={zone} value={zone}>
                {zone}
              </option>
            ))}
          </select>
        </label>
      )}
      <p {...stylex.props(p.hint)}>{t('zone.hint', { zone: settings.tz })}</p>
    </fieldset>
  );
}
