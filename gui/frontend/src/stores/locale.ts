import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import i18n, { getInitialLocale, type AppLocale } from '@/i18n'

const STORAGE_KEY = 'reference-locale'

export const useLocaleStore = defineStore('locale', () => {
  const currentLocale = ref<AppLocale>(getInitialLocale())
  const isEn = computed(() => currentLocale.value === 'en')
  const isZh = computed(() => currentLocale.value === 'zh-CN')

  /** Switch the active locale and persist the choice. */
  function setLocale(locale: AppLocale) {
    currentLocale.value = locale
    i18n.global.locale.value = locale
    localStorage.setItem(STORAGE_KEY, locale)
    document.documentElement.setAttribute('lang', locale === 'en' ? 'en' : 'zh-CN')
  }

  function toggleLocale() {
    setLocale(isEn.value ? 'zh-CN' : 'en')
  }

  return { currentLocale, isEn, isZh, setLocale, toggleLocale }
})
