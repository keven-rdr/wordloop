import { client } from '../api/generated/client.gen';

// O contrato declara servers: /api/v1 (relativo). O fetch precisa de URL absoluta fora do navegador (testes).
export function configureApi(origin: string = window.location.origin) {
  client.setConfig({ baseUrl: `${origin}/api/v1` });
}
