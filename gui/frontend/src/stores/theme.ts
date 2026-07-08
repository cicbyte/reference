import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const currentTheme = ref('dark')
  const systemTheme = ref('dark')

  const isDark = computed(() => currentTheme.value === 'dark')

  function setTheme(t: string) {
    currentTheme.value = t
    document.documentElement.setAttribute('data-theme', t)
    localStorage.setItem('reference-theme', t)
  }

  function toggleTheme() {
    setTheme(isDark.value ? 'light' : 'dark')
  }

  function detectSystemTheme() {
    systemTheme.value = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }

  function initializeTheme() {
    detectSystemTheme()
    const saved = localStorage.getItem('reference-theme')
    setTheme(saved || systemTheme.value)

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      systemTheme.value = e.matches ? 'dark' : 'light'
      if (!localStorage.getItem('reference-theme')) {
        setTheme(systemTheme.value)
      }
    })
  }

  return { currentTheme, systemTheme, isDark, setTheme, toggleTheme, initializeTheme }
})
