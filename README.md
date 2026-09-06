<div align="center">

# ミヤノ · Miyano

[![CI](https://github.com/AzuelZero/Miyano/actions/workflows/ci.yml/badge.svg)](https://github.com/AzuelZero/Miyano/actions/workflows/ci.yml)

**App personal de rutinas de entrenamiento, con énfasis en entrenamiento sin equipo o con cosas de casa.**

*Sin más, a empezar aunque sea una línea al día 💪🏼*

![LET'S GO!!!](https://i.redd.it/w3dnnpapbdx71.jpg)

</div>

---

## ✨ Qué es

Miyano (ミヤノ) es una aplicación móvil **offline-first** para crear y seguir rutinas de ejercicio en cualquier lugar y momento, con o sin equipo. Nace como proyecto personal para **aprender de verdad** tecnologías nuevas (foco en **Go**) mientras construyo algo útil para mí mismo.

> **Nombre:** *Miyano* (ミヤノ).

El backend ya está operativo: API REST con catálogo de **1.324 ejercicios** en español e inglés, arquitectura hexagonal, CI en verde y tests.

## 🧰 Stack

| Capa | Tecnología |
|---|---|
| **Backend** | Go · Chi · sqlc · pgx · argon2id · JWT (auth en camino) |
| **BD** | PostgreSQL · MongoDB · SQLite (local) · MinIO (avatares) |
| **Frontend** | Vue 3 · TypeScript · Vite · Tailwind v4 · Pinia · Capacitor |
| **Infra** | Docker · GitHub Actions · Coolify + VPS (planeado) |

Arquitectura: **offline-first** (el móvil funciona sin red; sincroniza cuando hay conexión).

## 🚀 Quick start

**Requisitos:** Go 1.26+ · Docker

```bash
# 1. Levanta las bases de datos de desarrollo
#    (hoy solo se usa PostgreSQL; Mongo y MinIO quedan definidos para fases futuras)
docker compose -f docker-compose.dev.yml up -d

# 2. Aplica la migración inicial (tablas + seed de niveles de forma)
docker compose -f docker-compose.dev.yml exec -T postgres psql -U miyano -d miyano < backend/migrations/0001_init.up.sql

# 3. Carga el catálogo de ejercicios (1.324, ES/EN) — idempotente
#    descarga el dataset en el primer arranque (download-at-deploy, ver NOTICE.md)
cd backend
export DATABASE_URL=postgres://miyano:dev@localhost:5432/miyano
go run ./cmd/importer

# 4. Arranca la API en :8080
go run ./cmd/api
```

> **PowerShell:** el paso 2 equivale a
> `Get-Content backend/migrations/0001_init.up.sql | docker compose -f docker-compose.dev.yml exec -T postgres psql -U miyano -d miyano`
> y el `export` del paso 3 a `$env:DATABASE_URL = "postgres://miyano:dev@localhost:5432/miyano"`.

### 🩺 Healthcheck

**http://localhost:8080/api/v1/health** → `{"status":"ok"}`

### 📚 Endpoints actuales

Públicos provisionalmente: se protegen con la Lección de Auth (JWT).

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/v1/health` | Estado del servicio |
| GET | `/api/v1/exercises` | Catálogo (filtro exacto `?equipment=body weight`) |
| GET | `/api/v1/exercises/{id}` | Detalle con traducciones ES/EN |

## 🔧 Variables de entorno

| Variable | Obligatoria | Por defecto | Descripción |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | Cadena de conexión PostgreSQL (p. ej. `postgres://miyano:dev@localhost:5432/miyano`) |
| `PORT` | — | `8080` | Puerto HTTP de la API |

## 🧪 Tests

```bash
cd backend
go test ./...           # tests unitarios y de integración
```

## 🏗️ Arquitectura

Ver documentación técnica en `docs/` (y en el repositorio de apuntes del autor).

## 🤝 Cómo contribuir

Proyecto personal, no se aceptan contribuciones externas de momento.

## 📄 Licencia y atribuciones

- **Código propio:** MIT (ver [`LICENSE`](./LICENSE)).
- **Dataset de ejercicios:** de [`hasaneyldrm/exercises-dataset`](https://github.com/hasaneyldrm/exercises-dataset) (MIT para texto/estructura).
- ⚠️ **Las imágenes y GIFs del dataset son © Gym Visual**, redistribuidas con permiso **solo para uso personal y no comercial**. Ver [`NOTICE.md`](./NOTICE.md).

## 💡 Inspiración

La idea del generador de rutinas (fase 2) está inspirada en [workout.lol](https://workout.lol) (MIT, de @Vincenius), cuyo patrón *equipo → músculos → rutina* estudiamos como referencia.
