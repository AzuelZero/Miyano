<div align="center">

# ミヤノ · Miyano

**App personal de rutinas de entrenamiento, con énfasis en entrenamiento sin equipo o con cosas de casa.**

*Sin más, a empezar aunque sea una línea al día 💪🏼*

![LET'S GO!!!](https://i.redd.it/w3dnnpapbdx71.jpg)

</div>

---

## ✨ Qué es

Miyano (ミヤノ) es una aplicación móvil **offline-first** para crear y seguir rutinas de ejercicio en cualquier lugar y momento, con o sin equipo. Nace como proyecto personal para **aprender de verdad** tecnologías nuevas (foco en **Go**) mientras construyo algo útil para mí mismo.

> **Nombre:** *Miyano* (ミヤノ).

## 🧰 Stack

| Capa | Tecnología |
|---|---|
| **Backend** | Go · Chi · sqlc · pgx · argon2id · JWT |
| **BD** | PostgreSQL · MongoDB · SQLite (local) · MinIO (avatares) |
| **Frontend** | Vue 3 · TypeScript · Vite · Tailwind v4 · Pinia · Capacitor |
| **Infra** | Docker · GitHub Actions · Coolify + VPS (planeado) |

Arquitectura: **offline-first** (el móvil funciona sin red; sincroniza cuando hay conexión).

## 🚀 Quick start

### Backend (Go)
```bash
cd backend
go run ./cmd/api        # arranca la API en :8080
```
Healthcheck: http://localhost:8080/health

### Frontend (Vue + Capacitor)
> *Se añadirá cuando se implemente.*

## 🔧 Variables de entorno
> *Se definirán al añadir la conexión a BD.*

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
