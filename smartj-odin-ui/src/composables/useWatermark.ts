export function useWatermark() {
  function createWatermark(text: string) {
    removeWatermark()
    const canvas = document.createElement('canvas')
    canvas.width = 300
    canvas.height = 200
    const ctx = canvas.getContext('2d')!
    ctx.rotate((-20 * Math.PI) / 180)
    ctx.font = '14px Arial'
    ctx.fillStyle = 'rgba(180, 180, 180, 0.15)'
    ctx.fillText(text, 50, 140)

    const div = document.createElement('div')
    div.id = 'sm-watermark'
    div.style.cssText = [
      'position: fixed',
      'top: 0',
      'left: 0',
      'right: 0',
      'bottom: 0',
      'z-index: 9999',
      'pointer-events: none',
      `background: url(${canvas.toDataURL()}) repeat`,
    ].join(';')
    document.body.appendChild(div)
  }

  function removeWatermark() {
    const el = document.getElementById('sm-watermark')
    if (el) el.remove()
  }

  return { createWatermark, removeWatermark }
}
