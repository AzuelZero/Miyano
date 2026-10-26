import { customRef, type Ref } from 'vue'

/** A ref that only propagates its value after `delay` ms of calm (search inputs). */
export function useDebouncedRef<T>(initial: T, delay = 250): Ref<T> {
  let value = initial
  let timer: ReturnType<typeof setTimeout> | undefined

  return customRef<T>((track, trigger) => ({
    get() {
      track()
      return value
    },
    set(newValue: T) {
      clearTimeout(timer)
      timer = setTimeout(() => {
        value = newValue
        trigger()
      }, delay)
    },
  }))
}
