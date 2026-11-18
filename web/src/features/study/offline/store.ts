import type { ReviewInput, ReviewResult, StudyPack } from '../../../api/generated/types.gen';

// Uma resposta guardada no aparelho. Sem `result`, ainda não foi confirmada pelo servidor.
export type StoredReview = Readonly<{ id: string; sessionId: string; review: ReviewInput; result?: ReviewResult }>;

// O pacote da sessão: com ele a sessão continua sem rede e sobrevive a recarregar a página.
export type StoredPack = Readonly<{ sessionId: string; pack: StudyPack; closed: boolean }>;

export type Store = Readonly<{
  putReviews: (items: readonly StoredReview[]) => Promise<void>;
  reviews: () => Promise<StoredReview[]>;
  deleteReviews: (ids: readonly string[]) => Promise<void>;
  putPack: (item: StoredPack) => Promise<void>;
  packs: () => Promise<StoredPack[]>;
  deletePacks: (ids: readonly string[]) => Promise<void>;
}>;

// Sem IndexedDB (navegação privada em alguns navegadores, testes): guarda na memória da página.
export function memoryStore(): Store {
  const reviews = new Map<string, StoredReview>();
  const packs = new Map<string, StoredPack>();
  return {
    putReviews: async (items) => {
      for (const item of items) reviews.set(item.id, item);
    },
    reviews: async () => [...reviews.values()],
    deleteReviews: async (ids) => {
      for (const id of ids) reviews.delete(id);
    },
    putPack: async (item) => void packs.set(item.sessionId, item),
    packs: async () => [...packs.values()],
    deletePacks: async (ids) => {
      for (const id of ids) packs.delete(id);
    },
  };
}

const DATABASE = 'wordloop-study';
const REVIEWS = 'reviews';
const PACKS = 'packs';

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DATABASE, 1);
    request.onupgradeneeded = () => {
      request.result.createObjectStore(REVIEWS, { keyPath: 'id' });
      request.result.createObjectStore(PACKS, { keyPath: 'sessionId' });
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

type Work<T> = (table: IDBObjectStore) => IDBRequest<T> | undefined;

// Uma transação numa tabela; devolve o resultado do pedido que `work` fizer (se fizer).
function transact<T>(db: IDBDatabase, name: string, write: boolean, work: Work<T>) {
  return new Promise<T | undefined>((resolve, reject) => {
    const tx = db.transaction(name, write ? 'readwrite' : 'readonly');
    const request = work(tx.objectStore(name));
    tx.oncomplete = () => resolve(request?.result);
    tx.onerror = () => reject(tx.error);
    tx.onabort = () => reject(tx.error);
  });
}

export async function indexedDbStore(): Promise<Store> {
  const db = await openDatabase();
  const put = (name: string, items: readonly unknown[]) =>
    transact(db, name, true, (table) => {
      for (const item of items) table.put(item);
      return undefined;
    }).then(() => undefined);
  const all = <T>(name: string) => transact<T[]>(db, name, false, (table) => table.getAll()).then((rows) => rows ?? []);
  const remove = (name: string, ids: readonly string[]) =>
    transact(db, name, true, (table) => {
      for (const id of ids) table.delete(id);
      return undefined;
    }).then(() => undefined);
  return {
    putReviews: (items) => put(REVIEWS, items),
    reviews: () => all<StoredReview>(REVIEWS),
    deleteReviews: (ids) => remove(REVIEWS, ids),
    putPack: (item) => put(PACKS, [item]),
    packs: () => all<StoredPack>(PACKS),
    deletePacks: (ids) => remove(PACKS, ids),
  };
}

let current: Promise<Store> | undefined;

async function choose(): Promise<Store> {
  try {
    return await indexedDbStore();
  } catch {
    return memoryStore();
  }
}

export const store = () => {
  current ??= choose();
  return current;
};

// Só para testes: troca (ou limpa) o armazenamento em uso.
export function setStoreForTests(next?: Store) {
  current = next && Promise.resolve(next);
}
