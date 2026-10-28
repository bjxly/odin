import type { Directive } from 'vue'

const LazyClass = 'sm-img-lazy'
const LoadedClass = 'sm-img-lazy--loaded'
const ErrorClass = 'sm-img-lazy--error'

const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement
        const src = img.dataset.src
        if (src) {
          img.src = src
          img.onload = () => {
            img.classList.remove(LazyClass)
            img.classList.add(LoadedClass)
          }
          img.onerror = () => {
            img.classList.remove(LazyClass)
            img.classList.add(ErrorClass)
          }
          observer.unobserve(img)
        }
      }
    })
  },
  { rootMargin: '50px' }
)

export const lazy: Directive<HTMLImageElement, string> = {
  mounted(el, binding) {
    const src = binding.value
    if (!src) return
    el.classList.add(LazyClass)
    el.dataset.src = src
    observer.observe(el)
  },
  updated(el, binding) {
    if (binding.value !== binding.oldValue && binding.value) {
      el.dataset.src = binding.value
      el.classList.remove(LoadedClass, ErrorClass)
      el.classList.add(LazyClass)
      observer.observe(el)
    }
  },
  unmounted(el) {
    observer.unobserve(el)
  },
}

export { LazyClass, LoadedClass, ErrorClass }
