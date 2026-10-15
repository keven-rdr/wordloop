import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { useLoginProviders } from '../hooks/useLoginProviders';
import { form } from './formStyles';
import { ui } from './styles';

type Provider = 'google' | 'github';

const labels = { google: 'provider.google', github: 'provider.github' } as const;

function ProviderIcon({ provider }: Readonly<{ provider: Provider }>) {
  if (provider === 'google') return <img src="/brand/google.svg" alt="" {...stylex.props(form.providerIcon)} />;
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" {...stylex.props(form.providerIcon)}>
      <path d="M12 .5a11.5 11.5 0 0 0-3.64 22.41c.58.1.79-.25.79-.56v-2c-3.2.7-3.87-1.37-3.87-1.37-.52-1.33-1.28-1.69-1.28-1.69-1.05-.71.08-.7.08-.7 1.15.08 1.76 1.19 1.76 1.19 1.03 1.76 2.69 1.25 3.35.96.1-.75.4-1.25.73-1.54-2.55-.29-5.24-1.28-5.24-5.69 0-1.26.45-2.28 1.19-3.09-.12-.29-.52-1.46.11-3.05 0 0 .97-.31 3.17 1.18a11 11 0 0 1 5.78 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.23 2.76.11 3.05.74.81 1.19 1.83 1.19 3.09 0 4.42-2.7 5.4-5.26 5.68.41.36.78 1.05.78 2.12v3.14c0 .31.21.67.8.56A11.5 11.5 0 0 0 12 .5z" />
    </svg>
  );
}

export function SocialButton({ provider }: Readonly<{ provider: Provider }>) {
  const { t } = useTranslation('auth');
  const enabled = useLoginProviders().includes(provider);
  const content = (
    <>
      <ProviderIcon provider={provider} />
      {t(labels[provider])}
      {!enabled && <span {...stylex.props(form.badge)}>{t('provider.soon')}</span>}
    </>
  );
  if (enabled) {
    return (
      <a
        href={`/api/v1/auth/login?provider=${provider}`}
        {...stylex.props(ui.button, ui.block, form.provider, form.providerOn)}
      >
        {content}
      </a>
    );
  }
  return (
    <button type="button" disabled {...stylex.props(ui.button, ui.block, form.provider, form.providerOff)}>
      {content}
    </button>
  );
}

export function SocialButtons() {
  return (
    <>
      <SocialButton provider="google" />
      <SocialButton provider="github" />
    </>
  );
}

export function Divider({ label }: Readonly<{ label: string }>) {
  return <p {...stylex.props(form.divider)}>{label}</p>;
}
