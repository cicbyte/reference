/// <reference types="vite/client" />

declare module '*.vue' {
    import type {DefineComponent} from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
}

// Wails injects bound Go methods at window.go.<package>.<Struct>
interface Window {
    go?: any
}
