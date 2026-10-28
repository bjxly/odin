import { onMounted, onBeforeUnmount, type Ref } from 'vue'

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  radius: number
  alpha: number
  alphaDir: number
}

export interface ParticleCanvasOptions {
  /** 桌面端粒子数量，默认 180 */
  countDesktop?: number
  /** 移动端粒子数量，默认 45 */
  countMobile?: number
  /** 连线最大距离，默认 125 */
  linkDistance?: number
  /** 鼠标排斥半径，默认 145 */
  mouseRadius?: number
  /** 粒子速度系数，默认 0.38 */
  speed?: number
  /** 粒子半径范围，默认 [0.5, 2.0] */
  radiusRange?: [number, number]
  /** 透明度范围，默认 [0.15, 0.5] */
  alphaRange?: [number, number]
  /** 透明度变化速率，默认 0.0025 */
  alphaSpeed?: number
  /** 鼠标排斥力系数，默认 0.018 */
  mouseForce?: number
  /** 速度衰减，默认 0.99 */
  damping?: number
  /** 连线颜色 RGB，默认 '100, 180, 255' */
  lineColorRgb?: string
  /** 连线最大透明度，默认 0.2 */
  lineMaxOpacity?: number
  /** 连线宽度，默认 0.55 */
  lineWidth?: number
  /** 粒子颜色 RGB，默认 '130, 195, 255' */
  dotColorRgb?: string
}

/**
 * 登录页 Canvas 粒子动画 composable
 * 可在多个主题中复用，通过 options 调整视觉参数
 */
export function useParticleCanvas(canvasRef: Ref<HTMLCanvasElement | undefined>, options: ParticleCanvasOptions = {}) {
  const {
    countDesktop = 180,
    countMobile = 45,
    linkDistance = 125,
    mouseRadius = 145,
    speed = 0.38,
    radiusRange = [0.5, 2.0],
    alphaRange = [0.15, 0.5],
    alphaSpeed = 0.0025,
    mouseForce = 0.018,
    damping = 0.99,
    lineColorRgb = '100, 180, 255',
    lineMaxOpacity = 0.2,
    lineWidth = 0.55,
    dotColorRgb = '130, 195, 255',
  } = options

  let particles: Particle[] = []
  let animFrameId = 0
  const mouse = { x: -9999, y: -9999 }
  let canvasW = 0
  let canvasH = 0

  function getParticleCount() {
    return window.innerWidth < 768 ? countMobile : countDesktop
  }

  function initParticles() {
    particles = []
    const count = getParticleCount()
    const [rMin, rMax] = radiusRange
    const [aMin, aMax] = alphaRange
    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * canvasW,
        y: Math.random() * canvasH,
        vx: (Math.random() - 0.5) * speed,
        vy: (Math.random() - 0.5) * speed,
        radius: Math.random() * (rMax - rMin) + rMin,
        alpha: Math.random() * (aMax - aMin) + aMin,
        alphaDir: Math.random() > 0.5 ? 1 : -1,
      })
    }
  }

  function draw(ctx: CanvasRenderingContext2D) {
    ctx.clearRect(0, 0, canvasW, canvasH)
    // 连线
    for (let i = 0; i < particles.length; i++) {
      for (let j = i + 1; j < particles.length; j++) {
        const dx = particles[i].x - particles[j].x
        const dy = particles[i].y - particles[j].y
        const dist = Math.sqrt(dx * dx + dy * dy)
        if (dist < linkDistance) {
          const opacity = (1 - dist / linkDistance) * lineMaxOpacity
          ctx.beginPath()
          ctx.strokeStyle = `rgba(${lineColorRgb}, ${opacity})`
          ctx.lineWidth = lineWidth
          ctx.moveTo(particles[i].x, particles[i].y)
          ctx.lineTo(particles[j].x, particles[j].y)
          ctx.stroke()
        }
      }
    }
    // 粒子
    for (const p of particles) {
      ctx.beginPath()
      ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2)
      ctx.fillStyle = `rgba(${dotColorRgb}, ${p.alpha})`
      ctx.fill()
    }
  }

  function update() {
    for (const p of particles) {
      // 鼠标排斥
      const dx = p.x - mouse.x
      const dy = p.y - mouse.y
      const dist = Math.sqrt(dx * dx + dy * dy)
      if (dist < mouseRadius && dist > 0) {
        const force = ((mouseRadius - dist) / mouseRadius) * mouseForce
        p.vx += (dx / dist) * force
        p.vy += (dy / dist) * force
      }
      p.vx *= damping
      p.vy *= damping
      p.x += p.vx
      p.y += p.vy
      // 透明度呼吸
      p.alpha += p.alphaDir * alphaSpeed
      if (p.alpha > alphaRange[1]) {
        p.alpha = alphaRange[1]
        p.alphaDir = -1
      } else if (p.alpha < alphaRange[0]) {
        p.alpha = alphaRange[0]
        p.alphaDir = 1
      }
      // 边界环绕
      if (p.x < -10) p.x = canvasW + 10
      if (p.x > canvasW + 10) p.x = -10
      if (p.y < -10) p.y = canvasH + 10
      if (p.y > canvasH + 10) p.y = -10
    }
  }

  function animate() {
    const canvas = canvasRef.value
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    update()
    draw(ctx)
    animFrameId = requestAnimationFrame(animate)
  }

  function resizeCanvas() {
    const canvas = canvasRef.value
    if (!canvas) return
    canvasW = window.innerWidth
    canvasH = window.innerHeight
    canvas.width = canvasW
    canvas.height = canvasH
    initParticles()
  }

  function onMouseMove(e: MouseEvent) {
    mouse.x = e.clientX
    mouse.y = e.clientY
  }
  function onMouseLeave() {
    mouse.x = -9999
    mouse.y = -9999
  }
  function onVisibilityChange() {
    if (document.hidden) cancelAnimationFrame(animFrameId)
    else animFrameId = requestAnimationFrame(animate)
  }

  onMounted(() => {
    resizeCanvas()
    window.addEventListener('resize', resizeCanvas)
    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseleave', onMouseLeave)
    document.addEventListener('visibilitychange', onVisibilityChange)
    animFrameId = requestAnimationFrame(animate)
  })

  onBeforeUnmount(() => {
    cancelAnimationFrame(animFrameId)
    window.removeEventListener('resize', resizeCanvas)
    window.removeEventListener('mousemove', onMouseMove)
    window.removeEventListener('mouseleave', onMouseLeave)
    document.removeEventListener('visibilitychange', onVisibilityChange)
  })
}
