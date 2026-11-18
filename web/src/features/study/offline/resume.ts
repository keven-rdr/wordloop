import type { ReviewInput, StudyPack } from '../../../api/generated/types.gen';
import { type SessionState, startSession } from '../hooks/sessionReducer';
import { loadPack, sessionReviews } from './queue';

export type Resume = Readonly<{ pack: StudyPack; reviews: readonly ReviewInput[] }>;

// Retoma a sessão do ponto em que parou: sai da fila quem já foi acertado; o resto segue na ordem do pacote.
export function resumeSession(resume: Resume, now: number): SessionState {
  const done = new Set(resume.reviews.filter((review) => review.client_correct).map((review) => review.card_id));
  const queue = resume.pack.cards.filter((card) => !done.has(card.id));
  return { ...startSession(resume.pack.session_id, queue, now), reviews: resume.reviews };
}

// O pacote vem da memória (sessão recém-aberta) ou do aparelho (página recarregada). Sem ele, não há o que retomar.
export async function loadSession(sessionId: string, cached?: StudyPack): Promise<Resume | undefined> {
  const stored = await loadPack(sessionId).catch(() => undefined);
  if (stored?.closed) return undefined;
  const pack = cached ?? stored?.pack;
  if (!pack || Date.parse(pack.expires_at) < Date.now()) return undefined;
  const saved = await sessionReviews(sessionId).catch(() => []);
  const reviews = saved.map((item) => item.review).sort((a, b) => a.answered_at.localeCompare(b.answered_at));
  return { pack, reviews };
}
