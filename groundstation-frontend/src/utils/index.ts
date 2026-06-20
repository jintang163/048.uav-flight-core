export const formatDistance = (distance: number): string => {
  if (distance < 1000) {
    return `${distance.toFixed(0)} m`
  }
  return `${(distance / 1000).toFixed(2)} km`
}

export const formatAltitude = (altitude: number): string => {
  return `${altitude.toFixed(1)} m`
}

export const formatSpeed = (speed: number): string => {
  return `${speed.toFixed(1)} m/s`
}

export const formatTime = (seconds: number): string => {
  if (seconds < 60) {
    return `${seconds.toFixed(0)}s`
  }
  if (seconds < 3600) {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}m ${secs}s`
  }
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  return `${hours}h ${mins}m ${secs}s`
}

export const formatDateTime = (timestamp: number): string => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

export const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

export const degToRad = (degrees: number): number => {
  return degrees * (Math.PI / 180)
}

export const radToDeg = (radians: number): number => {
  return radians * (180 / Math.PI)
}

export const calculateDistance = (lat1: number, lng1: number, lat2: number, lng2: number): number => {
  const R = 6371000
  const dLat = degToRad(lat2 - lat1)
  const dLng = degToRad(lng2 - lng1)
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(degToRad(lat1)) * Math.cos(degToRad(lat2)) *
    Math.sin(dLng / 2) * Math.sin(dLng / 2)
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
  return R * c
}

export const calculateBearing = (lat1: number, lng1: number, lat2: number, lng2: number): number => {
  const dLng = degToRad(lng2 - lng1)
  const y = Math.sin(dLng) * Math.cos(degToRad(lat2))
  const x = Math.cos(degToRad(lat1)) * Math.sin(degToRad(lat2)) -
    Math.sin(degToRad(lat1)) * Math.cos(degToRad(lat2)) * Math.cos(dLng)
  return (radToDeg(Math.atan2(y, x)) + 360) % 360
}

export const clamp = (value: number, min: number, max: number): number => {
  return Math.min(Math.max(value, min), max)
}

export const lerp = (start: number, end: number, t: number): number => {
  return start + (end - start) * t
}

export const generateId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 11)}`
}

export const deepCopy = <T>(obj: T): T => {
  return JSON.parse(JSON.stringify(obj))
}

export const throttle = <T extends (...args: unknown[]) => unknown>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean
  return function (this: unknown, ...args: Parameters<T>) {
    if (!inThrottle) {
      func.apply(this, args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

export const debounce = <T extends (...args: unknown[]) => unknown>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: ReturnType<typeof setTimeout>
  return function (this: unknown, ...args: Parameters<T>) {
    clearTimeout(timeout)
    timeout = setTimeout(() => func.apply(this, args), wait)
  }
}

export const getBatteryColor = (percentage: number): string => {
  if (percentage > 60) return '#52c41a'
  if (percentage > 30) return '#faad14'
  return '#ff4d4f'
}

export const getSignalColor = (strength: number): string => {
  if (strength > -50) return '#52c41a'
  if (strength > -70) return '#faad14'
  return '#ff4d4f'
}

export const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    connected: '#52c41a',
    armed: '#1890ff',
    flying: '#13c2c2',
    error: '#ff4d4f',
    disconnected: '#8c8c8c',
    disarmed: '#faad14',
    takeoff: '#722ed1',
    landing: '#eb2f96',
    hovering: '#13c2c2',
    return_to_home: '#fa8c16'
  }
  return colorMap[status] || '#8c8c8c'
}

export const getSeverityColor = (severity: string): string => {
  const colorMap: Record<string, string> = {
    critical: '#ff0000',
    error: '#ff4d4f',
    warning: '#faad14',
    info: '#1890ff'
  }
  return colorMap[severity] || '#8c8c8c'
}

export const normalizeAngle = (angle: number): number => {
  while (angle < 0) angle += 360
  while (angle >= 360) angle -= 360
  return angle
}

export const getRelativeAltitude = (altitude: number, homeAltitude: number): number => {
  return altitude - homeAltitude
}

export const formatCoordinates = (lat: number, lng: number): string => {
  const latDir = lat >= 0 ? 'N' : 'S'
  const lngDir = lng >= 0 ? 'E' : 'W'
  return `${Math.abs(lat).toFixed(6)}°${latDir}, ${Math.abs(lng).toFixed(6)}°${lngDir}`
}

export const isPointInPolygon = (
  point: { lat: number; lng: number },
  polygon: { lat: number; lng: number }[]
): boolean => {
  let inside = false
  for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
    const xi = polygon[i].lat, yi = polygon[i].lng
    const xj = polygon[j].lat, yj = polygon[j].lng
    if (((yi > point.lng) !== (yj > point.lng)) &&
        (point.lat < (xj - xi) * (point.lng - yi) / (yj - yi) + xi)) {
      inside = !inside
    }
  }
  return inside
}

export const isPointInCircle = (
  point: { lat: number; lng: number },
  center: { lat: number; lng: number },
  radius: number
): boolean => {
  return calculateDistance(point.lat, point.lng, center.lat, center.lng) <= radius
}

export const playAlertSound = (): void => {
  const audioContext = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)()
  const oscillator = audioContext.createOscillator()
  const gainNode = audioContext.createGain()
  
  oscillator.connect(gainNode)
  gainNode.connect(audioContext.destination)
  
  oscillator.frequency.value = 880
  oscillator.type = 'sine'
  
  gainNode.gain.setValueAtTime(0.3, audioContext.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5)
  
  oscillator.start(audioContext.currentTime)
  oscillator.stop(audioContext.currentTime + 0.5)
}

export const speak = (text: string): void => {
  if ('speechSynthesis' in window) {
    const utterance = new SpeechSynthesisUtterance(text)
    utterance.lang = 'zh-CN'
    utterance.rate = 1
    utterance.pitch = 1
    window.speechSynthesis.speak(utterance)
  }
}

export const speakAlert = (text: string): void => {
  playAlertSound()
  speak(text)
}

export const formatDuration = (seconds: number): string => {
  return formatTime(seconds)
}

export const getModeText = (mode: string): string => {
  const modeMap: Record<string, string> = {
    manual: '手动模式',
    stabilize: '自稳模式',
    acro: '特技模式',
    alt_hold: '定高模式',
    guided: '引导模式',
    auto: '自动模式',
    rtl: '返航模式',
    land: '降落模式',
    circle: '环绕模式',
    loiter: '悬停模式',
    follow: '跟随模式',
    unknown: '未知模式'
  }
  return modeMap[mode] || mode
}

export const requestNotificationPermission = async (): Promise<boolean> => {
  if (!('Notification' in window)) {
    return false
  }
  if (Notification.permission === 'granted') {
    return true
  }
  if (Notification.permission !== 'denied') {
    const permission = await Notification.requestPermission()
    return permission === 'granted'
  }
  return false
}

let audioContextInstance: AudioContext | null = null

export const initAudioContext = (): void => {
  if (!audioContextInstance && 'AudioContext' in window) {
    audioContextInstance = new AudioContext()
  }
}

export const getAudioContext = (): AudioContext | null => {
  return audioContextInstance
}

export const showNotification = (title: string, body: string, icon?: string): void => {
  if ('Notification' in window && Notification.permission === 'granted') {
    new Notification(title, { body, icon })
  }
}
