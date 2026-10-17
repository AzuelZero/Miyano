# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

(Móvil-first; empaquetado nativo con Capacitor en un bloque posterior — el lenguaje de diseño sigue siendo web.)

## Users

El autor (Azuel), único usuario real de momento: futbolista que entrena en casa y gimnasio, con niveles de forma variables. Segundo público implícito: reclutadores/equipos técnicos que evalúan el repo como portfolio (repo público, escaparate verificable).

## Product Purpose

App personal de entrenamiento offline-first: catálogo de 1.324 ejercicios con GIFs de técnica, rutinas manuales con tags, auto-escala por nivel de forma (×0.70/1.00/1.30) y (fases futuras) calendario "¿qué toca hoy?", generador de rutinas y progresión adaptativa. El éxito tiene tres definiciones confirmadas por el autor: **uso personal real** (APK funcionando offline en su móvil), **aprendizaje** (proyecto-escuela para pasar de junior a senior) y **escaparate profesional** (código y proceso verificables en público).

## Positioning

Offline de verdad (el móvil es la fuente de verdad local; la sync llega cuando hay red) + rutinas que se auto-escalan al nivel sin duplicar variantes + progresión adaptativa planificada. Los vecinos (wger, openGym) no combinan offline-first radical con auto-escala por nivel.

## Operating Context

Uso diario en el gimnasio y en casa, a menudo SIN cobertura — offline es requisito, no nice-to-have. Sesiones cortas de consulta ("¿qué toca? ¿cuántas reps con mi nivel?"). Backend propio Go+PostgreSQL ya funcionando en local (auth JWT + catálogo + rutinas, PRs #7-#8).

## Capabilities and Constraints

- Backend completo y verificado: auth (argon2id+JWT+refresh rotatorio), catálogo protegido, rutinas con auto-escala y tiempo calculado.
- Frontend: este shell (login/registro/home) es el primer código; catálogo y rutinas vienen después (B2/B3).
- Constraint duro: todo asset autohospedado (fuentes incluidas) — sin CDNs, por offline y privacidad.
- Interfaz en español; código en inglés (ADR D009).
- Pendiente explícito (no inventar como hecho): sync engine offline, SQLite local, Capacitor/APK, calendario, avatar MinIO.

## Brand Commitments

- Nombre: **ミヤノ** (Miyano, katakana).
- Tipografía **Sora** (variable, autohospedada) — elegida por el autor en shootout 17/10/2026.
- Paleta **D013** dark-first: fondo `#080616`, superficie `#1A1953`, secundario `#162E93`, primario `#2F2FE4`, más `--color-primary-bright #7C7CFF` (nacido del cálculo WCAG 17/10).
- Tono: "una línea al día 💪🏼" — motivacional sin humo; nada de claims vacíos.

## Evidence on Hand

- Dataset real importado: 1.324 ejercicios + traducciones ES/EN (verificable en la API).
- Backend funcionando con CI verde y runbook reproducible (repo público AzuelZero/Miyano).
- Bitácora pública del proceso (ADR D001-D016, enciclopedia de lecciones).
- NO existen aún: capturas reales de la app, testimonios, métricas de usuarios — no fabricar.

## Product Principles

1. **Offline primero, siempre**: cualquier decisión que exija red para lo básico está mal.
2. **Los números mandan**: reps, tiempos y niveles se leen rápido y se alinean (tabular).
3. **Una sola verdad por cosa**: tokens en un fichero, estado del servidor en Query, sesión en Pinia, decisiones en la bitácora.
4. **Hecho > perfecto, pero verificado**: cada paso con su prueba (D011); el slop no pasa revisión.
5. **El proceso es parte del producto**: aprender y dejar rastro verificable.

## Accessibility & Inclusion

WCAG 2.1 AA como estándar (contraste calculado en frío, focus visible, `prefers-reduced-motion`, navegación por teclado). El usuario principal no tiene necesidades específicas declaradas — el estándar es el compromiso.
