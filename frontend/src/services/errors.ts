import { ApiError } from './api'

/** Server errors → human Spanish that names the recovery (DESIGN.md copy rule). */
export function authErrorMessage(err: unknown): string {
  if (!(err instanceof ApiError)) {
    return 'No se pudo conectar con el servidor. ¿Está la API corriendo?'
  }
  switch (err.status) {
    case 400:
      return err.message.includes('password')
        ? 'La contraseña debe tener al menos 8 caracteres.'
        : err.message
    case 401:
      return 'Credenciales inválidas. Revisa el correo y la contraseña.'
    case 409:
      return 'No se pudo crear la cuenta: ese correo ya está registrado.'
    case 429:
      return 'Demasiados intentos. Espera un minuto y vuelve a probar.'
    default:
      return 'Algo se rompió de nuestro lado. Inténtalo de nuevo en un momento.'
  }
}
