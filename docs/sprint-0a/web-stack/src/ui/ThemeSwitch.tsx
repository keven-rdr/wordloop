import { Switch } from '@base-ui/react/switch';
import * as stylex from '@stylexjs/stylex';
import { color, radius } from './tokens/color.stylex';

const styles = stylex.create({
  root: { width: 52, height: 32, borderRadius: radius.full, backgroundColor: color.surface, borderWidth: 0, padding: 4 },
  thumb: { display: 'block', width: 24, height: 24, borderRadius: radius.full, backgroundColor: color.primary },
});

export function ThemeSwitch(props: Readonly<{ checked: boolean; onChange: (v: boolean) => void; label: string }>) {
  return (
    <Switch.Root checked={props.checked} onCheckedChange={props.onChange} aria-label={props.label} {...stylex.props(styles.root)}>
      <Switch.Thumb {...stylex.props(styles.thumb)} />
    </Switch.Root>
  );
}
