import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import en from './locales/en'

export type AppLocale = 'zh-CN' | 'en'

const STORAGE_KEY = 'reference-locale'

/** Read the persisted locale, defaulting to zh-CN. */
export function getInitialLocale(): AppLocale {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'en' || saved === 'zh-CN') return saved
  return 'zh-CN'
}

const i18n = createI18n({
  legacy: false,            // use Composition API mode
  locale: getInitialLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    en,
  },
})

export default i18n
