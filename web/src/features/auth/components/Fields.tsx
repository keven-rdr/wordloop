import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { form } from './formStyles';

type FieldProps = Readonly<{
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  autoComplete: string;
  placeholder?: string;
}>;

export function TextField({ id, label, value, onChange, autoComplete, placeholder }: FieldProps) {
  return (
    <div {...stylex.props(form.field)}>
      <label htmlFor={id} {...stylex.props(form.label)}>
        {label}
      </label>
      <input
        id={id}
        type="email"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        autoComplete={autoComplete}
        placeholder={placeholder}
        {...stylex.props(form.input)}
      />
    </div>
  );
}

export function PasswordField({ id, label, value, onChange, autoComplete }: FieldProps) {
  const { t } = useTranslation('auth');
  const [visible, setVisible] = useState(false);
  return (
    <div {...stylex.props(form.field)}>
      <label htmlFor={id} {...stylex.props(form.label)}>
        {label}
      </label>
      <div {...stylex.props(form.passwordWrap)}>
        <input
          id={id}
          type={visible ? 'text' : 'password'}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete={autoComplete}
          {...stylex.props(form.input, form.inputWithButton)}
        />
        <button
          type="button"
          onClick={() => setVisible((current) => !current)}
          aria-label={t(visible ? 'password.hide' : 'password.show')}
          {...stylex.props(form.reveal)}
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
            {...stylex.props(form.revealIcon)}
          >
            <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" />
            <circle cx="12" cy="12" r="3" />
            {visible && <path d="M4 4l16 16" />}
          </svg>
        </button>
      </div>
    </div>
  );
}
