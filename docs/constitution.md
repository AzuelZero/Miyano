# Miyano — Constitución del proyecto

> Principios **no negociables** que gobiernan cada spec, cada PR y cada línea de código.
> Condensado de los ADRs D001–D015 (razonamiento completo en `Miyano-Docs`).
> Cualquier cambio a esta constitución = nuevo ADR que lo justifique.

1. **El flujo de trabajo es SDD con OpenSpec**: propose → (spec/design/tasks) → apply → archive. Nada de código sin change aprobado. Ceremonia proporcional: cambios triviales (<30 min, sin dinero/datos/seguridad) van directos con Conventional Commits.
2. **Git**: `main` protegida (PR + CI obligatorio). Ramas `feature/*` efímeras. Cada PR lleva revisión semáforo simulada (🟢🟡🔴❓) ANTES del merge, con veredicto.
3. **El código va SIEMPRE en inglés** y con la matriz de cases estándar: camelCase locales, PascalCase exportados (en Go es semántica, no estilo), snake_case SQL/ficheros Go, SCREAMING solo env vars. Comentarios mínimos senior-style: la pedagogía vive en la enciclopedia de Miyano-Docs, no en el repo.
4. **Arquitectura hexagonal**: `domain` puro (sin I/O) ← `ports` (interfaces) ← `adapter` (infra) ← `service` ← `transport`. Los handlers nunca hablan con la BD directamente; las dependencias apuntan hacia dentro.
5. **Verificabilidad (la regla de las reglas)**: ningún checkbox se marca sin ver el output esperado. Cada doc mantiene su runbook "✅ Cómo verificar" actualizado. *Si no sé cómo verificarlo, no he terminado.*
6. **Fail fast**: la config se valida al arranque y muere con mensajes claros (nunca arrancar "bien" para fallar luego). Sin secretos en código: `.env` ignorado, secretos en el hosting.
7. **Seguridad por hábito**: argon2id (RFC 9106) + `ConstantTimeCompare`; JWT con `WithValidMethods` + exp requerida; timeouts siempre; `permissions` mínimas en CI; auditorías periódicas registradas (checklist en Miyano-Docs).
8. **Specs vivas**: si el código cambia, la spec cambia en el mismo PR. Specs mentirosas son peores que no tener. Las scenarios son WHEN/THEN (EARS) verificables.
9. **Datos**: el catálogo (dataset hasaneyldrm) es texto MIT importable; las imágenes/GIFs son © Gym Visual → **nunca commiteadas** (download-at-deploy). Atribución en `NOTICE.md` siempre visible.
10. **Docs gemelas**: código y decisiones viven aquí; el aprendizaje sobreexplicado (enciclopedia, bitácora, ADRs) vive en Miyano-Docs. Ningún conocimiento crítico vive solo en un chat.
