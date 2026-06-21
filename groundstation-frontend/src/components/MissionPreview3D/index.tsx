import React, { useEffect, useRef, useState, useCallback } from 'react'
import styled from 'styled-components'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { Button, Space, Slider, Tag, Badge, Tooltip, Modal, Progress } from 'antd'
import {
  PlayCircleOutlined,
  PauseOutlined,
  ReloadOutlined,
  EnvironmentOutlined,
  RocketOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  RiseOutlined,
  FallOutlined,
  EyeOutlined,
  CameraOutlined
} from '@ant-design/icons'
import type { Waypoint } from '@/types'

const Container = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  background: #0a0f1a;
  border-radius: 8px;
  overflow: hidden;
`

const CanvasWrapper = styled.div`
  width: 100%;
  height: 100%;
`

const ControlPanel = styled.div`
  position: absolute;
  top: 16px;
  left: 16px;
  right: 16px;
  z-index: 10;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  pointer-events: none;

  > * {
    pointer-events: auto;
  }
`

const ControlGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`

const Toolbar = styled.div`
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 8px;
  display: flex;
  gap: 6px;
`

const ControlButton = styled(Button)`
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;

  &:hover {
    background: rgba(24, 144, 255, 0.2) !important;
    border-color: #1890ff !important;
    color: #1890ff !important;
  }

  &.active {
    background: #1890ff;
    border-color: #1890ff;
    color: #fff;
  }
`

const InfoPanel = styled.div`
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 12px 16px;
  min-width: 240px;
`

const InfoTitle = styled.div`
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 6px;
`

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;
  font-size: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);

  &:last-child {
    border-bottom: none;
  }
`

const InfoLabel = styled.span`
  color: rgba(255, 255, 255, 0.5);
`

const InfoValue = styled.span<{ $color?: string }>`
  font-weight: 600;
  color: ${props => props.$color || 'rgba(255, 255, 255, 0.9)'};
  font-family: 'Courier New', monospace;
`

const BottomPanel = styled.div`
  position: absolute;
  bottom: 16px;
  left: 16px;
  right: 16px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 10px;
  pointer-events: none;

  > * {
    pointer-events: auto;
  }
`

const PlaybackPanel = styled.div`
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 16px;
`

const SliderContainer = styled.div`
  flex: 1;
  .ant-slider {
    margin: 0;
  }
`

const WaypointLegend = styled.div`
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 10px 16px;
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  font-size: 12px;
`

const LegendItem = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.7);
`

const LegendDot = styled.div<{ $color: string }>`
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: ${props => props.$color};
`

interface MissionPreview3DProps {
  waypoints: Waypoint[]
  onClose?: () => void
}

interface CollisionPoint {
  index: number
  position: THREE.Vector3
  terrainHeight: number
  flightAltitude: number
  penetration: number
}

const latLngToPosition = (
  lat: number,
  lng: number,
  altitude: number,
  centerLat: number,
  centerLng: number
): THREE.Vector3 => {
  const scale = 100000
  const x = (lng - centerLng) * scale
  const z = (centerLat - lat) * scale
  const y = altitude * 0.1
  return new THREE.Vector3(x, y, z)
}

const generateTerrainHeight = (x: number, z: number): number => {
  const scale1 = 0.005
  const scale2 = 0.015
  const scale3 = 0.03

  const h1 = Math.sin(x * scale1) * Math.cos(z * scale1) * 80
  const h2 = Math.sin(x * scale2 + 1.3) * Math.cos(z * scale2 + 0.7) * 40
  const h3 = Math.sin(x * scale3 + 2.1) * Math.cos(z * scale3 + 1.5) * 15

  let height = h1 + h2 + h3 + 20

  const cx = 0
  const cz = 0
  const dist = Math.sqrt(x * x + z * z)
  const maxDist = 300

  if (dist > maxDist * 0.5) {
    const t = (dist - maxDist * 0.5) / (maxDist * 0.5)
    height += t * 60
  }

  const mountainX = 120
  const mountainZ = -80
  const mountainDist = Math.sqrt((x - mountainX) ** 2 + (z - mountainZ) ** 2)
  if (mountainDist < 100) {
    const mountainHeight = Math.max(0, 120 * (1 - mountainDist / 100))
    height += mountainHeight * mountainHeight / 120
  }

  const hillX = -100
  const hillZ = 100
  const hillDist = Math.sqrt((x - hillX) ** 2 + (z - hillZ) ** 2)
  if (hillDist < 80) {
    const hillHeight = Math.max(0, 70 * (1 - hillDist / 80))
    height += hillHeight
  }

  return Math.max(0, height * 0.1)
}

const MissionPreview3D: React.FC<MissionPreview3DProps> = ({ waypoints, onClose }) => {
  const canvasRef = useRef<HTMLDivElement>(null)
  const sceneRef = useRef<THREE.Scene | null>(null)
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null)
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null)
  const controlsRef = useRef<OrbitControls | null>(null)
  const uavMeshRef = useRef<THREE.Group | null>(null)
  const pathPointsRef = useRef<THREE.Vector3[]>([])
  const animationFrameRef = useRef<number>(0)
  const clockRef = useRef<THREE.Clock>(new THREE.Clock())

  const [isPlaying, setIsPlaying] = useState(false)
  const [progress, setProgress] = useState(0)
  const [playbackSpeed, setPlaybackSpeed] = useState(1)
  const [currentIndex, setCurrentIndex] = useState(0)
  const [altitude, setAltitude] = useState(0)
  const [speed, setSpeed] = useState(0)
  const [distance, setDistance] = useState(0)
  const [showTerrain, setShowTerrain] = useState(true)
  const [followMode, setFollowMode] = useState(false)
  const [collisionPoints, setCollisionPoints] = useState<CollisionPoint[]>([])
  const [collisionCheckResult, setCollisionCheckResult] = useState<'safe' | 'warning' | 'danger'>('safe')
  const [showCollisionModal, setShowCollisionModal] = useState(false)
  const [totalDistance, setTotalDistance] = useState(0)
  const [maxAltitude, setMaxAltitude] = useState(0)
  const [minAltitude, setMinAltitude] = useState(0)

  const cleanupScene = useCallback(() => {
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current)
    }
    if (rendererRef.current && canvasRef.current) {
      rendererRef.current.dispose()
      if (canvasRef.current.contains(rendererRef.current.domElement)) {
        canvasRef.current.removeChild(rendererRef.current.domElement)
      }
    }
    sceneRef.current = null
    cameraRef.current = null
    rendererRef.current = null
    controlsRef.current = null
    uavMeshRef.current = null
    pathPointsRef.current = []
  }, [])

  const createUAVModel = (): THREE.Group => {
    const group = new THREE.Group()

    const bodyGeo = new THREE.CapsuleGeometry(0.5, 2, 4, 8)
    const bodyMat = new THREE.MeshPhongMaterial({
      color: 0x1890ff,
      emissive: 0x1890ff,
      emissiveIntensity: 0.2,
      shininess: 80
    })
    const body = new THREE.Mesh(bodyGeo, bodyMat)
    body.rotation.z = Math.PI / 2
    group.add(body)

    const armGeo = new THREE.BoxGeometry(3, 0.1, 0.1)
    const armMat = new THREE.MeshPhongMaterial({ color: 0x334155 })
    const arm1 = new THREE.Mesh(armGeo, armMat)
    const arm2 = new THREE.Mesh(armGeo, armMat)
    arm2.rotation.y = Math.PI / 2
    group.add(arm1, arm2)

    const propGeo = new THREE.CylinderGeometry(0.8, 0.8, 0.05, 8)
    const propMat = new THREE.MeshPhongMaterial({
      color: 0x64748b,
      transparent: true,
      opacity: 0.7
    })

    const armEnd = 1.5
    const positions = [
      [armEnd, 0.15, armEnd],
      [-armEnd, 0.15, armEnd],
      [armEnd, 0.15, -armEnd],
      [-armEnd, 0.15, -armEnd]
    ]

    positions.forEach(([px, py, pz]) => {
      const prop = new THREE.Mesh(propGeo, propMat)
      prop.position.set(px, py, pz)
      prop.userData.isPropeller = true
      group.add(prop)
    })

    const light = new THREE.PointLight(0xff4d4f, 2, 5)
    light.position.set(0, 0.5, 1.2)
    group.add(light)

    return group
  }

  const createTerrain = (scene: THREE.Scene) => {
    const size = 500
    const segments = 128
    const geometry = new THREE.PlaneGeometry(size, size, segments, segments)
    geometry.rotateX(-Math.PI / 2)

    const positions = geometry.attributes.position
    const colors = new Float32Array(positions.count * 3)

    for (let i = 0; i < positions.count; i++) {
      const x = positions.getX(i)
      const z = positions.getZ(i)
      const h = generateTerrainHeight(x, z)
      positions.setY(i, h)

      const normalizedH = Math.min(h / 15, 1)
      let r, g, b
      if (normalizedH < 0.3) {
        r = 0.2 + normalizedH * 0.5
        g = 0.4 + normalizedH * 0.8
        b = 0.2 + normalizedH * 0.3
      } else if (normalizedH < 0.6) {
        r = 0.5 + (normalizedH - 0.3) * 0.5
        g = 0.5 + (normalizedH - 0.3) * 0.3
        b = 0.3
      } else {
        r = 0.7 + (normalizedH - 0.6) * 0.3
        g = 0.7 + (normalizedH - 0.6) * 0.3
        b = 0.6 + (normalizedH - 0.6) * 0.4
      }

      colors[i * 3] = r
      colors[i * 3 + 1] = g
      colors[i * 3 + 2] = b
    }

    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3))
    geometry.computeVertexNormals()

    const material = new THREE.MeshStandardMaterial({
      vertexColors: true,
      flatShading: true,
      roughness: 0.8,
      metalness: 0.1
    })

    const terrain = new THREE.Mesh(geometry, material)
    terrain.receiveShadow = true
    terrain.userData.isTerrain = true
    terrain.visible = showTerrain
    scene.add(terrain)

    const wireGeo = new THREE.WireframeGeometry(geometry)
    const wireMat = new THREE.LineBasicMaterial({
      color: 0x1890ff,
      transparent: true,
      opacity: 0.08
    })
    const wireframe = new THREE.LineSegments(wireGeo, wireMat)
    wireframe.userData.isTerrainWire = true
    wireframe.visible = showTerrain
    scene.add(wireframe)

    return terrain
  }

  const createWaypointMarker = (
    position: THREE.Vector3,
    index: number,
    isFirst: boolean,
    isLast: boolean
  ): THREE.Group => {
    const group = new THREE.Group()

    let color = 0x1890ff
    if (isFirst) color = 0x52c41a
    else if (isLast) color = 0xff4d4f

    const markerGeo = new THREE.SphereGeometry(1.2, 16, 16)
    const markerMat = new THREE.MeshPhongMaterial({
      color,
      emissive: color,
      emissiveIntensity: 0.5,
      transparent: true,
      opacity: 0.9
    })
    const marker = new THREE.Mesh(markerGeo, markerMat)
    group.add(marker)

    const ringGeo = new THREE.RingGeometry(1.5, 2, 32)
    const ringMat = new THREE.MeshBasicMaterial({
      color,
      transparent: true,
      opacity: 0.5,
      side: THREE.DoubleSide
    })
    const ring = new THREE.Mesh(ringGeo, ringMat)
    ring.rotation.x = -Math.PI / 2
    ring.position.y = -position.y + 0.1
    group.add(ring)

    const pillarGeo = new THREE.CylinderGeometry(0.1, 0.1, position.y, 8)
    const pillarMat = new THREE.MeshBasicMaterial({
      color,
      transparent: true,
      opacity: 0.4
    })
    const pillar = new THREE.Mesh(pillarGeo, pillarMat)
    pillar.position.y = -position.y / 2
    group.add(pillar)

    const canvas = document.createElement('canvas')
    canvas.width = 64
    canvas.height = 64
    const ctx = canvas.getContext('2d')!
    ctx.fillStyle = 'rgba(0,0,0,0.7)'
    ctx.beginPath()
    ctx.arc(32, 32, 28, 0, Math.PI * 2)
    ctx.fill()
    ctx.strokeStyle = `#${color.toString(16).padStart(6, '0')}`
    ctx.lineWidth = 2
    ctx.stroke()
    ctx.fillStyle = '#fff'
    ctx.font = 'bold 24px Arial'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(String(index + 1), 32, 32)

    const spriteMat = new THREE.CanvasTexture(canvas)
    const spriteMaterial = new THREE.SpriteMaterial({ map: spriteMat, depthTest: false })
    const sprite = new THREE.Sprite(spriteMaterial)
    sprite.scale.set(3, 3, 1)
    sprite.position.y = 2.5
    group.add(sprite)

    group.position.copy(position)
    return group
  }

  const createFlightPath = (scene: THREE.Scene, points: THREE.Vector3[]) => {
    if (points.length < 2) return

    const curve = new THREE.CatmullRomCurve3(points)
    curve.curveType = 'catmullrom'
    curve.tension = 0.1

    const pathPoints = curve.getPoints(500)
    pathPointsRef.current = pathPoints

    const tubeGeo = new THREE.TubeGeometry(curve, 200, 0.25, 8, false)
    const tubeMat = new THREE.MeshPhongMaterial({
      color: 0x1890ff,
      emissive: 0x1890ff,
      emissiveIntensity: 0.3,
      transparent: true,
      opacity: 0.6
    })
    const tube = new THREE.Mesh(tubeGeo, tubeMat)
    tube.userData.isFlightPath = true
    scene.add(tube)

    const lineGeo = new THREE.BufferGeometry().setFromPoints(pathPoints)
    const lineMat = new THREE.LineDashedMaterial({
      color: 0x52c41a,
      dashSize: 2,
      gapSize: 1,
      transparent: true,
      opacity: 0.8
    })
    const line = new THREE.Line(lineGeo, lineMat)
    line.computeLineDistances()
    line.userData.isFlightPathLine = true
    scene.add(line)

    let totalDist = 0
    for (let i = 1; i < points.length; i++) {
      totalDist += points[i].distanceTo(points[i - 1])
    }
    setTotalDistance(Math.round(totalDist * 10))
  }

  const checkCollisions = (points: THREE.Vector3[]): CollisionPoint[] => {
    const collisions: CollisionPoint[] = []
    const safetyMargin = 2

    for (let i = 0; i < points.length; i++) {
      const point = points[i]
      const terrainH = generateTerrainHeight(point.x, point.z)
      const flightAlt = point.y
      const penetration = terrainH + safetyMargin - flightAlt

      if (penetration > 0) {
        collisions.push({
          index: i,
          position: point.clone(),
          terrainHeight: terrainH * 10,
          flightAltitude: flightAlt * 10,
          penetration: penetration * 10
        })
      }
    }

    if (points.length > 1) {
      for (let i = 0; i < points.length - 1; i++) {
        const start = points[i]
        const end = points[i + 1]
        const steps = 20

        for (let j = 1; j < steps; j++) {
          const t = j / steps
          const x = start.x + (end.x - start.x) * t
          const z = start.z + (end.z - start.z) * t
          const y = start.y + (end.y - start.y) * t
          const terrainH = generateTerrainHeight(x, z)
          const penetration = terrainH + safetyMargin - y

          if (penetration > 3) {
            const existing = collisions.find(
              c => Math.abs(c.position.x - x) < 5 && Math.abs(c.position.z - z) < 5
            )
            if (!existing) {
              collisions.push({
                index: i,
                position: new THREE.Vector3(x, y, z),
                terrainHeight: terrainH * 10,
                flightAltitude: y * 10,
                penetration: penetration * 10
              })
            }
          }
        }
      }
    }

    return collisions
  }

  const createCollisionMarkers = (scene: THREE.Scene, collisions: CollisionPoint[]) => {
    collisions.forEach(col => {
      const coneGeo = new THREE.ConeGeometry(1.5, 4, 4)
      const coneMat = new THREE.MeshPhongMaterial({
        color: col.penetration > 5 ? 0xff4d4f : 0xfaad14,
        emissive: col.penetration > 5 ? 0xff4d4f : 0xfaad14,
        emissiveIntensity: 0.5,
        transparent: true,
        opacity: 0.8
      })
      const cone = new THREE.Mesh(coneGeo, coneMat)
      cone.position.copy(col.position)
      cone.position.y = Math.max(col.position.y, col.terrainHeight / 10 + 3)
      cone.rotation.x = Math.PI
      cone.userData.isCollisionMarker = true
      scene.add(cone)

      const warnGeo = new THREE.SphereGeometry(2, 16, 16)
      const warnMat = new THREE.MeshBasicMaterial({
        color: 0xff4d4f,
        transparent: true,
        opacity: 0.2
      })
      const warnSphere = new THREE.Mesh(warnGeo, warnMat)
      warnSphere.position.copy(col.position)
      warnSphere.userData.isCollisionMarker = true
      scene.add(warnSphere)
    })
  }

  const clearDynamicObjects = (scene: THREE.Scene) => {
    const toRemove: THREE.Object3D[] = []
    scene.traverse((obj) => {
      if (
        obj.userData.isWaypoint ||
        obj.userData.isFlightPath ||
        obj.userData.isFlightPathLine ||
        obj.userData.isCollisionMarker
      ) {
        toRemove.push(obj)
      }
    })
    toRemove.forEach(obj => {
      scene.remove(obj)
      if (obj instanceof THREE.Mesh) {
        obj.geometry.dispose()
        if (Array.isArray(obj.material)) {
          obj.material.forEach(m => m.dispose())
        } else {
          obj.material.dispose()
        }
      }
    })
  }

  const initScene = useCallback(() => {
    if (!canvasRef.current) return

    cleanupScene()

    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x0a0f1a)
    scene.fog = new THREE.Fog(0x0a0f1a, 200, 600)
    sceneRef.current = scene

    const camera = new THREE.PerspectiveCamera(
      60,
      canvasRef.current.clientWidth / canvasRef.current.clientHeight,
      0.1,
      2000
    )
    camera.position.set(80, 60, 80)
    cameraRef.current = camera

    const renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setSize(canvasRef.current.clientWidth, canvasRef.current.clientHeight)
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.shadowMap.enabled = true
    renderer.shadowMap.type = THREE.PCFSoftShadowMap
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.2
    canvasRef.current.appendChild(renderer.domElement)
    rendererRef.current = renderer

    const controls = new OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    controls.dampingFactor = 0.05
    controls.maxPolarAngle = Math.PI / 2 - 0.05
    controls.minDistance = 5
    controls.maxDistance = 400
    controls.target.set(0, 5, 0)
    controlsRef.current = controls

    const ambientLight = new THREE.AmbientLight(0x6688aa, 0.6)
    scene.add(ambientLight)

    const sunLight = new THREE.DirectionalLight(0xffffee, 1.2)
    sunLight.position.set(50, 100, 50)
    sunLight.castShadow = true
    sunLight.shadow.mapSize.width = 2048
    sunLight.shadow.mapSize.height = 2048
    sunLight.shadow.camera.near = 0.5
    sunLight.shadow.camera.far = 500
    sunLight.shadow.camera.left = -250
    sunLight.shadow.camera.right = 250
    sunLight.shadow.camera.top = 250
    sunLight.shadow.camera.bottom = -250
    scene.add(sunLight)

    const fillLight = new THREE.DirectionalLight(0x4488ff, 0.3)
    fillLight.position.set(-50, 50, -50)
    scene.add(fillLight)

    createTerrain(scene)

    const gridHelper = new THREE.GridHelper(500, 50, 0x1890ff, 0x1890ff33)
    ;(gridHelper.material as THREE.Material).transparent = true
    ;(gridHelper.material as THREE.Material).opacity = 0.25
    gridHelper.position.y = 0.01
    scene.add(gridHelper)

    const axesHelper = new THREE.AxesHelper(10)
    axesHelper.position.y = 0.02
    scene.add(axesHelper)

    const uav = createUAVModel()
    uav.visible = false
    uav.scale.setScalar(0.8)
    scene.add(uav)
    uavMeshRef.current = uav

    if (waypoints.length > 0) {
      updateWaypoints()
    }

    const animate = () => {
      animationFrameRef.current = requestAnimationFrame(animate)
      const delta = clockRef.current.getDelta()

      controls.update()

      if (uavMeshRef.current && isPlaying && pathPointsRef.current.length > 1) {
        uavMeshRef.current.traverse((obj) => {
          if (obj instanceof THREE.Mesh && obj.userData.isPropeller) {
            obj.rotation.y += delta * 80
          }
        })
      }

      renderer.render(scene, camera)
    }
    animate()

    const handleResize = () => {
      if (!canvasRef.current || !camera || !renderer) return
      camera.aspect = canvasRef.current.clientWidth / canvasRef.current.clientHeight
      camera.updateProjectionMatrix()
      renderer.setSize(canvasRef.current.clientWidth, canvasRef.current.clientHeight)
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [waypoints.length, showTerrain, cleanupScene])

  const updateWaypoints = useCallback(() => {
    const scene = sceneRef.current
    if (!scene || waypoints.length === 0) return

    clearDynamicObjects(scene)

    const centerLat = waypoints.reduce((sum, w) => sum + w.lat, 0) / waypoints.length
    const centerLng = waypoints.reduce((sum, w) => sum + w.lng, 0) / waypoints.length

    const positions: THREE.Vector3[] = waypoints.map((wp, i) => {
      const pos = latLngToPosition(wp.lat, wp.lng, wp.altitude, centerLat, centerLng)
      const marker = createWaypointMarker(pos, i, i === 0, i === waypoints.length - 1)
      marker.userData.isWaypoint = true
      scene.add(marker)
      return pos
    })

    createFlightPath(scene, positions)

    const collisions = checkCollisions(positions)
    setCollisionPoints(collisions)

    if (collisions.length > 0) {
      createCollisionMarkers(scene, collisions)
      const maxPen = Math.max(...collisions.map(c => c.penetration))
      setCollisionCheckResult(maxPen > 10 ? 'danger' : 'warning')
    } else {
      setCollisionCheckResult('safe')
    }

    if (positions.length > 0 && uavMeshRef.current) {
      uavMeshRef.current.position.copy(positions[0])
      uavMeshRef.current.visible = true

      if (positions.length > 1) {
        const dir = new THREE.Vector3().subVectors(positions[1], positions[0])
        uavMeshRef.current.rotation.y = Math.atan2(dir.x, dir.z)
      }
    }

    const alts = waypoints.map(w => w.altitude)
    setMaxAltitude(Math.max(...alts))
    setMinAltitude(Math.min(...alts))
    setCurrentIndex(0)
    setAltitude(waypoints[0]?.altitude || 0)
    setProgress(0)
  }, [waypoints])

  useEffect(() => {
    const cleanup = initScene()
    return () => {
      cleanup?.()
      cleanupScene()
    }
  }, [initScene])

  useEffect(() => {
    if (sceneRef.current) {
      updateWaypoints()
    }
  }, [updateWaypoints])

  useEffect(() => {
    if (sceneRef.current) {
      sceneRef.current.traverse((obj) => {
        if (obj.userData.isTerrain || obj.userData.isTerrainWire) {
          obj.visible = showTerrain
        }
      })
    }
  }, [showTerrain])

  useEffect(() => {
    if (!isPlaying || !uavMeshRef.current || pathPointsRef.current.length < 2) return

    const points = pathPointsRef.current
    let currentProgress = progress

    const updateFrame = () => {
      if (!isPlaying) return

      const increment = (playbackSpeed * 0.0005) / Math.max(waypoints.length / 5, 1)
      currentProgress = Math.min(currentProgress + increment, 1)
      setProgress(currentProgress)

      const index = Math.floor(currentProgress * (points.length - 1))
      const localT = (currentProgress * (points.length - 1)) - index
      const nextIndex = Math.min(index + 1, points.length - 1)

      const p0 = points[index]
      const p1 = points[nextIndex]
      const pos = new THREE.Vector3().lerpVectors(p0, p1, localT)

      uavMeshRef.current!.position.copy(pos)

      const dir = new THREE.Vector3().subVectors(p1, p0).normalize()
      if (dir.length() > 0.001) {
        uavMeshRef.current!.rotation.y = Math.atan2(dir.x, dir.z)
      }

      const waypointIdx = Math.min(
        Math.floor(currentProgress * waypoints.length),
        waypoints.length - 1
      )
      setCurrentIndex(waypointIdx)
      setAltitude(pos.y * 10)
      setSpeed(10 * playbackSpeed)
      setDistance(currentProgress * totalDistance)

      if (followMode && cameraRef.current && controlsRef.current) {
        const cameraOffset = new THREE.Vector3(0, 15, 25)
        const cameraTarget = pos.clone().add(
          new THREE.Vector3(
            Math.sin(uavMeshRef.current!.rotation.y) * cameraOffset.z,
            cameraOffset.y,
            Math.cos(uavMeshRef.current!.rotation.y) * cameraOffset.z
          )
        )
        cameraRef.current.position.lerp(cameraTarget, 0.05)
        controlsRef.current.target.lerp(pos, 0.1)
      }

      if (currentProgress >= 1) {
        setIsPlaying(false)
      }
    }

    const interval = setInterval(updateFrame, 16)
    return () => clearInterval(interval)
  }, [isPlaying, progress, playbackSpeed, waypoints.length, totalDistance, followMode])

  const handlePlayPause = () => {
    if (waypoints.length < 2) return
    if (progress >= 1) {
      setProgress(0)
    }
    setIsPlaying(!isPlaying)
  }

  const handleReset = () => {
    setIsPlaying(false)
    setProgress(0)
    setCurrentIndex(0)
    setAltitude(waypoints[0]?.altitude || 0)
    setDistance(0)
    setSpeed(0)

    if (uavMeshRef.current && pathPointsRef.current.length > 0) {
      uavMeshRef.current.position.copy(pathPointsRef.current[0])
      uavMeshRef.current.rotation.set(0, 0, 0)
      if (pathPointsRef.current.length > 1) {
        const dir = new THREE.Vector3().subVectors(
          pathPointsRef.current[1],
          pathPointsRef.current[0]
        )
        uavMeshRef.current.rotation.y = Math.atan2(dir.x, dir.z)
      }
    }

    if (cameraRef.current && controlsRef.current) {
      cameraRef.current.position.set(80, 60, 80)
      controlsRef.current.target.set(0, 5, 0)
    }
  }

  const handleProgressChange = (value: number) => {
    if (pathPointsRef.current.length < 2 || !uavMeshRef.current) return

    setProgress(value / 100)
    const points = pathPointsRef.current
    const progressNorm = value / 100
    const index = Math.floor(progressNorm * (points.length - 1))
    const localT = (progressNorm * (points.length - 1)) - index
    const nextIndex = Math.min(index + 1, points.length - 1)

    const p0 = points[index]
    const p1 = points[nextIndex]
    const pos = new THREE.Vector3().lerpVectors(p0, p1, localT)
    uavMeshRef.current.position.copy(pos)

    const dir = new THREE.Vector3().subVectors(p1, p0).normalize()
    if (dir.length() > 0.001) {
      uavMeshRef.current.rotation.y = Math.atan2(dir.x, dir.z)
    }

    const waypointIdx = Math.min(
      Math.floor(progressNorm * waypoints.length),
      waypoints.length - 1
    )
    setCurrentIndex(waypointIdx)
    setAltitude(pos.y * 10)
    setDistance(progressNorm * totalDistance)
  }

  const resetCamera = () => {
    if (cameraRef.current && controlsRef.current) {
      cameraRef.current.position.set(80, 60, 80)
      controlsRef.current.target.set(0, 5, 0)
    }
  }

  const getCollisionStatusColor = () => {
    switch (collisionCheckResult) {
      case 'safe': return '#52c41a'
      case 'warning': return '#faad14'
      case 'danger': return '#ff4d4f'
    }
  }

  const getCollisionStatusText = () => {
    switch (collisionCheckResult) {
      case 'safe': return '航线安全'
      case 'warning': return '存在风险'
      case 'danger': return '碰撞危险'
    }
  }

  return (
    <Container>
      <CanvasWrapper ref={canvasRef} />

      <ControlPanel>
        <ControlGroup>
          <Toolbar>
            <Tooltip title={showTerrain ? '隐藏地形' : '显示地形'}>
              <ControlButton
                className={showTerrain ? 'active' : ''}
                icon={<EnvironmentOutlined />}
                onClick={() => setShowTerrain(!showTerrain)}
              />
            </Tooltip>
            <Tooltip title={followMode ? '取消跟随' : '跟随无人机'}>
              <ControlButton
                className={followMode ? 'active' : ''}
                icon={<RocketOutlined />}
                onClick={() => setFollowMode(!followMode)}
              />
            </Tooltip>
            <Tooltip title="重置视角">
              <ControlButton
                icon={<CameraOutlined />}
                onClick={resetCamera}
              />
            </Tooltip>
            {onClose && (
              <Tooltip title="关闭预览">
                <ControlButton
                  icon={<EyeOutlined />}
                  onClick={onClose}
                />
              </Tooltip>
            )}
          </Toolbar>
        </ControlGroup>

        <InfoPanel>
          <InfoTitle>
            <RocketOutlined style={{ color: '#1890ff' }} />
            航线信息
            <Badge
              status={collisionCheckResult === 'safe' ? 'success' : collisionCheckResult === 'warning' ? 'warning' : 'error'}
              text={
                <span style={{ color: getCollisionStatusColor(), cursor: 'pointer' }} onClick={() => setShowCollisionModal(true)}>
                  {getCollisionStatusText()}
                </span>
              }
              style={{ marginLeft: 'auto' }}
            />
          </InfoTitle>
          <InfoRow>
            <InfoLabel>航点数量</InfoLabel>
            <InfoValue>{waypoints.length} 个</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>总航程</InfoLabel>
            <InfoValue $color="#1890ff">{totalDistance} m</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>最高高度</InfoLabel>
            <InfoValue $color="#52c41a">
              <RiseOutlined /> {maxAltitude.toFixed(1)} m
            </InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>最低高度</InfoLabel>
            <InfoValue $color="#faad14">
              <FallOutlined /> {minAltitude.toFixed(1)} m
            </InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>当前航点</InfoLabel>
            <InfoValue>
              #{currentIndex + 1}
              {waypoints[currentIndex] && (
                <Tag color="blue" style={{ marginLeft: 4 }}>
                  {waypoints[currentIndex].altitude.toFixed(0)}m
                </Tag>
              )}
            </InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>当前高度</InfoLabel>
            <InfoValue $color="#1890ff">{altitude.toFixed(1)} m</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>当前速度</InfoLabel>
            <InfoValue $color="#52c41a">{speed.toFixed(1)} m/s</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>已飞距离</InfoLabel>
            <InfoValue>{distance.toFixed(0)} / {totalDistance} m</InfoValue>
          </InfoRow>
          {collisionPoints.length > 0 && (
            <InfoRow>
              <InfoLabel style={{ color: '#ff4d4f' }}>
                <WarningOutlined /> 碰撞点
              </InfoLabel>
              <InfoValue
                $color="#ff4d4f"
                style={{ cursor: 'pointer', textDecoration: 'underline' }}
                onClick={() => setShowCollisionModal(true)}
              >
                {collisionPoints.length} 处
              </InfoValue>
            </InfoRow>
          )}
        </InfoPanel>
      </ControlPanel>

      <BottomPanel>
        <PlaybackPanel>
          <Space>
            <Tooltip title={isPlaying ? '暂停' : '播放'}>
              <ControlButton
                type={isPlaying ? 'primary' : 'default'}
                icon={isPlaying ? <PauseOutlined /> : <PlayCircleOutlined />}
                onClick={handlePlayPause}
                style={{ width: 44, height: 44, fontSize: 18 }}
                disabled={waypoints.length < 2}
              />
            </Tooltip>
            <Tooltip title="重置">
              <ControlButton
                icon={<ReloadOutlined />}
                onClick={handleReset}
                disabled={waypoints.length < 2}
              />
            </Tooltip>
            <span style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12, minWidth: 60 }}>
              x{playbackSpeed}
            </span>
            <Slider
              min={0.5}
              max={5}
              step={0.5}
              value={playbackSpeed}
              onChange={setPlaybackSpeed}
              style={{ width: 80 }}
            />
          </Space>
          <SliderContainer>
            <Slider
              min={0}
              max={100}
              step={0.1}
              value={progress * 100}
              onChange={handleProgressChange}
              disabled={waypoints.length < 2}
              tooltip={{ formatter: (v) => `${v?.toFixed(1)}%` }}
              styles={{
                track: {
                  background: collisionCheckResult === 'danger' ? '#ff4d4f' : collisionCheckResult === 'warning' ? '#faad14' : '#52c41a'
                }
              }}
            />
          </SliderContainer>
          <Space style={{ minWidth: 140, justifyContent: 'flex-end' }}>
            <Tag color="geekblue" style={{ margin: 0 }}>
              航点 {currentIndex + 1}/{waypoints.length}
            </Tag>
          </Space>
        </PlaybackPanel>

        <WaypointLegend>
          <LegendItem>
            <LegendDot $color="#52c41a" />
            <span>起点</span>
          </LegendItem>
          <LegendItem>
            <LegendDot $color="#1890ff" />
            <span>途经航点</span>
          </LegendItem>
          <LegendItem>
            <LegendDot $color="#ff4d4f" />
            <span>终点</span>
          </LegendItem>
          <LegendItem>
            <LegendDot $color="#faad14" />
            <span>航线轨迹</span>
          </LegendItem>
          <LegendItem>
            <LegendDot $color="#ff4d4f" />
            <span>碰撞警告</span>
          </LegendItem>
        </WaypointLegend>
      </BottomPanel>

      <Modal
        title={
          <Space>
            {collisionCheckResult === 'safe' ? (
              <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 20 }} />
            ) : (
              <WarningOutlined style={{ color: getCollisionStatusColor(), fontSize: 20 }} />
            )}
            航线安全检测报告
          </Space>
        }
        open={showCollisionModal}
        onCancel={() => setShowCollisionModal(false)}
        footer={[
          <Button key="close" onClick={() => setShowCollisionModal(false)}>
            关闭
          </Button>
        ]}
        width={600}
      >
        {collisionCheckResult === 'safe' ? (
          <div style={{ textAlign: 'center', padding: '30px 0' }}>
            <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a' }} />
            <div style={{ fontSize: 18, fontWeight: 600, marginTop: 16, color: '#52c41a' }}>
              航线安全，无碰撞风险
            </div>
            <div style={{ color: 'rgba(255,255,255,0.6)', marginTop: 8 }}>
              所有航点高度均高于地形，航线可以安全执行
            </div>
          </div>
        ) : (
          <div>
            <div style={{
              padding: '12px 16px',
              background: collisionCheckResult === 'danger' ? 'rgba(255,77,79,0.1)' : 'rgba(250,173,20,0.1)',
              border: `1px solid ${collisionCheckResult === 'danger' ? 'rgba(255,77,79,0.3)' : 'rgba(250,173,20,0.3)'}`,
              borderRadius: 8,
              marginBottom: 16,
              color: getCollisionStatusColor()
            }}>
              <WarningOutlined /> 检测到 {collisionPoints.length} 处潜在碰撞风险，请调整航线高度
            </div>
            <Progress
              percent={Math.round((1 - collisionPoints.length / Math.max(waypoints.length * 3, 1)) * 100)}
              status={collisionCheckResult === 'danger' ? 'exception' : 'normal'}
              strokeColor={getCollisionStatusColor()}
              showInfo
              style={{ marginBottom: 20 }}
              format={(p) => `安全评分: ${p}%`}
            />
            <div style={{ maxHeight: 300, overflow: 'auto' }}>
              {collisionPoints.map((col, i) => (
                <div
                  key={i}
                  style={{
                    padding: '10px 12px',
                    marginBottom: 8,
                    background: 'rgba(255,255,255,0.03)',
                    border: `1px solid ${col.penetration > 10 ? 'rgba(255,77,79,0.2)' : 'rgba(250,173,20,0.2)'}`,
                    borderRadius: 6
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Space>
                      <Tag color={col.penetration > 10 ? 'red' : 'orange'}>
                        {col.penetration > 10 ? '严重' : '警告'}
                      </Tag>
                      <span style={{ color: 'rgba(255,255,255,0.9)' }}>
                        航点 #{col.index + 1} 附近
                      </span>
                    </Space>
                    <span style={{
                      color: col.penetration > 10 ? '#ff4d4f' : '#faad14',
                      fontFamily: 'monospace',
                      fontWeight: 600
                    }}>
                      穿透 {col.penetration.toFixed(1)}m
                    </span>
                  </div>
                  <div style={{ display: 'flex', gap: 20, marginTop: 8, fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>
                    <span>飞行高度: <b style={{ color: '#1890ff' }}>{col.flightAltitude.toFixed(1)}m</b></span>
                    <span>地形高度: <b style={{ color: '#faad14' }}>{col.terrainHeight.toFixed(1)}m</b></span>
                    <span>建议抬高: <b style={{ color: '#52c41a' }}>{(col.penetration + 5).toFixed(1)}m</b></span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </Modal>
    </Container>
  )
}

export default MissionPreview3D
