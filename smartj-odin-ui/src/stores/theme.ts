import { defineStore } from 'pinia'

// 默认配置
const DEFAULT_CONFIG = {
  isDark: false,
  watermark: false,
  showNavIcon: true,
  showTagsView: true,
  navBgStart: '#1d1e3a',
  navBgEnd: '#2d2f5e',
  layoutMode: 'horizontal' as string,
}

export const useThemeStore = defineStore('theme', {
  state: () => ({
    ...DEFAULT_CONFIG,
  }),
  getters: {
    navBackground: (state) => `linear-gradient(90deg, ${state.navBgStart} 0%, ${state.navBgEnd} 100%)`,
  },
  actions: {
    toggleDark() {
      this.isDark = !this.isDark
      this.applyTheme()
    },
    setDark(dark: boolean) {
      this.isDark = dark
      this.applyTheme()
    },
    applyTheme() {
      if (this.isDark) {
        document.documentElement.setAttribute('data-theme', 'dark')
      } else {
        document.documentElement.removeAttribute('data-theme')
      }
      document.documentElement.style.setProperty('--sm-nav-bg-start', this.navBgStart)
      document.documentElement.style.setProperty('--sm-nav-bg-end', this.navBgEnd)
      this.saveConfig()
    },
    saveConfig() {
      const config = {
        isDark: this.isDark,
        watermark: this.watermark,
        showNavIcon: this.showNavIcon,
        showTagsView: this.showTagsView,
        navBgStart: this.navBgStart,
        navBgEnd: this.navBgEnd,
        layoutMode: this.layoutMode,
      }
      localStorage.setItem('sm-theme-config', JSON.stringify(config))
    },
    initTheme() {
      const saved = localStorage.getItem('sm-theme-config')
      if (saved) {
        try {
          const config = JSON.parse(saved)
          Object.assign(this, config)
        } catch {
          // ignore
        }
      }
      this.applyTheme()
    },
    resetConfig() {
      Object.assign(this, { ...DEFAULT_CONFIG })
      this.applyTheme()
    },
  },
})
