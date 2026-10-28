import { ref, onMounted, onUnmounted, computed } from 'vue'

export function useResponsive() {
  const windowWidth = ref(window.innerWidth)

  function onResize() {
    windowWidth.value = window.innerWidth
  }

  onMounted(() => window.addEventListener('resize', onResize))
  onUnmounted(() => window.removeEventListener('resize', onResize))

  const isMobile = computed(() => windowWidth.value < 768)
  const isTablet = computed(() => windowWidth.value >= 768 && windowWidth.value < 992)
  const isDesktop = computed(() => windowWidth.value >= 992)

  return { windowWidth, isMobile, isTablet, isDesktop }
}
