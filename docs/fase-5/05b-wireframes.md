# Fase 5 — Wireframes em texto

> Largura de referência: celular (360–390 px). Textos em pt-BR são **exemplos de chave `t()`**, não texto final. `[ ]` = botão; `( )` = opção; `▓` = barra preenchida; `◉` = selecionado/sugerido. Alvos de toque ≥ 48 px. O app tem barra de navegação inferior com 5 itens; no desktop vira menu lateral.

## W1 — Início (`/`)

```
┌──────────────────────────────────┐
│ wordloop                    ⚙ ☰ │
├──────────────────────────────────┤
│ Boa tarde, Keven                 │
│ ┌──────────────────────────────┐ │
│ │ Hoje: 12 de 20 min     ◔ 60% │ │   ← meta do dia (ciclo atual)
│ │ ▓▓▓▓▓▓▓░░░░                  │ │
│ │ [        Continuar         ] │ │   ← botão "3D" primário
│ └──────────────────────────────┘ │
│  Sequência        Acertos (7 d)  │
│  🔥 6 dias          82% ▲ +4     │   ← setas = melhora vs 7 d anteriores
│                                  │
│ Sua semana                       │
│  S  T  Q  Q  S  S  D             │   ← mapa de calor semanal (links p/ 12 sem.)
│  ■  ■  ■  □  ■  ·  ·             │   · = dia fora da meta (não quebra sequência)
│                                  │
│ Você conhece 312 palavras        │
│ ▓▓▓▓░░░░░░  ≈ 74% de um texto    │   ← cobertura estimada
│ Próximo texto liberado: 91% ──►  │
│                                  │
│ Revisões previstas: amanhã 18 · 24h │
│ Mais erradas: sort · between …   │
├──────────────────────────────────┤
│ Início  Estudar  Ler  Palavras Mais │
└──────────────────────────────────┘
```

## W2 — "Quanto tempo hoje?" (sheet sobre Estudar)

```
┌──────────────────────────────────┐
│ Quanto tempo você tem hoje?   ✕  │
│                                  │
│ ( 5 min )  (◉15 min)  ( 30 )  ( 60 ) │   ← pré-seleciona o da meta
│                                  │
│ O que estudar?                   │
│ (◉ Misturar)  ( Revisar )  ( Novas ) │
│                                  │
│ ~60 exercícios · 8 palavras novas│
│ 22 revisões vencidas             │
│                                  │
│ [          Começar              ]│
└──────────────────────────────────┘
```

## W3 — Palavra nova (apresentação, sem nota)

```
┌──────────────────────────────────┐
│ ✕  ▓▓░░░░░░░░░░░░░   3 / 60      │
├──────────────────────────────────┤
│        Palavra nova              │
│                                  │
│          ┌────────┐              │
│          │  🐱    │              │   ← OpenMoji / foto (concretas)
│          └────────┘              │
│            cat   [🔊] [🐢]       │   ← ouvir / ouvir devagar
│            gato                  │
│  "The cat is on the table."      │
│   O gato está sobre a mesa.      │
│                                  │
│ [         Entendi               ]│   → logo vem o teste de reconhecimento
└──────────────────────────────────┘
```

## W4 — Exercício de reconhecimento (EN → PT, múltipla escolha)

```
┌──────────────────────────────────┐
│ ✕  ▓▓▓▓░░░░░░░░░░░   4 / 60     │
├──────────────────────────────────┤
│ O que significa?                 │
│                                  │
│          sort    [🔊]            │
│                                  │
│ ┌──────────────────────────────┐ │
│ │ A  gato                      │ │
│ └──────────────────────────────┘ │
│ ┌──────────────────────────────┐ │
│ │ B  tipo / classificar        │ │
│ └──────────────────────────────┘ │
│ ┌──────────────────────────────┐ │
│ │ C  entre                     │ │
│ └──────────────────────────────┘ │
│ ┌──────────────────────────────┐ │
│ │ D  quase                     │ │   ← distratores da mesma faixa, sem irmãos semânticos
│ └──────────────────────────────┘ │
│ [ Não sei ]                      │   ← conta como erro, sem digitar
└──────────────────────────────────┘
```

## W5 — Exercício de produção (PT → EN, digitar) e feedback de acerto

```
Antes de responder                       Depois de acertar
┌───────────────────────────┐            ┌───────────────────────────┐
│ ✕  ▓▓▓▓▓░░░░░░░   9 / 60 │            │ ✕  ▓▓▓▓▓░░░░░░░   9 / 60 │
│ Escreva em inglês         │            │ ✔ Correto!                │   ← verde + ícone + texto
│ "o gato"                  │            │ the cat                   │
│ ┌───────────────────────┐ │            │                           │
│ │ the c|                │ │            │ Foi fácil, médio ou       │
│ └───────────────────────┘ │            │ difícil lembrar?          │
│ [Dica: 1ª letra] [Não sei]│            │ ┌────────┐┌────────┐┌────────┐
│ [        Verificar      ] │            │ │ Difícil││◉ Médio ││ Fácil │
└───────────────────────────┘            │ └────────┘└────────┘└───────┘
 Dica usada → nota máxima "Difícil"      │   ◉ = sugestão pelo tempo   │
 Teclado: autocapitalize/autocorrect OFF │ Tocar escolhe e avança; Enter aceita a sugestão
                                         └───────────────────────────┘
```

## W6 — Erro e "como foi?"

```
┌──────────────────────────────────┐
│ ✕  ▓▓▓▓▓▓░░░░░░   10 / 60       │
│ ✖ Quase! (ícone + texto, não só cor) │
│ Sua resposta:  the cet           │
│ Correto:       the cat           │   ← diferença destacada
│                                  │
│ Como foi?                        │
│ ┌────────┐ ┌────────┐ ┌────────┐ │
│ │ Não    │ │ Quase  │ │ Foi    │ │
│ │ sabia  │ │ lembrei│ │ distração│ │
│ └────────┘ └────────┘ └────────┘ │
│                                  │
│ ▸ Entender por quê               │   ← abre: imagem de contexto, frase, dica
│   ┌────┐ "The cat sleeps."       │
│   │ 🐱 │ O gato dorme.           │
│   └────┘                         │
│ [ Continuar ]                    │   ← o item volta em 3 cartões (Fase 1)
└──────────────────────────────────┘
```

## W7 — Resumo da sessão

```
┌──────────────────────────────────┐
│ Sessão concluída                 │
│ 15 min · 58 exercícios           │
│                                  │
│ Acerto: 84%     Novas: 8         │
│ 🔥 Sequência: 7 dias             │
│                                  │
│ O que melhorou                   │
│  +8 palavras conhecidas          │
│  Tempo de resposta ▼ 1,2 s       │
│  2 palavras difíceis ficaram mais fáceis │
│                                  │
│ Meta de hoje: ✔ concluída        │
│ Amanhã: 18 revisões previstas    │
│                                  │
│ [  Mais 5 minutos  ] [ Concluir ]│
└──────────────────────────────────┘
```

## W8 — Metas: ciclo atual e retrospectiva da semana

```
Ciclo atual (/goals)                    Fim do ciclo (retrospectiva)
┌───────────────────────────┐           ┌───────────────────────────┐
│ Meta da semana            │           │ Sua semana                │
│ 20 min/dia · 5 dias       │           │ 4 de 5 dias · 78 min      │
│ S T Q Q S S D             │           │ +21 palavras conhecidas   │
│ ✔ ✔ ✔ ○ ○ · ·             │           │ Acerto 79% → 84%  ▲       │
│ ◔ 3 de 5 dias             │           │                           │
│ Faltam 2 dias até domingo │           │ Qual a meta da próxima?   │
│ [ Ajustar meta ]          │           │ (Leve  15 min · 3 dias)   │
└───────────────────────────┘           │ (◉Constante 20 min · 5)   │ ← 0,8× · 1× · p75
                                        │ (Intensa 30 min · 6)      │   do seu tempo real
                                        │ ( Personalizar )          │
                                        │ [ Começar semana ] [Manter a mesma] │
                                        └───────────────────────────┘
```

## W9 — Lembretes: pedido de permissão (depois da 1ª sessão)

```
Android / desktop                       iPhone (PWA ainda não instalado)
┌───────────────────────────┐           ┌───────────────────────────┐
│ Quer receber desafios?    │           │ Para receber avisos no    │
│ Uma pergunta rápida no    │           │ iPhone, instale o app:    │
│ seu horário. Você controla│           │ 1 Toque em  ⎙ Compartilhar│
│ quando e quantos.         │           │ 2 "Adicionar à Tela de    │
│                           │           │    Início"                │
│ Horário:  09:00 – 21:00   │           │ 3 Abra o wordloop por lá  │
│ Máx. por dia: 3           │           │ [ Já instalei ] [Depois]  │
│ [ Ativar avisos ]         │           └───────────────────────────┘
│ [ Agora não ]             │            Só depois de instalado o botão
└───────────────────────────┘            "Ativar avisos" (toque do usuário) pede a permissão.
```

## W10 — Ler (texto graduado)

```
┌──────────────────────────────────┐
│ ←  Texto 3 · A casa do Sam  [🔊] │
│ Você conhece 96% deste texto     │   ← ≥95% = leitura guiada
│ ▓▓▓▓▓▓▓▓▓▓░  5 palavras novas    │
├──────────────────────────────────┤
│ Sam has a small house. It is     │
│ near a ▒park▒. He has a ▒garden▒ │   ← ▒ = desconhecida (toque abre glossário)
│ with many flowers.               │
│                                  │
│   ┌ park · parque ───────────┐   │
│   │ "We walk in the park."   │   │
│   │ [ Adicionar aos estudos ]│   │
│   └──────────────────────────┘   │
│ [ Mostrar tradução ]             │   ← oculta por padrão
│ [ Terminei: 5 palavras novas → ] │
└──────────────────────────────────┘
Bloqueado (<95%): "Você conhece 82% do texto mais fácil. Faltam 14 palavras." [ Estudar essas palavras ]
```

## W11 — Nivelamento (Sim/Não)

```
┌──────────────────────────────────┐
│ Marque as palavras que você      │
│ conhece                          │
│ Algumas são inventadas — não     │
│ marque essas!                    │
│ ( ) house   ( ) table  (◉) blorp │
│ ( ) go      ( ) because ( ) tamen│
│ ( ) between ( ) sort   ( ) lonse │
│                                  │
│ Etapa 2 de ~4        ▓▓▓░░░░░    │
│ [ Próxima ]   [ Não conheço nenhuma ]│
└──────────────────────────────────┘
```

## W12 — Palavras (biblioteca)

```
┌──────────────────────────────────┐
│ Palavras                  🔍     │
│ [Todas][Revisar][Mais erradas][Dominadas] │
│ ┌──────────────────────────────┐ │
│ │ sort  (v.)  classificar   ◔ 44% │   ← "domínio" (Fase 1), não a probabilidade R
│ │ próxima revisão: amanhã      │ │
│ └──────────────────────────────┘ │
│ ┌──────────────────────────────┐ │
│ │ between (prep.) entre     ◕ 66% │
│ └──────────────────────────────┘ │
└──────────────────────────────────┘
Detalhe: termo, sentidos, exemplos, imagem, histórico das últimas respostas,
relações (só quando liberadas), [ Reportar problema neste item ].
```

## W13 — Configurações › Lembretes

```
┌──────────────────────────────────┐
│ ← Lembretes                      │
│ Avisos        [━━◉] ligado       │
│ Janela        09:00 — 21:00      │
│ Dias          S T Q Q S S D      │
│ Máx. por dia  ( 1 )(2)(◉3)(5)    │
│ Silêncio      22:00 — 08:00      │
│ Esconder a pergunta na tela de bloqueio [━━○]│
│ Dispositivos                     │
│  • Pixel 8 (este)    [ Remover ] │
│  • iPhone            [ Remover ] │
│ [ Enviar notificação de teste ]  │
└──────────────────────────────────┘
```

## Estados transversais

```
Offline:      ┌ ◌ Sem conexão. Suas respostas ficam salvas e são enviadas depois (7 pendentes) ┐
Atualização:  ┌ Nova versão disponível.  [ Atualizar ] ┐   (aviso do service worker; nunca recarrega no meio da sessão)
Sessão expirada: modal "Entre de novo para continuar" — a fila de respostas não se perde
Vazio:        ilustração simples + 1 frase + 1 botão  (ex.: "Nada para revisar agora. Que tal aprender palavras novas?")
Sem voz em inglês: o exercício "ouvir" some e a tela de configurações explica por quê
```
