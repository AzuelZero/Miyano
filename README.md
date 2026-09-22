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

# 2. Aplica las migraciones (esquema + seed de niveles de forma)
for f in backend/migrations/*.up.sql; do
  docker compose -f docker-compose.dev.yml exec -T postgres psql -U miyano -d miyano < "$f"
done

# 3. Carga el catálogo de ejercicios (1.324, ES/EN) — idempotente
#    descarga el dataset en el primer arranque (download-at-deploy, ver NOTICE.md)
cd backend
export DATABASE_URL=postgres://miyano:dev@localhost:5432/miyano
go run ./cmd/importer

# 4. Arranca la API en :8080 (JWT_SECRET: >= 32 caracteres, fail-fast)
export JWT_SECRET="dev-secret-change-me-32-chars-min!"
go run ./cmd/api
```

> **PowerShell:** el paso 2 equivale a
> `Get-Content backend/migrations/0001_init.up.sql | docker compose -f docker-compose.dev.yml exec -T postgres psql -U miyano -d miyano`
> y los `export` a `$env:DATABASE_URL = "postgres://miyano:dev@localhost:5432/miyano"` y `$env:JWT_SECRET = "dev-secret-change-me-32-chars-min!"`.

### 🩺 Healthcheck

**http://localhost:8080/api/v1/health** → `{"status":"ok"}`

### 📚 Endpoints actuales

Autenticación con JWT: `POST /api/v1/auth/register` devuelve el par de tokens; el catálogo requiere `Authorization: Bearer <access_token>`.

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/v1/health` | Estado del servicio |
| POST | `/api/v1/auth/register` | Registro (email + password ≥8 + display_name) → tokens |
| POST | `/api/v1/auth/login` | Login → access (15 min) + refresh (7 días) |
| POST | `/api/v1/auth/refresh` | Rotación de refresh token |
| GET | `/api/v1/exercises` | 🔒 Catálogo (filtro exacto `?equipment=body weight`) |
| GET | `/api/v1/exercises/{id}` | 🔒 Detalle con traducciones ES/EN |

Los endpoints `/auth/*` están limitados a 5 peticiones/minuto por IP. Las passwords se guardan hasheadas con argon2id (RFC 9106) y los refresh tokens persisten hasheados (SHA-256) para poder revocarlos y rotarlos.

## 🔧 Variables de entorno

| Variable | Obligatoria | Por defecto | Descripción |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | Cadena de conexión PostgreSQL (p. ej. `postgres://miyano:dev@localhost:5432/miyano`) |
| `JWT_SECRET` | ✅ | — | Secreto de firma HS256 (≥ 32 caracteres) |
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
