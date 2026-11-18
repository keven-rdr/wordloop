import * as stylex from '@stylexjs/stylex';
import { Navigate } from '@tanstack/react-router';
import { useEffect, useReducer, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { sessionReducer } from '../hooks/sessionReducer';
import { type Destination, useFinishSession } from '../hooks/useFinishSession';
import { useOfflineQueue } from '../hooks/useOfflineQueue';
import { useResume } from '../hooks/useResume';
import { saveReviews } from '../offline/queue';
import { type Resume, resumeSession } from '../offline/resume';
import { ExerciseScreen } from './ExerciseScreen';
import { ExitDialog } from './ExitDialog';
import { OfflineBanner } from './OfflineBanner';
import { SendingPanel } from './SendingPanel';
import { SessionTop } from './SessionTop';
import { s } from './styles';

const FLUSH_EVERY = 10;

function ActiveSession({ resume }: Readonly<{ resume: Resume }>) {
  const { pack } = resume;
  const [state, dispatch] = useReducer(sessionReducer, undefined, () => resumeSession(resume, Date.now()));
  const [leaving, setLeaving] = useState(false);
  const destination = useRef<Destination | undefined>(undefined);
  const { online, pending, flush, refresh } = useOfflineQueue();
  const { finish, sending } = useFinishSession(pack, flush);
  const card = state.queue[0];

  const end = (to: Destination) => {
    destination.current = to;
    void finish(state.reviews, to);
  };

  // Cada resposta vai para o aparelho na hora (nada se perde sem rede); a cada 10, tenta enviar.
  const answered = state.reviews.length;
  useEffect(() => {
    if (answered === 0) return;
    void saveReviews(pack.session_id, state.reviews.slice(-1))
      .then(async () => {
        if (answered % FLUSH_EVERY === 0) await flush();
        else await refresh();
      })
      .catch(() => undefined);
  }, [answered, state.reviews, pack.session_id, flush, refresh]);

  useEffect(() => {
    if (!card && !destination.current) end('summary');
  });

  useEffect(() => {
    const onEscape = (event: KeyboardEvent) => event.key === 'Escape' && setLeaving((open) => !open);
    window.addEventListener('keydown', onEscape);
    return () => window.removeEventListener('keydown', onEscape);
  }, []);

  if (!card || sending) {
    return (
      <div {...stylex.props(s.screen)}>
        <SendingPanel />
      </div>
    );
  }
  return (
    <div {...stylex.props(s.screen)}>
      <SessionTop done={answered} left={state.queue.length} onClose={() => setLeaving(true)} />
      <OfflineBanner online={online} pending={pending} />
      <ExerciseScreen
        key={`${card.id}-${answered}`}
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

// A sessão vem da memória do app ou do aparelho (página recarregada): sem nenhuma das duas, volta para o plano.
export function SessionPage({ sessionId }: Readonly<{ sessionId: string }>) {
  const { ready } = useTranslation('study');
  const resume = useResume(sessionId);
  if (resume === 'missing') return <Navigate to="/study" replace />;
  if (!ready || resume === 'loading') return <div aria-busy="true" />;
  return <ActiveSession resume={resume} />;
}
