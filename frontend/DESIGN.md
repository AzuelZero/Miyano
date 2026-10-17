---
name: Miyano
description: Offline-first workout app — dark, precise, measured.
colors:
  background: "#080616"
  surface: "#1A1953"
  secondary: "#162E93"
  primary: "#2F2FE4"
  primary-bright: "#7C7CFF"
  text: "#E6E8F2"
  muted: "#9A9FB8"
  danger: "#FF7B7B"
typography:
  display:
    fontFamily: "Sora Variable, system-ui, sans-serif"
    fontWeight: 700
    letterSpacing: "0.08em"
  body:
    fontFamily: "Sora Variable, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.55
  label:
    fontFamily: "Sora Variable, system-ui, sans-serif"
    fontSize: "0.72rem"
    fontWeight: 600
    letterSpacing: "0.08em"
rounded:
  md: "10px"
  lg: "14px"
  pill: "999px"
spacing:
  sm: "8px"
  md: "16px"
  lg: "24px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "#FFFFFF"
    rounded: "{rounded.md}"
    padding: "12px 16px"
  button-primary-hover:
    backgroundColor: "{colors.primary-bright}"
  input-field:
    backgroundColor: "{colors.background}"
    textColor: "{colors.text}"
    rounded: "{rounded.md}"
    padding: "10px 14px"
  card-surface:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.lg}"
    padding: "24px"
---

# Design System: Miyano

## Overview

**Creative North Star: "El laboratorio de Haibara"**

*(Elegida por el autor, 17/10/2026.)*

Miyano es el laboratorio nocturno de una científica: frío y preciso por fuera, cálido por dentro, y TODO se mide. La superficie es azul-indigo profundo casi negro; la jerarquía no se grita, se calibra; y los números —reps, segundos, factores— son los verdaderos protagonistas tipográficos. La marca vive en detalles exactos, no en decoración: un focus ring violeta claro, dígitos tabulares alineados, un momento autoral por pantalla y nada más.

Los componentes se sienten **táctiles y confiados** *(elección del autor)*: rellenos sólidos, bordes definidos, feedback inmediato — la app debe sentirse sólida bajo dedos sudados de gimnasio. Densidad media-baja: pocas cosas por pantalla, cada una con aire. Es una app de consulta rápida con dedos y sin cobertura — scanability antes que expresión (modo Operate).

**Key Characteristics:**
- Dark-first calibrado: tres capas tonales (fondo → superficie → secundario), sin sombras decorativas
- Un solo acento con dos registros: primario para RELLENAR, primary-bright para SEÑALAR (focus, links, activo)
- Sora en todo (display=sans, una sola voz), dígitos SIEMPRE tabulares
- Un sistema de radius (10/14/pill) y ya
- Copys en español natural que nombran la acción y la recuperación

## Colors

Paleta D013 (ADR del 04/09/2026) + `primary-bright` nacido del cálculo WCAG del 17/10/2026. Contrastes verificados en frío (texto ≥4.5:1, UI ≥3:1).

### Primary
- **Azul Haibara** (#2F2FE4): acciones — SOLO como relleno (texto blanco encima: 7.93:1). Botones primarios, selección de texto, estados activos llenos.
- **Violeta Reactivo** (#7C7CFF): el acento UI sobre oscuro — focus rings, links, indicadores activos (5.90:1 sobre fondo, 4.73:1 sobre superficie). Nació porque el azul puro solo alcanza 2.53:1 como UI (WCAG 1.4.11).

### Secondary
- **Azul Profundo** (#162E93): rellenos interactivos secundarios, badges, scrollbar. Nunca texto sobre él salvo blanco.

### Neutral
- **Noche de Laboratorio** (#080616): el fondo de todo. Casi negro, inequívocamente indigo.
- **Superficie de Trabajo** (#1A1953): cards y paneles elevados (13.1:1 con texto principal).
- **Pizarra Clara** (#E6E8F2): texto principal (16.4:1 sobre fondo).
- **Gris de Nota** (#9A9FB8): texto secundario/labels (7.66:1 — pasa AA, no es decorativo).
- **Rojo de Alerta** (#FF7B7B): errores, siempre con texto que explica la recuperación.

### Named Rules
**La Regla del Relleno y la Señal.** El primario rellena (botones, selección); el primary-bright señala (focus, links, activo). Nunca al revés: el primario puro sobre fondo oscuro NO alcanza contraste UI (2.53:1) — esa es la razón de su existencia.

**La Regla del 10%.** El acento (bright) aparece en ≤10% de cualquier pantalla. Su rareza es lo que lo hace legible.

## Typography

**Display Font:** Sora Variable (fallback system-ui)
**Body Font:** Sora Variable (fallback system-ui)

**Character:** Una sola voz — geométrica, técnica, con dígitos de personalidad. La jerarquía se construye con peso y tamaño, no con fuentes distintas.

### Hierarchy
- **Display** (700, tracking 0.08em): wordmark ミヤノ y títulos de pantalla.
- **Title** (600, 1.25rem): nombres de rutina, secciones.
- **Body** (400, 1rem, line-height 1.55): todo lo demás. Measure 65-75ch en textos largos.
- **Label** (600, 0.72rem, uppercase, tracking 0.08em): labels de formulario y metadatos.
- **Números** (600, tabular SIEMPRE): 3×12 · 60s · R90 — `font-variant-numeric: tabular-nums`.

### Named Rules
**La Regla del Dígito.** Todo dato numérico del dominio (reps, sets, segundos, factores) se renderiza con la utility `tabular-nums`. Una columna que baila es un bug visual.

## Layout

Móvil-first real: max-width de contenido ~640px centrado en desktop. AppShell con topbar mínima (wordmark + logout) y nav inferior móvil. Espaciado a saltos de 8px; más aire ARRIBA de un heading que debajo. Grupos tight, separación generosa.

## Elevation & Depth

**Apilado tonal, no sombras.** La profundidad la dan las tres capas de color (fondo → superficie → secundario). Una sombra suave MAY appear only as state response (focus glow) — jamás como decoración (zero-offset halos prohibidos por craft-floor).

### Named Rules
**La Regla Tonal.** Profundidad = capa de color, no blur. Si algo necesita parecer "encima", sube su capa tonal; no le pongas shadow.

## Shapes

Esquinas generosas y consistentes: 10px en controles (inputs, botones), 14px en contenedores (cards), pill (999px) solo en badges. Sin bordes decorativos >1px, sin glassmorphism, sin gradientes.

## Components

### Buttons
- **Shape:** 10px; padding 12px 16px; peso 600.
- **Primary:** relleno Azul Haibara, texto blanco; hover → Violeta Reactivo; focus-visible → ring 2px Violeta Reactivo offset 2px.
- **Ghost:** transparente, texto Pizarra Clara, borde 1px rgba(blanco .14); para acciones secundarias.
- **Loading:** spinner + texto deshabilitado — JAMÁS desaparecer el label.

### Cards / Containers
- **Corner:** 14px · **Background:** Superficie de Trabajo · **Borde:** rgba(blanco .06) · **Sin sombra** (Regla Tonal) · **Padding:** 24px.

### Inputs / Fields
- **Style:** fondo Noche de Laboratorio (inverso a la card), borde 1px rgba(blanco .14), 10px.
- **Focus:** ring 2px Violeta Reactivo; **Error:** borde Rojo de Alerta + mensaje que nombra la recuperación debajo.
- **Label:** Label-style (uppercase 0.72rem) SIEMPRE visible encima — placeholder no es label.

### Navigation
- **Topbar:** wordmark ミヤノ a la izquierda, acciones de sesión a la derecha. Fondo = fondo (sin borde salvo 1px al hacer scroll).
- **Nav inferior (móvil):** iconos Lucide 20px + label 0.72rem; activo = Violeta Reactivo.

## Do's and Don'ts

### Do:
- **Do** usar `tabular-nums` en todo número del dominio (Regla del Dígito).
- **Do** nombrar la recuperación en cada error ("Espera un minuto e inténtalo de nuevo" para 429).
- **Do** themear las superficies del navegador: selection, caret, scrollbar, focus (ya en `main.css`).
- **Do** un solo momento autoral por pantalla, y respetar `prefers-reduced-motion`.

### Don't:
- **Don't** usar el primario #2F2FE4 como texto/anillo sobre fondo oscuro (2.53:1 — para eso existe primary-bright).
- **Don't** añadir sombras, glass, gradientes ni border-left decorativos (Regla Tonal + craft-floor).
- **Don't** usar emoji o unicode como icono — Lucide o nada.
- **Don't** mostrar un número del dominio sin `tabular-nums`.
- **Don't** inventar testimonios, métricas ni claims — PRODUCT.md manda.
