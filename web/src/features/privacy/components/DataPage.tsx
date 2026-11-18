import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { Navigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { getMeOptions } from '../../../api/generated/@tanstack/react-query.gen';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { useEraseAccount, useExportData, useLeave } from '../hooks/useDataRights';
import { EraseDialog } from './EraseDialog';
import { d } from './styles';

function ExportSection() {
  const { t } = useTranslation('privacy');
  const { download, downloading, failed, done } = useExportData();
  return (
    <section aria-labelledby="export-title" {...stylex.props(d.card)}>
      <h2 id="export-title" {...stylex.props(d.title)}>
        {t('export.title')}
      </h2>
      <p {...stylex.props(d.text)}>{t('export.text')}</p>
      <button type="button" disabled={downloading} onClick={download} {...stylex.props(d.action)}>
        {downloading ? t('export.downloading') : t('export.action')}
      </button>
      <p role="status" {...stylex.props(failed ? d.error : d.status)}>
        {failed && t('export.failed')}
        {done && !failed && t('export.done')}
      </p>
    </section>
  );
}

function EraseSection() {
  const { t } = useTranslation('privacy');
  const [open, setOpen] = useState(false);
  const { erase, erasing, failure } = useEraseAccount();
  const leave = useLeave();
  return (
    <section aria-labelledby="erase-section" {...stylex.props(d.card, d.danger)}>
      <h2 id="erase-section" {...stylex.props(d.title)}>
        {t('erase.title')}
      </h2>
      <p {...stylex.props(d.text)}>{t('erase.text')}</p>
      <ul {...stylex.props(d.list)}>
        <li>{t('erase.items.account')}</li>
        <li>{t('erase.items.study')}</li>
        <li>{t('erase.items.log')}</li>
      </ul>
      <p {...stylex.props(d.text)}>{t('erase.irreversible')}</p>
      <button type="button" onClick={() => setOpen(true)} {...stylex.props(d.action, d.secondary)}>
        {t('erase.open')}
      </button>
      {open && (
        <EraseDialog
          erasing={erasing}
          failure={failure}
          onConfirm={erase}
          onClose={() => setOpen(false)}
          onRelogin={() => void leave()}
        />
      )}
    </section>
  );
}

// "Seus dados": baixar uma cópia de tudo e, se a pessoa quiser, apagar a conta (sem volta). Sem sessão, volta ao login.
export function DataPage() {
  const { t, ready } = useTranslation('privacy');
  const me = useQuery({ ...getMeOptions(), retry: false, refetchOnWindowFocus: false });
  if (me.isError) return <Navigate to="/login" replace />;
  if (!ready || me.isPending) return <div aria-busy="true" />;
  return (
    <BrandScreen title={t('title')} subtitle={t('subtitle')}>
      <div {...stylex.props(d.stack)}>
        <ExportSection />
        <EraseSection />
      </div>
    </BrandScreen>
  );
}
