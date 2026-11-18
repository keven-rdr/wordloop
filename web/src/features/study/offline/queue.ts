import type { ReviewInput, ReviewResult, StudyPack } from '../../../api/generated/types.gen';
import { type StoredPack, type StoredReview, store } from './store';

const isPending = (item: StoredReview) => !item.result;

// Guarda respostas novas; as que já existem (com ou sem resultado) ficam como estão.
export async function saveReviews(sessionId: string, reviews: readonly ReviewInput[]) {
  const s = await store();
  const known = new Set((await s.reviews()).map((item) => item.id));
  const fresh = reviews.filter((review) => !known.has(review.id));
  await s.putReviews(fresh.map((review) => ({ id: review.id, sessionId, review })));
}

export async function pendingReviews(): Promise<StoredReview[]> {
  return (await (await store()).reviews()).filter(isPending);
}

export async function sessionReviews(sessionId: string): Promise<StoredReview[]> {
  return (await (await store()).reviews()).filter((item) => item.sessionId === sessionId);
}

// O servidor respondeu (aplicada, repetida ou rejeitada): a resposta não precisa ser enviada de novo.
export async function recordResults(results: readonly ReviewResult[]) {
  const s = await store();
  const byId = new Map(results.map((result) => [result.id, result]));
  const settled = (await s.reviews()).filter((item) => byId.has(item.id));
  await s.putReviews(settled.map((item) => ({ ...item, result: byId.get(item.id) })));
}

export const savePack = async (pack: StudyPack) =>
  (await store()).putPack({ sessionId: pack.session_id, pack, closed: false });

export async function loadPack(sessionId: string): Promise<StoredPack | undefined> {
  return (await (await store()).packs()).find((item) => item.sessionId === sessionId);
}

const expired = (item: StoredPack) => Date.parse(item.pack.expires_at) < Date.now();

// Limpa o que já cumpriu o papel: pacotes encerrados ou vencidos e respostas já confirmadas
// de sessões que acabaram. Resposta pendente nunca é apagada.
export async function collect() {
  const s = await store();
  const [packs, reviews] = await Promise.all([s.packs(), s.reviews()]);
  const gone = packs.filter((item) => item.closed || expired(item)).map((item) => item.sessionId);
  await s.deletePacks(gone);
  const alive = new Set(packs.map((item) => item.sessionId).filter((id) => !gone.includes(id)));
  await s.deleteReviews(reviews.filter((item) => item.result && !alive.has(item.sessionId)).map((item) => item.id));
}

export async function closeSession(sessionId: string) {
  const found = await loadPack(sessionId);
  if (found) await (await store()).putPack({ ...found, closed: true });
  await collect();
}
