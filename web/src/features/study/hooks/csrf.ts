// O servidor só aceita escrita com este cabeçalho (proteção contra CSRF).
export const csrfHeaders = { 'X-Requested-With': 'XMLHttpRequest' } as const;
