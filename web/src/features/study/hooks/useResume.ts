import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import type { StudyPack } from '../../../api/generated/types.gen';
import { loadSession, type Resume } from '../offline/resume';
import { packKey } from './summary';

// Acha a sessão para retomar: o pacote em memória ou o guardado no aparelho, mais o que já foi respondido.
// 'loading' enquanto procura; 'missing' se não há o que retomar (vencida, encerrada ou desconhecida).
export function useResume(sessionId: string): Resume | 'loading' | 'missing' {
  const queryClient = useQueryClient();
  const [cached] = useState(() => queryClient.getQueryData<StudyPack>(packKey(sessionId)));
  const [found, setFound] = useState<Resume | 'loading' | 'missing'>('loading');
  useEffect(() => {
    let live = true;
    void loadSession(sessionId, cached).then((resume) => live && setFound(resume ?? 'missing'));
    return () => {
      live = false;
    };
  }, [sessionId, cached]);
  return found;
}
