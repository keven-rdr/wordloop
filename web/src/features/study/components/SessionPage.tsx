import * as stylex from '@stylexjs/stylex';
import { useQueryClient } from '@tanstack/react-query';
import { Navigate } from '@tanstack/react-router';
import { useEffect, useReducer, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { StudyPack } from '../../../api/generated/types.gen';
import { sessionReducer, startSession } from '../hooks/sessionReducer';
import { packKey } from '../hooks/summary';
import { type Destination, useFinishSession } from '../hooks/useFinishSession';
import { ExerciseScreen } from './ExerciseScreen';
import { ExitDialog } from './ExitDialog';
import { SendingPanel } from './SendingPanel';
import { SessionTop } from './SessionTop';
import { s } from './styles';

function ActiveSession({ pack }: Readonly<{ pack: StudyPack }>) {
  const [state, dispatch] = useReducer(sessionReducer, undefined, () =>
    startSession(pack.session_id, pack.cards, Date.now()),
  );
  const [leaving, setLeaving] = useState(false);
  const destination = useRef<Destination | undefined>(undefined);
  const { finish, sending, failed } = useFinishSession(pack);
  const card = state.queue[0];

  const end = (to: Destination) => {
    destination.current = to;
    void finish(state.reviews, to);
  };

  useEffect(() => {
    if (!card && !destination.current) end('summary');
  });

  useEffect(() => {
    const onEscape = (event: KeyboardEvent) => event.key === 'Escape' && setLeaving((open) => !open);
    window.addEventListener('keydown', onEscape);
    return () => window.removeEventListener('keydown', onEscape);
  }, []);

  if (!card || sending || failed) {
    return (
      <div {...stylex.props(s.screen)}>
        <SendingPanel failed={failed} onRetry={() => end(destination.current ?? 'summary')} />
      </div>
    );
  }
  return (
    <div {...stylex.props(s.screen)}>
      <SessionTop done={state.reviews.length} left={state.queue.length} onClose={() => setLeaving(true)} />
      <ExerciseScreen
        key={`${card.id}-${state.reviews.length}`}
        card={card}
        checked={state.checked}
        hintUsed={state.hintUsed}
        onHint={() => dispatch({ type: 'hint' })}
        onCheck={(answer) => dispatch({ type: 'check', now: Date.now(), ...answer })}
        onGrade={(grade) => dispatch({ type: 'grade', now: Date.now(), id: crypto.randomUUID(), ...grade })}
      />
      {leaving && <ExitDialog onStay={() => setLeaving(false)} onLeave={() => end('plan')} />}
    </div>
  );
}

// O pacote da sessão vive no cache do app: sem ele (página recarregada), volta para o plano.
export function SessionPage({ sessionId }: Readonly<{ sessionId: string }>) {
  const { ready } = useTranslation('study');
  const pack = useQueryClient().getQueryData<StudyPack>(packKey(sessionId));
  if (!pack) return <Navigate to="/study" replace />;
  if (!ready) return <div aria-busy="true" />;
  return <ActiveSession pack={pack} />;
}
