import { Progress } from '@base-ui/react/progress';
import * as stylex from '@stylexjs/stylex';
import { color, radius } from './tokens/color.stylex';

const styles = stylex.create({
  track: { height: 12, borderRadius: radius.full, backgroundColor: color.surface, overflow: 'hidden' },
  fill: (pct: number) => ({ height: '100%', width: `${pct}%`, backgroundColor: color.primary }),
});

export function ProgressBar({ value }: Readonly<{ value: number }>) {
  return (
    <Progress.Root value={value} {...stylex.props(styles.track)}>
      <Progress.Track {...stylex.props(styles.track)}>
        <Progress.Indicator {...stylex.props(styles.fill(value))} />
      </Progress.Track>
    </Progress.Root>
  );
}
