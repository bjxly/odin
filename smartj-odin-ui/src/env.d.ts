/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 是否启用登录与权限控制（默认 true；为 false 时进入免登录模式） */
  readonly VITE_AUTH_ENABLED?: string
  /** 免登录模式下由运维预签发的长期 JWT */
  readonly VITE_INTERNAL_TOKEN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '@vue/runtime-core' {
  export interface ComponentCustomProperties {
    $t: (key: string, ...args: any[]) => string
    $tc: (key: string, ...args: any[]) => string
    $te: (key: string, locale?: string) => boolean
    $d: (value: number | Date, ...args: any[]) => string
    $n: (value: number, ...args: any[]) => string
  }
}
export {}
