import type { MenuTreeVO } from '@/types/sys'

// 依据菜单标题，递归查找目标菜单路径
export function findMenuPath(menus: MenuTreeVO[], targetTitles: string[]): string | undefined {
  for (const menu of menus) {
    if (menu?.meta?.title?.trim() === targetTitles[0]) {
      if (targetTitles.length === 1) {
        return menu.path // 找到最终目标，返回路径
      }
      if (menu.children) {
        const subPath = findMenuPath(menu.children, targetTitles.slice(1))
        if (subPath) return subPath
      }
    }
  }
  return undefined
}
