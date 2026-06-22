import { useEffect, useCallback, useRef, useState } from 'react'
import { useAppDispatch } from '@/store'
import {
  addHIDDevice,
  removeHIDDevice,
  setHIDDevices,
  setActiveHIDDevice,
  setHIDAxes,
  setHIDButtons,
  setHIDCalibrationNeeded
} from '@/store/slices/remote-cockpit'
import type {
  HIDDeviceInfo,
  HIDDeviceType,
  HIDAxisState,
  HIDButtonState
} from '@/types'

const GAMEPAD_POLL_INTERVAL_MS = 16
const AXIS_DEADZONE = 0.08
const BUTTON_PRESS_THRESHOLD = 0.5

interface GamepadMapping {
  pitch: number
  roll: number
  yaw: number
  throttle: number
  arm: number
  disarm: number
  takeoff: number
  land: number
  rtl: number
  pause: number
  mode_switch: number
  emergency_stop: number
}

const DEFAULT_XBOX_MAPPING: GamepadMapping = {
  pitch: 1,
  roll: 0,
  yaw: 3,
  throttle: 4,
  arm: 0,
  disarm: 1,
  takeoff: 2,
  land: 3,
  rtl: 8,
  pause: 9,
  mode_switch: 5,
  emergency_stop: 4
}

const DEFAULT_JOYSTICK_MAPPING: GamepadMapping = {
  pitch: 1,
  roll: 0,
  yaw: 2,
  throttle: 3,
  arm: 0,
  disarm: 1,
  takeoff: 2,
  land: 3,
  rtl: 4,
  pause: 5,
  mode_switch: 6,
  emergency_stop: 7
}

const detectDeviceType = (id: string, name: string): HIDDeviceType => {
  const lowerName = name.toLowerCase()
  if (lowerName.includes('rc') || lowerName.includes('remote') || lowerName.includes('transmitter')) {
    return 'rc_transmitter' as HIDDeviceType
  }
  if (lowerName.includes('stick') || lowerName.includes('joystick') || lowerName.includes('flight')) {
    return 'joystick' as HIDDeviceType
  }
  if (lowerName.includes('xbox') || lowerName.includes('gamepad') || lowerName.includes('controller') || lowerName.includes('playstation') || lowerName.includes('ps')) {
    return 'gamepad' as HIDDeviceType
  }
  return 'unknown' as HIDDeviceType
}

const getDeviceMapping = (deviceType: HIDDeviceType): GamepadMapping => {
  if (deviceType === 'joystick' as HIDDeviceType || deviceType === 'rc_transmitter' as HIDDeviceType) {
    return DEFAULT_JOYSTICK_MAPPING
  }
  return DEFAULT_XBOX_MAPPING
}

const applyDeadzone = (value: number, deadzone: number = AXIS_DEADZONE): number => {
  if (Math.abs(value) < deadzone) {
    return 0
  }
  const sign = value > 0 ? 1 : -1
  const normalized = (Math.abs(value) - deadzone) / (1 - deadzone)
  return sign * Math.min(Math.max(normalized, 0), 1)
}

const normalizeThrottle = (value: number): number => {
  return (value + 1) / 2
}

const createHIDDeviceInfo = (gamepad: Gamepad): HIDDeviceInfo => {
  return {
    id: gamepad.id,
    name: gamepad.id,
    type: detectDeviceType(gamepad.id, gamepad.id),
    connected: gamepad.connected,
    vendor_id: 0,
    product_id: 0,
    mapping_profile: 'default',
    last_active: Date.now()
  }
}

export const useHIDGamepad = () => {
  const dispatch = useAppDispatch()
  const [supported, setSupported] = useState<boolean>(typeof navigator !== 'undefined' && 'getGamepads' in navigator)
  const [devices, setDevices] = useState<HIDDeviceInfo[]>([])
  const [activeDeviceId, setActiveDeviceId] = useState<string | null>(null)
  const pollTimerRef = useRef<number | null>(null)
  const lastButtonsRef = useRef<boolean[]>([])
  const calibrationDataRef = useRef<{ centers: number[]; ranges: number[] } | null>(null)

  const handleGamepadConnected = useCallback((event: GamepadEvent) => {
    const device = createHIDDeviceInfo(event.gamepad)
    dispatch(addHIDDevice(device))

    setDevices(prev => {
      const existing = prev.find(d => d.id === device.id)
      if (!existing) {
        return [...prev, device]
      }
      return prev.map(d => d.id === device.id ? device : d)
    })

    if (!activeDeviceId) {
      setActiveDeviceId(device.id)
      dispatch(setActiveHIDDevice(device.id))
    }
  }, [dispatch, activeDeviceId])

  const handleGamepadDisconnected = useCallback((event: GamepadEvent) => {
    const id = event.gamepad.id
    dispatch(removeHIDDevice(id))

    setDevices(prev => prev.filter(d => d.id !== id))

    if (activeDeviceId === id) {
      setActiveDeviceId(null)
      dispatch(setActiveHIDDevice(null))
    }
  }, [dispatch, activeDeviceId])

  const readGamepadState = useCallback(() => {
    if (!activeDeviceId) return

    const gamepads = navigator.getGamepads()
    const activeGamepad = gamepads.find(g => g && g.id === activeDeviceId)

    if (!activeGamepad) return

    const deviceType = devices.find(d => d.id === activeDeviceId)?.type || ('unknown' as HIDDeviceType)
    const mapping = getDeviceMapping(deviceType)

    let pitch = 0
    let roll = 0
    let yaw = 0
    let throttle = 0

    if (activeGamepad.axes[mapping.pitch] !== undefined) {
      pitch = applyDeadzone(-activeGamepad.axes[mapping.pitch])
    }
    if (activeGamepad.axes[mapping.roll] !== undefined) {
      roll = applyDeadzone(activeGamepad.axes[mapping.roll])
    }
    if (activeGamepad.axes[mapping.yaw] !== undefined) {
      yaw = applyDeadzone(activeGamepad.axes[mapping.yaw])
    }
    if (activeGamepad.axes[mapping.throttle] !== undefined) {
      const rawThrottle = activeGamepad.axes[mapping.throttle]
      throttle = deviceType === 'gamepad' as HIDDeviceType
        ? applyDeadzone(-rawThrottle)
        : normalizeThrottle(applyDeadzone(rawThrottle))
    }

    const axesState: HIDAxisState = { pitch, roll, yaw, throttle }
    dispatch(setHIDAxes(axesState))

    const buttonsState: HIDButtonState = {
      arm: false,
      disarm: false,
      takeoff: false,
      land: false,
      rtl: false,
      pause: false,
      mode_switch: false,
      emergency_stop: false
    }

    const buttonMappings: Array<keyof HIDButtonState> = [
      'arm', 'disarm', 'takeoff', 'land', 'rtl', 'pause', 'mode_switch', 'emergency_stop'
    ]

    buttonMappings.forEach(buttonName => {
      const buttonIndex = mapping[buttonName]
      if (buttonIndex !== undefined && activeGamepad.buttons[buttonIndex]) {
        const pressed = activeGamepad.buttons[buttonIndex].pressed ||
          activeGamepad.buttons[buttonIndex].value > BUTTON_PRESS_THRESHOLD
        buttonsState[buttonName] = pressed
      }
    })

    const hasChanged = buttonMappings.some(name => {
      const idx = mapping[name]
      if (idx === undefined) return false
      const wasPressed = lastButtonsRef.current[idx] || false
      const isPressed = buttonsState[name]
      return wasPressed !== isPressed
    })

    if (hasChanged) {
      dispatch(setHIDButtons(buttonsState))
      lastButtonsRef.current = activeGamepad.buttons.map(b => b.pressed || b.value > BUTTON_PRESS_THRESHOLD)
    }
  }, [dispatch, activeDeviceId, devices])

  const selectDevice = useCallback((deviceId: string | null) => {
    setActiveDeviceId(deviceId)
    dispatch(setActiveHIDDevice(deviceId))
    lastButtonsRef.current = []
  }, [dispatch])

  const calibrateDevice = useCallback(async () => {
    if (!activeDeviceId) return

    const gamepads = navigator.getGamepads()
    const gamepad = gamepads.find(g => g && g.id === activeDeviceId)
    if (!gamepad) return

    dispatch(setHIDCalibrationNeeded(true))

    const centers = gamepad.axes.map(axis => axis)
    const ranges = gamepad.axes.map(() => 1)
    calibrationDataRef.current = { centers, ranges }

    setTimeout(() => {
      dispatch(setHIDCalibrationNeeded(false))
    }, 100)
  }, [dispatch, activeDeviceId])

  const getAvailableDevices = useCallback((): HIDDeviceInfo[] => {
    if (!supported) return []
    const gamepads = navigator.getGamepads()
    return gamepads
      .filter((g): g is Gamepad => g !== null && g.connected)
      .map(g => createHIDDeviceInfo(g))
  }, [supported])

  useEffect(() => {
    if (!supported) return

    window.addEventListener('gamepadconnected', handleGamepadConnected)
    window.addEventListener('gamepaddisconnected', handleGamepadDisconnected)

    const existingDevices = getAvailableDevices()
    if (existingDevices.length > 0) {
      setDevices(existingDevices)
      dispatch(setHIDDevices(existingDevices))
      if (!activeDeviceId) {
        setActiveDeviceId(existingDevices[0].id)
        dispatch(setActiveHIDDevice(existingDevices[0].id))
      }
    }

    pollTimerRef.current = window.setInterval(readGamepadState, GAMEPAD_POLL_INTERVAL_MS)

    return () => {
      window.removeEventListener('gamepadconnected', handleGamepadConnected)
      window.removeEventListener('gamepaddisconnected', handleGamepadDisconnected)
      if (pollTimerRef.current) {
        clearInterval(pollTimerRef.current)
        pollTimerRef.current = null
      }
    }
  }, [supported, handleGamepadConnected, handleGamepadDisconnected, readGamepadState, getAvailableDevices, dispatch, activeDeviceId])

  return {
    supported,
    devices,
    activeDeviceId,
    selectDevice,
    calibrateDevice,
    refreshDevices: getAvailableDevices
  }
}

export default useHIDGamepad
