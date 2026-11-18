import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { p } from './panelStyles';

type Props = Readonly<{ online: boolean; pending: number }>;

// Faixa discreta: sem rede, as respostas ficam salvas no aparelho; com respostas esperando, mostra quantas.
// Some quando está tudo enviado.
export function OfflineBanner({ online, pending }: Props) {
  const { t } = useTranslation('study');
  const waiting = pending > 0;
  return (
    <div role="status" {...stylex.props(!(online && !waiting) && p.banner)}>
      {!online && <p {...stylex.props(p.bannerText)}>{t('offline.banner')}</p>}
      {waiting && <p {...stylex.props(p.bannerText)}>{t('offline.pending', { count: pending })}</p>}
    </div>
  );
}
