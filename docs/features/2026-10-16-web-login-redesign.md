# Web: nova tela de entrar (marca, tema claro/escuro e cadastro)

> Arquivo: `docs/features/2026-10-16-web-login-redesign.md` · Issue: #__ · Branch: `feat/web-login-redesign` · Estado: BUILD

## Contexto
A tela de entrar era só um título e um botão. O dono pediu uma tela de entrada no estilo das referências (painel de marca ao lado de um cartão de login), com as cores e o estilo do favicon, tema animado e rotas para as outras formas de entrar: criar conta e entrar com Google.

## Definições confirmadas
- Cores e estilo vêm do favicon: violeta `#7B5CE0`, rosa `#EE85B5`, amarelo `#EBC33A` e papel branco. Trocam o verde-água provisório dos tokens.
- Rotas: `/login` (passkey, Google, link para criar conta) e `/signup` (criar conta). O início (`/`) sem sessão leva a `/login`; com sessão mostra "Olá" e as ações da conta.
- Tema claro e escuro, seguindo o sistema por padrão, com botão para trocar (a escolha fica guardada no navegador).
- Animação respeita `prefers-reduced-motion`.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Layout | tela dividida: painel da marca (gradiente que se move, três cartões do favicon flutuando e luzes) e cartão do formulário; no celular o painel vira uma faixa no topo | tela centralizada simples: sem a identidade pedida |
| Google | botão visível, **desligado** e marcado "em breve" | ligá-lo ao OIDC que já existe: ele criaria contas novas sozinho, contra a decisão do [ADR 0028](../adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md) (Google só vincula conta que já existe). Entra na PR 4 |
| Cadastro por e-mail | a rota e o formulário existem, mas o campo e o botão ficam desligados com o aviso de que o acesso é por convite | formulário que finge funcionar: o envio de e-mail (Brevo) só chega na PR 3 |
| Tema | vars de cor do StyleX seguem `prefers-color-scheme`; o botão aplica `lightTheme`/`darkTheme` na raiz e guarda a escolha em `localStorage` (com `try/catch`) | biblioteca de tema: uma dependência para o que 40 linhas resolvem |
| Tokens | cores da marca em `brand.stylex.ts`; `size.touch` (48 px) e `size.card` em `color.stylex.ts` | cores soltas nos componentes: proibido (regra 8) |
| Sem chave crua | a faixa da marca usa o nome do app do namespace `common` (sempre carregado) e só mostra o texto longo quando `auth` carregou | mostrar a chave até carregar: o teste `i18n` reprova |

## Fora desta PR
Botão de ligar/desligar o 2FA (depende do contrato da [PR do 2FA opcional](2026-10-16-2fa-opcional.md)); cadastro por e-mail real (PR 3); Google e GitHub reais (PR 4); animações de transição entre telas.

## Checklist de progresso
- [x] Tokens da marca, tema claro/escuro e botão de trocar
- [x] `AuthShell`, painel da marca animado, `LoginPage`, `SignupPage`; convite no mesmo layout
- [x] Início sem sessão leva a `/login`; páginas internas com `AppPage`
- [x] Textos `pt-BR` e `en-US`
- [x] Testes: Vitest (navegação, redirecionamentos, tema) e Playwright (fluxo login → cadastro → login, troca de tema)
- [x] Conferido no navegador em desktop e celular, claro e escuro
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: `npm run test` e `npm run test:e2e`. Manual: `npm run dev:web`, abrir `/` (deve ir a `/login`), alternar o tema, abrir `Criar conta` e voltar; reduzir movimento no sistema e conferir que a animação para.
