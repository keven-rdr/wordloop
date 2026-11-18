import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { eraseMyAccount, exportMyData, logout } from '../../../api/generated/sdk.gen';
import { EraseError, type EraseFailure, exportFileName, failureOf } from '../domain/erase';

// O servidor só aceita escrita com este cabeçalho (proteção contra CSRF).
const csrfHeaders = { 'X-Requested-With': 'XMLHttpRequest' } as const;

// Entrega o JSON como arquivo, sem passar por servidor nenhum: os dados vão direto do navegador para o disco.
export function saveAsFile(data: unknown, name: string) {
  const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  document.body.append(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 0);
}

// Baixa todos os dados da pessoa. Não fica em cache: dado pessoal não deve sobrar na memória do app.
export function useExportData() {
  const mutation = useMutation({
    mutationFn: async () => {
      const { data, response } = await exportMyData();
      if (!response?.ok || !data) throw new Error('export');
      saveAsFile(data, exportFileName(new Date()));
    },
    networkMode: 'always',
  });
  return {
    download: () => mutation.mutate(),
    downloading: mutation.isPending,
    failed: mutation.isError,
    done: mutation.isSuccess,
  };
}

// Sai da conta (apaga o cookie) e volta para o login; tenta uma vez e segue mesmo se falhar.
export function useLeave() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  return async () => {
    await logout({ headers: csrfHeaders }).catch(() => undefined);
    queryClient.clear();
    void navigate({ to: '/login' });
  };
}

// O motivo da falha: o que o servidor disse (EraseError) ou, para qualquer outro erro (rede, por exemplo), 'failed'.
const reasonOf = (error: unknown): EraseFailure => (error instanceof EraseError ? error.failure : 'failed');

// Apaga a conta. Se der certo, sai e volta ao login; se não, o motivo sobe como EraseError para a tela explicar.
export function useEraseAccount() {
  const leave = useLeave();
  const navigate = useNavigate();
  const mutation = useMutation({
    mutationFn: async (confirmation: string) => {
      const { response } = await eraseMyAccount({ headers: csrfHeaders, body: { confirmation } });
      const failure = failureOf(response?.status ?? 0); // sem resposta (rede): o 0 cai em 'failed'
      if (failure) throw new EraseError(failure);
    },
    networkMode: 'always',
    onSuccess: leave,
    onError: (error) => {
      if (error instanceof EraseError && error.failure === 'session') void navigate({ to: '/login' });
    },
  });
  return {
    erase: (confirmation: string) => mutation.mutate(confirmation),
    erasing: mutation.isPending,
    failure: mutation.isError ? reasonOf(mutation.error) : undefined,
  };
}
