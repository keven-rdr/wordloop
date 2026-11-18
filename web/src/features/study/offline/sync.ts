import type { ReviewInput, ReviewResult } from '../../../api/generated/types.gen';
import { collect, pendingReviews, recordResults } from './queue';

export const BATCH_SIZE = 200;

export type Send = (batch: readonly ReviewInput[]) => Promise<readonly ReviewResult[]>;

let inflight: Promise<boolean> | undefined;

async function drain(send: Send): Promise<boolean> {
  try {
    const waiting = await pendingReviews();
    for (let at = 0; at < waiting.length; at += BATCH_SIZE) {
      const batch = waiting.slice(at, at + BATCH_SIZE);
      await recordResults(await send(batch.map((item) => item.review)));
    }
    await collect();
    return true;
  } catch {
    return false;
  }
}

// Envia o que está pendente, em lotes, com os mesmos UUIDs de sempre (o servidor ignora repetidas).
// Devolve false se algo falhou: o que não foi confirmado continua na fila. Nunca roda duas vezes ao mesmo tempo.
export function flushPending(send: Send): Promise<boolean> {
  inflight ??= drain(send).finally(() => {
    inflight = undefined;
  });
  return inflight;
}
