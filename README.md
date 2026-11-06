# vagas-radar

Programa de linha de comando em Go que vigia sites de vagas remotas, filtra pelo
que me interessa (Python, Go, Luau/Roblox, IA, estágio) e avisa **só das vagas
novas**, no terminal ou num canal do Discord.

Fiz porque passei horas abrindo site por site à procura de vaga e freelance. Agora
o computador faz essa parte e eu só leio o que importa.

## Como funciona

1. Busca em todas as fontes **ao mesmo tempo** (uma goroutine por fonte, com tempo
   limite). Se uma fonte cair, as outras seguem e o erro aparece no log.
2. Filtra pelo título e pelas tags. Os termos casam como palavra inteira: `go` casa
   com "Go Engineer", mas não com "Django".
3. Descarta o que já avisei antes, guardando as chaves num SQLite.
4. Manda as novas para o terminal ou para o Discord, em blocos que respeitam o limite
   de 2000 caracteres da mensagem.
5. Só **depois** de o aviso sair é que marca as vagas como vistas. Se a rede falhar, a
   vaga não se perde: sai no próximo ciclo.

## Fontes

| Fonte | Observação |
|---|---|
| [Remotive](https://remotive.com) | API pública de vagas remotas |
| [Remote OK](https://remoteok.com) | API pública de vagas remotas |

Os dois serviços pedem crédito, e por isso cada aviso termina com `(via Fonte)` e o
link da vaga original. Elas trazem poucas vagas em português e quase nenhuma de
estágio no Brasil, então este radar serve melhor para freelance e vaga remota
internacional.

Para somar uma fonte nova, implemente a interface `source.Source` (dois métodos:
`Name` e `Fetch`).

## Como rodar

Precisa de Go 1.27 (a versão do `go.mod`).

```bash
go run ./cmd/radar -dry-run          # só mostra, não marca nada como visto
go run ./cmd/radar                   # roda uma vez
go run ./cmd/radar -interval 30m     # repete a cada 30 minutos
```

Para receber no Discord, crie um webhook no canal (Configurações do canal >
Integrações) e passe por variável de ambiente:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
```

O webhook vem só de variável de ambiente, nunca de flag, porque flags aparecem na
lista de processos.

| Flag | Padrão | Para que serve |
|---|---|---|
| `-include` | python,go,golang,lua,luau,roblox,ai,llm,... | termos que interessam |
| `-exclude` | senior,sr,staff,lead,manager,... | descarta pelo título |
| `-max` | 15 | máximo de vagas por aviso (o resto sai no próximo ciclo) |
| `-interval` | 0 | repetir a cada tanto tempo; 0 roda uma vez |
| `-db` | radar.db | arquivo do SQLite |
| `-dry-run` | desligado | mostra no terminal e não marca como visto |

## Testes

```bash
go test ./... -cover
```

Cobertura: filtro 92%, fontes 89%, avisos 89%, banco 82%. As fontes são testadas
contra servidores HTTP falsos (`httptest`), sem depender da internet. O `main.go`
não tem teste.

## Limitações

- O filtro olha título e tags, não a descrição. Algumas vagas passam sem serem
  relevantes (a palavra `ai` é genérica) e outras boas podem ficar de fora.
- Só vagas remotas internacionais. 99Freelas e Workana não têm API pública e eu não
  faço scraping de página com login.
- Não tenta deduplicar a mesma vaga publicada em duas fontes.
