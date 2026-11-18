import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { Trans } from 'react-i18next';
import { form } from './formStyles';
import { ui } from './styles';

// "Ao continuar, você aceita os Termos e a Política": os dois links levam às páginas públicas.
export function LegalNotice() {
  return (
    <p {...stylex.props(form.footnote)}>
      <Trans
        i18nKey="legal.agree"
        ns="auth"
        components={{
          terms: <Link to="/terms" {...stylex.props(ui.link)} />,
          privacy: <Link to="/privacy" {...stylex.props(ui.link)} />,
        }}
      />
    </p>
  );
}
