/** 将文本保存至剪切板中 */
export async function saveStringToClipBoard(v: string | undefined): Promise<boolean> {
  const value = v || ''
  // 检查是否存在clipboard api
  if (!navigator.clipboard) {
    return false
  }
  await navigator.clipboard.writeText(value)
  return true
}
