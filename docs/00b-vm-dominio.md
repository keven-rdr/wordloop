# VM e domínio — recomendações

> Data: 2026-10-06. Preços vêm de páginas de terceiros (não do site do provedor): `[VERIFICAR]` antes de comprar.

## 1. VM

| Opção | Specs | Preço/mês | Observação |
|---|---|---|---|
| **Contabo Cloud VPS 10** | 4 vCPU, 8 GB RAM, 75 GB NVMe (ou 150 GB SSD), 200 Mbit/s | ~€4,50 | **Recomendada para começar** |
| Contabo Cloud VPS 20 (a que você listou) | 6 vCPU, 12 GB RAM, 100 GB NVMe (ou 200 GB SSD), 300 Mbit/s | ~€7,00 | Folga para observabilidade e runner |
| Hetzner CX23 | 2 vCPU, 4 GB, 40 GB | ~€3,99 + €0,50 IPv4 | Apertado para Keycloak + 2 ambientes |
| Hetzner CX33 | 4 vCPU, 8 GB, 80 GB | ~€6,49 + €0,50 IPv4 | Reputação melhor de suporte; reajuste de jun/2026 |

Fontes: [onedollarvps.com (Contabo)](https://onedollarvps.com/pricing/contabo-pricing), [vpsbenchmarks.com](https://www.vpsbenchmarks.com/hosters/contabo/plans/cloud-vps-20), [findstack (Hetzner)](https://findstack.com/resources/hetzner-price-increase-2026). Nenhuma é o site oficial.

**Estimativa de RAM** `[PREMISSA]`, para tst + prd na mesma VM: Keycloak ~1 GB, Postgres ~0,6 GB (uma instância, dois bancos), 2 APIs Go ~0,2 GB, proxy e nginx ~0,2 GB, SO e Docker ~0,6 GB → **~3 GB sem observabilidade; ~4,5 GB com Prometheus/Loki/Grafana**. 8 GB basta; 12 GB é conforto, não necessidade. Diferença ~€2,50/mês. Se quiser comprar já a de 12 GB, a decisão é razoável.

**Atualização (Fase 3):** com **um Keycloak por ambiente** (~1,25–1,5 GB cada), a conta sobe para **~5–6 GB sem observabilidade**. Isso aperta os 8 GB da Cloud VPS 10; a **Cloud VPS 20 (12 GB), que você escolheu, passa a ser a recomendação.**

**Decisão final do usuário (pós-Fase 3):** um único Keycloak com dois realms, então a estimativa volta a **~3,5–4,5 GB sem observabilidade**; 8 GB comportaria, mas o usuário escolheu os 12 GB (Cloud VPS 20), o que dá folga para observabilidade e backups locais.

**Cuidados:** (1) o preço promocional pode depender do prazo de contrato e haver taxa de instalação `[VERIFICAR]`; (2) VPS barato tem desempenho de disco variável: **backup fora da VM** (Backblaze B2 ou Cloudflare R2, ambos com camada gratuita `[VERIFICAR]`) e teste de restauração; (3) escolha a região mais próxima do Brasil que a Contabo oferecer `[VERIFICAR]`; (4) porta 25 costuma ficar bloqueada: e-mail por provedor transacional; (5) upgrade costuma ser possível, rebaixar não `[VERIFICAR]`.

## 2. Domínio

Verifiquei por RDAP (`rdap.registro.br`, `rdap.verisign.com`, `pubapi.registry.google`): 404 = não registrado. **É indicativo; a confirmação real é no registrador** (nomes reservados e premium existem).

| Nome | .com.br | .com | .app | Comentário |
|---|---|---|---|---|
| **palavreiro** | livre | livre | livre | pt-BR, memorável; "palavreiro" é quem fala muito, o que combina com vocabulário |
| **inglesloop** | livre | livre | livre | descritivo, mistura pt/en |
| wordloop | livre | ocupado | ocupado | ciclo de repetição |
| vocaloop | livre | ocupado | ocupado | idem |
| lexiloop | livre | ocupado | ocupado | idem |
| reword | livre | ocupado | ocupado | "reescrever/rever palavras" |
| ebbo | livre | ocupado | ocupado | alusão a Ebbinghaus |
| esquecimento | livre | ocupado | livre | brinca com a curva do esquecimento |

Ocupados em todas: ebbi, lembra, lembrei, revoca, palavrinha, relembra, vocabo, wordwise, falaai, recordo, lexo, memorize.

**Recomendação:** `palavreiro.com.br` (+ `.app` se quiser proteger o nome). `.com.br` no Registro.br tem custo anual baixo e exige CPF `[VERIFICAR]`; `.app` custa em torno de US$ 14/ano `[VERIFICAR]` e força HTTPS (HSTS), o que combina com PWA. **Evite "Anki" no nome do produto** (marca de terceiros); o repositório foi renomeado para `wordloop` (antes `poorAnki`).

**Escolha do usuário:** **wordloop** (`wordloop.com.br` livre pelo RDAP de 06/10/2026; `.com` e `.app` ocupados). Registrar o `.com.br` no Registro.br.

## 3. MIT

Aguardando o documento que você vai enviar.
