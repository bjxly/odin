import {
  Engine,
  Scene,
  ArcRotateCamera,
  HemisphericLight,
  DirectionalLight,
  Vector3,
  Color3,
  Color4,
  MeshBuilder,
  StandardMaterial,
  TransformNode,
  Mesh,
  ShadowGenerator,
  Animation,
  CubicEase,
  EasingFunction,
  HighlightLayer,
  AbstractMesh,
  DynamicTexture,
  Matrix,
} from '@babylonjs/core'

// ======================== Types ========================

export interface WarehouseConfig {
  width: number
  length: number
  height: number
  aisleWidth: number
  rackDepth: number
  pairCount: number
  rackRows: number
  rackColumns: number
  rackLevels: number
  levelHeight: number
  beamThickness: number
  uprightWidth: number
  colors: { upright: string; beam: string; floor: string; wall: string }
}

export interface RackSection {
  id: string
  label: string
  categoryLabel: string
  row: number // 0-3, which back-to-back pair along the aisle
  levels: number
  columns: number
  palletWidth: number
  palletDepth: number
  facing: number // +1 = faces +X, -1 = faces -X (back-to-back partner)
}

export interface CoreBoxSpec {
  name: string
  width: number
  depth: number
  height: number
  color: string
}

export interface PalletConfig {
  width: number
  depth: number
  height: number
  maxBoxes: number
  defaultBoxes: number
}

export interface PalletSlot {
  palletId: string
  rackId: string
  level: number
  column: number
  boxSpec: string
  boxCount: number
  isOccupied: boolean
  projectNo?: string
  holeNo?: string
  depthFrom?: number
  depthTo?: number
  coreLength?: number
}

export interface SelectionInfo {
  type: 'rack' | 'pallet' | 'box'
  rackId?: string
  label?: string
  categoryLabel?: string
  level?: number
  column?: number
  boxSpec?: string
  boxCount?: number
  projectNo?: string
  holeNo?: string
  depthFrom?: number
  depthTo?: number
  coreLength?: number
}

export interface TooltipInfo {
  x: number
  y: number
  rackId: string
  slotLabel: string
  projectNo: string
  holeNo: string
  depthFrom: number
  depthTo: number
  coreLength: number
  boxSpec: string
  boxCount: number
}

// ======================== Scene Controller ========================

export class WarehouseScene {
  private engine!: Engine
  private scene!: Scene
  private camera!: ArcRotateCamera
  private highlight!: HighlightLayer
  private forkliftRoot!: TransformNode
  private forksNode!: TransformNode
  private allMeshes: Mesh[] = []
  private selectedMesh: AbstractMesh | null = null
  private rackPositions = new Map<string, Vector3>()
  private isAnimating = false

  public config!: WarehouseConfig
  public rackSections!: RackSection[]
  public inventory!: PalletSlot[]
  public boxSpecs!: CoreBoxSpec[]
  public palletConfig!: PalletConfig
  public onSelectionChange?: (info: SelectionInfo | null) => void
  public onTooltipShow?: (info: TooltipInfo) => void
  public onTooltipHide?: () => void

  constructor(
    private canvas: HTMLCanvasElement,
    data: {
      warehouse: WarehouseConfig
      rackSections: RackSection[]
      inventory: PalletSlot[]
      boxSpecs: CoreBoxSpec[]
      palletConfig: PalletConfig
    }
  ) {
    this.config = data.warehouse
    this.rackSections = data.rackSections
    this.inventory = data.inventory
    this.boxSpecs = data.boxSpecs
    this.palletConfig = data.palletConfig
  }

  // ==================== Lifecycle ====================

  init() {
    this.engine = new Engine(this.canvas, true, { preserveDrawingBuffer: true, stencil: true })
    this.scene = new Scene(this.engine)
    this.scene.clearColor = new Color4(0.06, 0.08, 0.14, 1)

    this.setupCamera()
    this.setupLighting()
    this.buildWarehouse()
    this.setupInteraction()

    this.engine.runRenderLoop(() => this.scene.render())
    window.addEventListener('resize', this.onResize)
  }

  dispose() {
    window.removeEventListener('resize', this.onResize)
    this.scene?.dispose()
    this.engine?.dispose()
  }

  rebuild() {
    this.scene.stopAllAnimations()
    this.isAnimating = false
    this.allMeshes.forEach((m) => m.dispose())
    this.allMeshes = []
    this.rackPositions.clear()
    this.highlight?.dispose()
    this.selectedMesh = null
    this.buildWarehouse()
    this.setupInteraction()
  }

  updateConfig(partial: Partial<WarehouseConfig>) {
    Object.assign(this.config, partial)
    this.rebuild()
  }

  // ==================== Camera ====================

  private setupCamera() {
    const center = new Vector3(0, 3, 0)
    this.camera = new ArcRotateCamera('camera', -Math.PI / 3, Math.PI / 3.2, 35, center, this.scene)
    this.camera.attachControl(this.canvas, true)
    this.camera.lowerRadiusLimit = 5
    this.camera.upperRadiusLimit = 80
    this.camera.wheelDeltaPercentage = 0.01
  }

  // ==================== Lighting ====================

  private setupLighting() {
    new HemisphericLight('ambient', new Vector3(0, 1, 0), this.scene)

    const dir = new DirectionalLight('dir', new Vector3(-1, -2, -1), this.scene)
    dir.position = new Vector3(20, 30, 20)
    dir.intensity = 0.7

    const shadow = new ShadowGenerator(1024, dir)
    shadow.useBlurExponentialShadowMap = true
    shadow.blurScale = 2
  }

  // ==================== Label Helpers ====================

  private createLabelPlane(name: string, text: string, width: number, height: number, fontSize: number, bgColor: string, textColor: string): Mesh {
    const plane = MeshBuilder.CreatePlane(name, { width, height }, this.scene)
    const mat = new StandardMaterial(`${name}_mat`, this.scene)
    const tex = new DynamicTexture(`${name}_tex`, { width: 256, height: 128 }, this.scene, true)
    const ctx = tex.getContext() as unknown as CanvasRenderingContext2D
    ctx.fillStyle = bgColor
    ctx.fillRect(0, 0, 256, 128)
    ctx.fillStyle = textColor
    ctx.font = `bold ${fontSize}px Arial`
    ;(ctx as any).textAlign = 'center'
    ;(ctx as any).textBaseline = 'middle'
    ctx.fillText(text, 128, 64)
    tex.update()
    mat.diffuseTexture = tex
    mat.emissiveTexture = tex
    mat.backFaceCulling = false
    plane.material = mat
    this.allMeshes.push(plane)
    return plane
  }

  // ==================== Warehouse Build ====================

  private buildWarehouse() {
    this.buildGround()
    this.buildWalls()
    this.buildAllRacks()
    this.buildForklift()
    this.highlight = new HighlightLayer('highlight', this.scene)
  }

  private buildGround() {
    const { width, length, colors } = this.config
    const ground = MeshBuilder.CreateGround('ground', { width, height: length, subdivisions: 2 }, this.scene)
    const mat = new StandardMaterial('groundMat', this.scene)
    mat.diffuseColor = Color3.FromHexString(colors.floor)
    mat.specularColor = new Color3(0.05, 0.05, 0.05)
    ground.material = mat
    ground.receiveShadows = true
    this.allMeshes.push(ground)
  }

  private buildWalls() {
    const { width, length, height, colors } = this.config
    const mat = new StandardMaterial('wallMat', this.scene)
    mat.diffuseColor = Color3.FromHexString(colors.wall)
    mat.alpha = 0.08

    const specs = [
      { name: 'wallBack', w: width, h: height, d: 0.1, pos: new Vector3(0, height / 2, -length / 2) },
      { name: 'wallLeft', w: 0.1, h: height, d: length, pos: new Vector3(-width / 2, height / 2, 0) },
      { name: 'wallRight', w: 0.1, h: height, d: length, pos: new Vector3(width / 2, height / 2, 0) },
    ]
    for (const s of specs) {
      const wall = MeshBuilder.CreateBox(s.name, { width: s.w, height: s.h, depth: s.d }, this.scene)
      wall.position = s.pos
      wall.material = mat
      this.allMeshes.push(wall)
    }
  }

  // ==================== Rack System ====================

  private calcRackOrigin(row: number, facing: number): { x: number; z: number } {
    const { aisleWidth, rackDepth, pairCount } = this.config
    // 4 back-to-back pairs along Z axis (aisle direction)
    const pairZ = (row - (pairCount - 1) / 2) * (rackDepth + aisleWidth)
    // X: back-to-back shelves at ±rackDepth/2
    const x = (facing * rackDepth) / 2
    return { x, z: pairZ }
  }

  private buildAllRacks() {
    for (const section of this.rackSections) {
      this.buildRackSection(section)
    }
  }

  private buildRackSection(section: RackSection) {
    const { rackLevels, levelHeight, beamThickness, uprightWidth, rackDepth, colors } = this.config
    const { palletWidth, palletDepth, columns } = section
    const origin = this.calcRackOrigin(section.row, section.facing)
    const totalH = rackLevels * levelHeight
    const rackW = palletWidth * columns

    // Rack parent node — NO rotation, long side (columns) naturally faces the aisle
    const rackNode = new TransformNode(`rack_${section.id}`, this.scene)
    rackNode.position = new Vector3(origin.x, 0, origin.z)
    this.rackPositions.set(section.id, new Vector3(origin.x, 0, origin.z))

    // 4 Uprights at corners
    const upMat = new StandardMaterial(`up_mat_${section.id}`, this.scene)
    upMat.diffuseColor = Color3.FromHexString(colors.upright)
    upMat.specularColor = new Color3(0.2, 0.2, 0.2)

    const hw = rackW / 2
    const hd = rackDepth / 2
    for (const [dx, dz] of [
      [-hw, -hd],
      [-hw, hd],
      [hw, -hd],
      [hw, hd],
    ]) {
      const up = MeshBuilder.CreateBox(`up_${section.id}_${dx}_${dz}`, { width: uprightWidth, height: totalH, depth: uprightWidth }, this.scene)
      up.position = new Vector3(dx, totalH / 2, dz)
      up.parent = rackNode
      up.material = upMat
      this.allMeshes.push(up)
    }

    // Beams (front + back per level)
    const beamMat = new StandardMaterial(`beam_mat_${section.id}`, this.scene)
    beamMat.diffuseColor = Color3.FromHexString(colors.beam)
    beamMat.specularColor = new Color3(0.1, 0.1, 0.1)

    for (let lv = 0; lv < rackLevels; lv++) {
      const y = lv * levelHeight + beamThickness / 2
      for (const dz of [-hd, hd]) {
        const beam = MeshBuilder.CreateBox(`beam_${section.id}_${lv}_${dz}`, { width: rackW + uprightWidth, height: beamThickness, depth: uprightWidth }, this.scene)
        beam.position = new Vector3(0, y, dz)
        beam.parent = rackNode
        beam.material = beamMat
        this.allMeshes.push(beam)
      }
    }

    // ★ Rack ID labels on aisle-facing side (front face, at top)
    // facing=+1: label at +Z side facing +Z; facing=-1: label at -Z side facing -Z
    const labelY = totalH - levelHeight * 0.3
    const labelZ = section.facing * (hd + 0.05)
    const labelRotY = section.facing === 1 ? 0 : Math.PI

    // Main rack ID label (A1, A2, B1, etc.)
    const idLabel = this.createLabelPlane(`label_${section.id}`, section.label, 1.2, 0.5, 80, '#e85d2a', '#ffffff')
    idLabel.position = new Vector3(0, labelY, labelZ)
    idLabel.rotation.y = labelRotY
    idLabel.parent = rackNode

    // Category label below ID
    const catLabel = this.createLabelPlane(`catlabel_${section.id}`, section.categoryLabel, 1.8, 0.35, 44, '#e85d2a', '#ffffff')
    catLabel.position = new Vector3(0, labelY - 0.55, labelZ)
    catLabel.rotation.y = labelRotY
    catLabel.parent = rackNode

    // ★ Pallet slot labels on aisle-facing beam — format: A1-列-层 (e.g. A1-1-1, A1-2-1)
    const plabelZ = section.facing * (hd + 0.02)
    const plabelRotY = section.facing === 1 ? 0 : Math.PI
    for (let lv = 0; lv < rackLevels; lv++) {
      for (let col = 0; col < columns; col++) {
        const slotX = (col - (columns - 1) / 2) * palletWidth
        const slotY = lv * levelHeight + beamThickness
        const slotLabel = `${section.label}-${col + 1}-${lv + 1}`
        const plabelPlane = this.createLabelPlane(`plabel_${section.id}_${lv}_${col}`, slotLabel, 0.55, 0.2, 40, '#334155', '#f1f5f9')
        plabelPlane.position = new Vector3(slotX, slotY + 0.08, plabelZ)
        plabelPlane.rotation.y = plabelRotY
        plabelPlane.parent = rackNode
      }
    }

    // ★ Build occupied/empty slots
    for (let lv = 0; lv < rackLevels; lv++) {
      for (let col = 0; col < columns; col++) {
        const slotX = (col - (columns - 1) / 2) * palletWidth
        const slotY = lv * levelHeight + beamThickness
        const inv = this.inventory.find((i) => i.rackId === section.id && i.level === lv && i.column === col)

        if (inv?.isOccupied) {
          this.buildOccupiedSlot(section, inv, slotX, slotY, 0, rackNode, palletWidth, palletDepth)
        } else {
          // Empty slot indicator
          const empty = MeshBuilder.CreateBox(`empty_${section.id}_${lv}_${col}`, { width: palletWidth * 0.85, height: 0.02, depth: palletDepth * 0.85 }, this.scene)
          empty.position = new Vector3(slotX, slotY + 0.01, 0)
          empty.parent = rackNode
          const em = new StandardMaterial(`emptyMat_${section.id}_${lv}_${col}`, this.scene)
          em.diffuseColor = new Color3(0.25, 0.25, 0.25)
          em.alpha = 0.25
          empty.material = em
          empty.metadata = { type: 'slot', rackId: section.id, level: lv, column: col, label: section.label, categoryLabel: section.categoryLabel }
          this.allMeshes.push(empty)
        }
      }
    }

    // Make the whole rack node clickable via an invisible bounding box
    const rackBounds = MeshBuilder.CreateBox(`rackHit_${section.id}`, { width: rackW, height: totalH, depth: rackDepth }, this.scene)
    rackBounds.position = new Vector3(0, totalH / 2, 0)
    rackBounds.parent = rackNode
    rackBounds.isVisible = false
    rackBounds.isPickable = true
    rackBounds.metadata = { type: 'rack', rackId: section.id, label: section.label, categoryLabel: section.categoryLabel }
  }

  private buildOccupiedSlot(section: RackSection, inv: PalletSlot, sx: number, sy: number, sz: number, parent: TransformNode, pw: number, pd: number) {
    const pc = this.palletConfig

    // Pallet base
    const pallet = MeshBuilder.CreateBox(`pallet_${inv.palletId}`, { width: pw * 0.9, height: pc.height / 100, depth: pd * 0.9 }, this.scene)
    pallet.position = new Vector3(sx, sy + pc.height / 100 / 2, sz)
    pallet.parent = parent
    const pm = new StandardMaterial(`palletMat_${inv.palletId}`, this.scene)
    pm.diffuseColor = new Color3(0.55, 0.42, 0.28)
    pm.specularColor = new Color3(0.05, 0.05, 0.05)
    pallet.material = pm
    pallet.metadata = {
      type: 'pallet',
      rackId: section.id,
      label: section.label,
      categoryLabel: section.categoryLabel,
      level: inv.level,
      column: inv.column,
      palletId: inv.palletId,
      boxSpec: inv.boxSpec,
      boxCount: inv.boxCount,
      projectNo: inv.projectNo,
      holeNo: inv.holeNo,
      depthFrom: inv.depthFrom,
      depthTo: inv.depthTo,
      coreLength: inv.coreLength,
    }
    this.allMeshes.push(pallet)

    // Core boxes stack
    const spec = this.boxSpecs.find((b) => b.name === inv.boxSpec) || this.boxSpecs[0]
    const stack = this.buildBoxStack(inv.palletId, spec, inv.boxCount, pc.height / 100, inv)
    stack.position = new Vector3(sx, sy + pc.height / 100, sz)
    stack.parent = parent
    stack.metadata = {
      type: 'box',
      rackId: section.id,
      label: section.label,
      categoryLabel: section.categoryLabel,
      level: inv.level,
      column: inv.column,
      palletId: inv.palletId,
      boxSpec: inv.boxSpec,
      boxCount: inv.boxCount,
      slotLabel: `${section.label}-${inv.column + 1}-${inv.level + 1}`,
      projectNo: inv.projectNo,
      holeNo: inv.holeNo,
      depthFrom: inv.depthFrom,
      depthTo: inv.depthTo,
      coreLength: inv.coreLength,
    }
  }

  private buildBoxStack(palletId: string, spec: CoreBoxSpec, count: number, _palletH: number, inv: PalletSlot): TransformNode {
    const node = new TransformNode(`stack_${palletId}`, this.scene)
    const mat = new StandardMaterial(`boxMat_${palletId}`, this.scene)
    mat.diffuseColor = Color3.FromHexString(spec.color)
    mat.specularColor = new Color3(0.1, 0.1, 0.1)

    for (let i = 0; i < count; i++) {
      const box = MeshBuilder.CreateBox(`box_${palletId}_${i}`, { width: spec.width / 100, height: spec.height / 100, depth: spec.depth / 100 }, this.scene)
      box.position.y = i * (spec.height / 100) + spec.height / 100 / 2
      box.parent = node
      box.material = mat
      // Each individual box carries the inventory metadata for tooltip
      box.metadata = {
        type: 'box',
        rackId: inv.rackId,
        label: inv.rackId,
        slotLabel: `${inv.rackId}-${inv.column + 1}-${inv.level + 1}`,
        projectNo: inv.projectNo,
        holeNo: inv.holeNo,
        depthFrom: inv.depthFrom,
        depthTo: inv.depthTo,
        coreLength: inv.coreLength,
        boxSpec: inv.boxSpec,
        boxCount: inv.boxCount,
      }
      this.allMeshes.push(box)
    }
    return node
  }

  // ==================== Forklift ====================

  private buildForklift() {
    this.forkliftRoot = new TransformNode('forklift', this.scene)
    const { rackDepth } = this.config
    // Forklift at the FRONT entrance of the first pair (A1/A2), between the two shelves
    // X=0 (aisle center between A1 and A2), Z just in front of the rack face
    const firstPairZ = this.rackPositions.get('A1')?.z ?? 0
    this.forkliftRoot.position = new Vector3(0, 0, firstPairZ - rackDepth / 2 - 1.5)

    // Body
    const body = MeshBuilder.CreateBox('forklift_body', { width: 1.2, height: 1.8, depth: 2.2 }, this.scene)
    body.position.y = 1.1
    body.parent = this.forkliftRoot
    const bm = new StandardMaterial('forkliftMat', this.scene)
    bm.diffuseColor = new Color3(0.9, 0.75, 0)
    bm.specularColor = new Color3(0.2, 0.2, 0.2)
    body.material = bm
    this.allMeshes.push(body)

    // Cabin
    const cabin = MeshBuilder.CreateBox('forklift_cabin', { width: 1.1, height: 1.0, depth: 1.0 }, this.scene)
    cabin.position = new Vector3(0, 2.2, -0.4)
    cabin.parent = this.forkliftRoot
    cabin.material = bm
    this.allMeshes.push(cabin)

    // Mast (vertical, in front of body)
    const mast = MeshBuilder.CreateBox('forklift_mast', { width: 0.12, height: 5.5, depth: 0.12 }, this.scene)
    mast.position = new Vector3(0, 2.75, 1.2)
    mast.parent = this.forkliftRoot
    const mm = new StandardMaterial('mastMat', this.scene)
    mm.diffuseColor = new Color3(0.35, 0.35, 0.35)
    mast.material = mm
    this.allMeshes.push(mast)

    // Forks (animated, extend along +Z toward racks)
    this.forksNode = new TransformNode('forks', this.scene)
    this.forksNode.position = new Vector3(0, 0.5, 1.4)
    this.forksNode.parent = this.forkliftRoot

    for (const dx of [-0.25, 0.25]) {
      const fork = MeshBuilder.CreateBox(`fork_${dx}`, { width: 0.12, height: 0.06, depth: 1.2 }, this.scene)
      fork.position.x = dx
      fork.parent = this.forksNode
      fork.material = mm
      this.allMeshes.push(fork)
    }

    // Wheels
    const wm = new StandardMaterial('wheelMat', this.scene)
    wm.diffuseColor = new Color3(0.15, 0.15, 0.15)
    for (const [dx, dz] of [
      [-0.5, -0.8],
      [0.5, -0.8],
      [-0.5, 0.8],
      [0.5, 0.8],
    ]) {
      const wheel = MeshBuilder.CreateCylinder(`wheel_${dx}_${dz}`, { diameter: 0.4, height: 0.18 }, this.scene)
      wheel.position = new Vector3(dx, 0.2, dz)
      wheel.rotation.z = Math.PI / 2
      wheel.parent = this.forkliftRoot
      wheel.material = wm
      this.allMeshes.push(wheel)
    }
  }

  // ==================== Animation (FIXED) ====================

  /**
   * 修复：使用 setTimeout 等待动画帧完成后再 resolve，
   * 而不是 onAfterRenderObservable（每帧都触发，导致立即 resolve）
   */
  private animatePosition(mesh: TransformNode | Mesh, target: Vector3, durationMs: number): Promise<void> {
    return new Promise((resolve) => {
      const fps = 30
      const totalFrames = Math.max(Math.round(durationMs / (1000 / fps)), 1)
      const anim = new Animation('move_' + target.toString(), 'position', fps, Animation.ANIMATIONTYPE_VECTOR3, Animation.ANIMATIONLOOPMODE_CONSTANT)
      const ease = new CubicEase()
      ease.setEasingMode(EasingFunction.EASINGMODE_EASEINOUT)
      anim.setEasingFunction(ease)
      anim.setKeys([
        { frame: 0, value: mesh.position.clone() },
        { frame: totalFrames, value: target.clone() },
      ])
      mesh.animations.push(anim)
      this.scene.beginAnimation(mesh, 0, totalFrames, false, 1, () => {
        // onAnimationEnd 回调 — 动画帧播放完毕后再 resolve
        resolve()
      })
    })
  }

  private async animateForkliftPutAway(targetRackId: string, targetLevel: number) {
    const rackPos = this.rackPositions.get(targetRackId)
    const rackSection = this.rackSections.find((r) => r.id === targetRackId)
    if (!rackPos || !rackSection || this.isAnimating) return
    this.isAnimating = true

    const { levelHeight, rackDepth } = this.config
    const facing = rackSection.facing
    const pairZ = rackPos.z // Z of the back-to-back pair
    // Approach X: just outside the rack face along X axis
    const approachX = facing * (rackDepth / 2 + 1.5)
    const targetY = targetLevel * levelHeight + levelHeight / 2

    try {
      // 1. Drive along aisle (Z) to target pair
      await this.animatePosition(this.forkliftRoot, new Vector3(0, 0, pairZ), 2000)
      // 2. Move laterally (X) toward rack face
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, 0, pairZ), 1500)
      // 3. Lift to target height
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, targetY - 1, pairZ), 1500)
      // 4. Drive forward into rack
      await this.animatePosition(this.forkliftRoot, new Vector3(rackPos.x, targetY - 1, pairZ), 1200)
      // 5. Reverse out
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, targetY - 1, pairZ), 1000)
      // 6. Lower & return to aisle center
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, 0, pairZ), 1500)
      await this.animatePosition(this.forkliftRoot, new Vector3(0, 0, this.forkliftRoot.position.z), 2000)
    } finally {
      this.isAnimating = false
    }
  }

  private async animateForkliftPick(rackId: string, level: number) {
    const rackPos = this.rackPositions.get(rackId)
    const rackSection = this.rackSections.find((r) => r.id === rackId)
    if (!rackPos || !rackSection || this.isAnimating) return
    this.isAnimating = true

    const { levelHeight, rackDepth } = this.config
    const facing = rackSection.facing
    const pairZ = rackPos.z
    const approachX = facing * (rackDepth / 2 + 1.5)
    const targetY = level * levelHeight + levelHeight / 2

    try {
      await this.animatePosition(this.forkliftRoot, new Vector3(0, 0, pairZ), 2000)
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, 0, pairZ), 1500)
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, targetY - 1, pairZ), 1500)
      await this.animatePosition(this.forkliftRoot, new Vector3(rackPos.x, targetY - 1, pairZ), 1200)
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, targetY - 1, pairZ), 1000)
      await this.animatePosition(this.forkliftRoot, new Vector3(approachX, 0, pairZ), 1500)
      await this.animatePosition(this.forkliftRoot, new Vector3(0, 0, this.forkliftRoot.position.z), 2000)
    } finally {
      this.isAnimating = false
    }
  }

  async executePutAway(rackId: string, level: number, column: number, boxSpec: string, boxCount: number) {
    await this.animateForkliftPutAway(rackId, level)
    // Add to inventory
    const newId = `P${String(this.inventory.length + 1).padStart(3, '0')}`
    this.inventory.push({
      palletId: newId,
      rackId,
      level,
      column,
      boxSpec,
      boxCount,
      isOccupied: true,
      projectNo: 'NEW',
      holeNo: 'ZK-NEW',
      depthFrom: 0,
      depthTo: 0,
      coreLength: 0,
    })
    this.rebuild()
  }

  async executePick(rackId: string, level: number, column: number) {
    await this.animateForkliftPick(rackId, level)
    // Remove from inventory
    const idx = this.inventory.findIndex((i) => i.rackId === rackId && i.level === level && i.column === column && i.isOccupied)
    if (idx >= 0) this.inventory.splice(idx, 1)
    this.rebuild()
  }

  // ==================== Interaction ====================

  private setupInteraction() {
    this.scene.onPointerObservable.add((info) => {
      const pick = info.pickInfo
      if (!pick) return
      if (info.type === 2) this.handlePointerMove(pick)
      if (info.type === 3) this.handlePointerClick(pick)
    })
  }

  private handlePointerMove(pick: { hit: boolean; pickedMesh: AbstractMesh | null }) {
    this.canvas.style.cursor = pick.hit ? 'pointer' : 'default'
    this.highlight.removeAllMeshes()
    if (this.selectedMesh) this.highlight.addMesh(this.selectedMesh as Mesh, Color3.FromHexString('#4096ff'))
    if (pick.hit && pick.pickedMesh && pick.pickedMesh !== this.selectedMesh) {
      this.highlight.addMesh(pick.pickedMesh as Mesh, Color3.FromHexString('#ffc53d'))
    }

    // ★ Tooltip for core boxes
    if (pick.hit && pick.pickedMesh) {
      const meta = pick.pickedMesh.metadata
      if (meta?.type === 'box' && meta.projectNo) {
        // Project the mesh world position to screen coordinates
        const worldPos = pick.pickedMesh.getAbsolutePosition()
        const parent = pick.pickedMesh.parent
        if (parent) {
          const parentWorld = parent.getWorldMatrix()
          const absolutePos = Vector3.TransformCoordinates(worldPos, parentWorld)
          const viewport = this.camera.viewport
          const screenPos = Vector3.Project(absolutePos, Matrix.Identity(), this.scene.getTransformMatrix(), viewport)
          // Vector3.Project returns normalized [0,1] coords, convert to CSS pixels
          this.onTooltipShow?.({
            x: screenPos.x * this.canvas.clientWidth,
            y: screenPos.y * this.canvas.clientHeight,
            rackId: meta.rackId,
            slotLabel: meta.slotLabel || '',
            projectNo: meta.projectNo,
            holeNo: meta.holeNo,
            depthFrom: meta.depthFrom ?? 0,
            depthTo: meta.depthTo ?? 0,
            coreLength: meta.coreLength ?? 0,
            boxSpec: meta.boxSpec,
            boxCount: meta.boxCount,
          })
          return
        }
      }
    }
    this.onTooltipHide?.()
  }

  private handlePointerClick(pick: { hit: boolean; pickedMesh: AbstractMesh | null }) {
    if (!pick.hit || !pick.pickedMesh) {
      this.selectedMesh = null
      this.highlight.removeAllMeshes()
      this.onSelectionChange?.(null)
      return
    }
    this.selectedMesh = pick.pickedMesh
    this.highlight.removeAllMeshes()
    this.highlight.addMesh(pick.pickedMesh as Mesh, Color3.FromHexString('#4096ff'))
    const meta = pick.pickedMesh.metadata
    if (meta) this.onSelectionChange?.(meta as SelectionInfo)
  }

  // ==================== Camera Presets ====================

  setTopView() {
    this.camera.alpha = 0
    this.camera.beta = 0.1
    this.camera.radius = 45
  }

  setFrontView() {
    // Front view: from the side, showing all pairs receding into distance (matches screenshot)
    this.camera.alpha = -Math.PI / 2.5
    this.camera.beta = Math.PI / 3.5
    this.camera.radius = 30
  }

  setSideView() {
    // Side view: looking across from the side
    this.camera.alpha = Math.PI / 2
    this.camera.beta = Math.PI / 3
    this.camera.radius = 30
  }

  resetView() {
    this.camera.alpha = -Math.PI / 2.5
    this.camera.beta = Math.PI / 3.5
    this.camera.radius = 30
  }

  // ==================== Stats ====================

  getStats() {
    const total = this.rackSections.reduce((s, r) => s + r.levels * r.columns, 0)
    const occupied = this.inventory.filter((i) => i.isOccupied).length
    const totalBoxes = this.inventory.reduce((s, i) => s + i.boxCount, 0)
    return { totalSlots: total, occupiedSlots: occupied, emptySlots: total - occupied, totalBoxes, utilization: total > 0 ? Math.round((occupied / total) * 100) : 0 }
  }

  getAvailableSlots(): { rackId: string; level: number; column: number; label: string }[] {
    const slots: { rackId: string; level: number; column: number; label: string }[] = []
    for (const section of this.rackSections) {
      for (let lv = 0; lv < section.levels; lv++) {
        for (let col = 0; col < section.columns; col++) {
          if (!this.inventory.some((i) => i.rackId === section.id && i.level === lv && i.column === col && i.isOccupied)) {
            slots.push({ rackId: section.id, level: lv, column: col, label: `${section.label}-${col + 1}-${lv + 1}` })
          }
        }
      }
    }
    return slots
  }

  getOccupiedSlots(): PalletSlot[] {
    return this.inventory.filter((i) => i.isOccupied)
  }

  // ==================== Resize ====================

  private onResize = () => this.engine.resize()
}
