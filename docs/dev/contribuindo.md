# Guia de Contribuição

Obrigado por querer contribuir com o Astrolink! Este guia explica como você pode ajudar.

---

## Como posso contribuir?

### Reportar bugs
- Abra uma issue no GitHub com o template de bug report
- Inclua: versão do Astrolink, SO, logs relevantes, passos para reproduzir
- Se possível, inclua screenshot ou video

### Sugerir features
- Abra uma issue com o template de feature request
- Explique o problema que a feature resolve (não apenas "o que" mas "por quê")
- Mencione se você está disposto a implementar

### Contribuir código
- Veja a seção "Fluxo de Contribuição" abaixo
- Procure issues com label `good first issue` para começar

### Melhorar documentação
- PRs de documentação são muito bem-vindos
- Traduções para outros idiomas (após v1.0)

### Criar temas para o marketplace
- Veja `docs/specs/marketplace.md` para as diretrizes de temas
- Submeta via marketplace após o lançamento

### Testar e dar feedback
- Instale a versão alpha/beta e reporte problemas
- Feedback de UX (o que foi confuso, o que faltou) é ouro

---

## Código de Conduta

Este projeto adota o [Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). Em resumo:

- ✅ Seja respeitoso e construtivo
- ✅ Aceite feedback graciosamente
- ✅ Foque no problema, não na pessoa
- ❌ Sem assédio, discriminação ou linguagem ofensiva
- ❌ Sem spam ou autopromoção excessiva

Violações: enviar email para conduct@astrolink.app.

---

## Fluxo de Contribuição

### 1. Fork e Clone

```bash
# Fork no GitHub, depois:
git clone https://github.com/SEU_USUARIO/astrolink.git
cd astrolink

# Adicionar remote upstream
git remote add upstream https://github.com/astrolink/astrolink.git
```

### 2. Criar Branch

```bash
# Sempre criar branch a partir de main atualizada
git checkout main
git pull upstream main
git checkout -b feat/minha-feature
# ou
git checkout -b fix/descricao-do-bug
```

**Convenção de nomes de branch:**
- `feat/` — nova feature
- `fix/` — correção de bug
- `docs/` — documentação
- `refactor/` — refatoração sem mudança de comportamento
- `test/` — adição de testes
- `chore/` — manutenção (deps, CI, etc.)

### 3. Desenvolver

```bash
# Seguir o guia de setup: docs/dev/setup-local.md
make dev

# Manter seu branch atualizado com main
git fetch upstream
git rebase upstream/main
```

### 4. Commits

Seguimos **Conventional Commits**:

```
<tipo>(<escopo>): <descrição curta>

[corpo opcional — explique o POR QUÊ]

[rodapé opcional — Closes #123, Breaking Change: ...]
```

**Tipos válidos:**
- `feat` — nova funcionalidade (MINOR no semver)
- `fix` — correção de bug (PATCH no semver)
- `docs` — documentação
- `style` — formatação (sem mudança de lógica)
- `refactor` — refatoração
- `test` — adição/correção de testes
- `chore` — manutenção
- `perf` — melhoria de performance
- `ci` — mudanças no CI/CD

**Escopos comuns:**
`portal`, `admin`, `api`, `db`, `nds`, `auth`, `payments`, `vouchers`, `cloud`, `cli`

**Exemplos:**
```
feat(portal): adicionar suporte a planos por dados (GB)
fix(api): corrigir expiração incorreta de vouchers universais
docs(setup): atualizar guia de instalação para Ubuntu 24.04
refactor(nds): extrair cliente SSH para package separado
test(vouchers): adicionar testes de resgates simultâneos
```

**Commits atômicos:** cada commit deve representar uma mudança lógica completa. Não misture features e bug fixes no mesmo commit.

### 5. Testes

```bash
# Rodar todos os testes antes de abrir PR
make test

# Cobertura mínima: 70% no core
make test-coverage

# Lint (obrigatório passar)
make lint
```

### 6. Pull Request

```bash
git push origin feat/minha-feature
```

Abrir PR no GitHub:
- Título: seguir Conventional Commits
- Descrição: preencher o template
- Linkar issues relacionadas (`Closes #123`)
- Adicionar screenshots/videos se for mudança visual
- Marcar para review: `@astrolink/core-team`

**Template de PR:**
```markdown
## Descrição
O que esta PR faz? Por quê?

## Tipo de mudança
- [ ] Bug fix
- [ ] Nova feature
- [ ] Breaking change
- [ ] Documentação

## Testes
- [ ] Testes unitários adicionados/atualizados
- [ ] Testes de integração passando
- [ ] Testado manualmente (descreva abaixo)

## Screenshots (se mudança visual)

## Checklist
- [ ] Meu código segue o style guide do projeto
- [ ] Rodei `make lint` sem erros
- [ ] Rodei `make test` sem falhas
- [ ] Atualizei a documentação se necessário
- [ ] Adicionei entrada no CHANGELOG
```

---

## Processo de Review

1. **Automated checks:** CI roda lint, testes, security scan
2. **Peer review:** ao menos 1 aprovação de core team
3. **Merge:** squash merge para features, merge commit para releases

**Tempo médio de review:** 2–5 dias úteis.

**Feedback:** todo feedback é construtivo. Se discordar, argumente tecnicamente.

---

## Primeiro PR? Tente estas issues

Procure issues com labels:
- `good first issue` — simples, bem definidas, mentoria disponível
- `help wanted` — precisamos de ajuda, complexidade variada
- `documentation` — melhorias na documentação

---

## Releases e Versionamento

Seguimos **Semantic Versioning (semver):**

- `MAJOR.MINOR.PATCH`
- PATCH: bug fixes
- MINOR: novas features retrocompatíveis
- MAJOR: breaking changes

**Ciclo de releases:**
- Patches: conforme necessário
- Minor: a cada 4–6 semanas
- Major: com aviso de 3+ meses de antecedência

---

## Estrutura da Core Team

- **Mantenedores:** review final, merge, releases
- **Contribuidores regulares:** review de PRs, triage de issues
- **Contribuidores:** PRs aceitos

Contribuidores com 5+ PRs aceitos são convidados a se tornar contribuidores regulares.

---

## Dúvidas?

- GitHub Discussions: perguntas técnicas
- Discord: chat em tempo real (link no README)
- Email: dev@astrolink.app (questões sensíveis)
