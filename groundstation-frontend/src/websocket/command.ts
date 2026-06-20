import type WebSocketClient from './client'
import type { UAVMode, WaypointAction } from '@/types'
import { getWebSocketClient } from './client'

export const sendCommand = async (uavId: string, command: string, params: Record<string, unknown> = {}): Promise<void> => {
  const wsClient = getWebSocketClient()
  if (!wsClient) {
    throw new Error('WebSocket not connected')
  }
  wsClient.send('command', {
    uavId,
    command,
    params
  })
}

export const sendArmCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'arm',
    params: {}
  })
}

export const sendDisarmCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'disarm',
    params: {}
  })
}

export const sendTakeoffCommand = (wsClient: WebSocketClient, uavId: string, altitude: number): void => {
  wsClient.send('command', {
    uavId,
    command: 'takeoff',
    params: { altitude }
  })
}

export const sendLandCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'land',
    params: {}
  })
}

export const sendRTLCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'rtl',
    params: {}
  })
}

export const sendGotoCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  lat: number,
  lng: number,
  alt: number,
  relative: boolean = true
): void => {
  wsClient.send('command', {
    uavId,
    command: 'goto',
    params: { lat, lng, alt, relative }
  })
}

export const sendSetModeCommand = (wsClient: WebSocketClient, uavId: string, mode: UAVMode): void => {
  wsClient.send('command', {
    uavId,
    command: 'set_mode',
    params: { mode }
  })
}

export const sendSetHomeCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  lat: number,
  lng: number,
  alt: number
): void => {
  wsClient.send('command', {
    uavId,
    command: 'set_home',
    params: { lat, lng, alt }
  })
}

export const sendSetHomeCurrentCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'set_home_current',
    params: {}
  })
}

export const sendVelocityCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  vx: number,
  vy: number,
  vz: number,
  yawRate: number
): void => {
  wsClient.send('command', {
    uavId,
    command: 'velocity',
    params: { vx, vy, vz, yawRate }
  })
}

export const sendYawCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  angle: number,
  relative: boolean = false
): void => {
  wsClient.send('command', {
    uavId,
    command: 'yaw',
    params: { angle, relative }
  })
}

export const sendMissionStartCommand = (wsClient: WebSocketClient, uavId: string, missionId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'mission_start',
    params: { missionId }
  })
}

export const sendMissionPauseCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'mission_pause',
    params: {}
  })
}

export const sendMissionResumeCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'mission_resume',
    params: {}
  })
}

export const sendMissionStopCommand = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'mission_stop',
    params: {}
  })
}

export const sendMissionSetCurrentCommand = (wsClient: WebSocketClient, uavId: string, waypointIndex: number): void => {
  wsClient.send('command', {
    uavId,
    command: 'mission_set_current',
    params: { waypointIndex }
  })
}

export const sendMissionUploadCommand = (wsClient: WebSocketClient, uavId: string, waypoints: Array<{
  seq: number
  command: number
  param1: number
  param2: number
  param3: number
  param4: number
  x: number
  y: number
  z: number
}>): void => {
  wsClient.send('command', {
    uavId,
    command: 'mission_upload',
    params: { waypoints }
  })
}

export const sendDoActionCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  action: WaypointAction,
  params: Record<string, number>
): void => {
  wsClient.send('command', {
    uavId,
    command: 'do_action',
    params: { action, ...params }
  })
}

export const sendChangeSpeedCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  speedType: number,
  speed: number,
  relative: boolean = false
): void => {
  wsClient.send('command', {
    uavId,
    command: 'change_speed',
    params: { speedType, speed, relative }
  })
}

export const sendCalibrateCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  calibrationType: 'gyro' | 'compass' | 'accelerometer' | 'level'
): void => {
  wsClient.send('command', {
    uavId,
    command: 'calibrate',
    params: { type: calibrationType }
  })
}

export const sendRebootCommand = (wsClient: WebSocketClient, uavId: string, force: boolean = false): void => {
  wsClient.send('command', {
    uavId,
    command: 'reboot',
    params: { force }
  })
}

export const sendParameterSetCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  paramId: string,
  paramValue: number,
  paramType: number
): void => {
  wsClient.send('command', {
    uavId,
    command: 'param_set',
    params: { paramId, paramValue, paramType }
  })
}

export const sendParameterRequestCommand = (wsClient: WebSocketClient, uavId: string, paramId: string): void => {
  wsClient.send('command', {
    uavId,
    command: 'param_request',
    params: { paramId }
  })
}

export const sendCameraTriggerCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  triggerType: number,
  interval: number = 0,
  totalPhotos: number = 1
): void => {
  wsClient.send('command', {
    uavId,
    command: 'camera_trigger',
    params: { triggerType, interval, totalPhotos }
  })
}

export const sendMountControlCommand = (
  wsClient: WebSocketClient,
  uavId: string,
  pitch: number,
  roll: number,
  yaw: number,
  mode: number
): void => {
  wsClient.send('command', {
    uavId,
    command: 'mount_control',
    params: { pitch, roll, yaw, mode }
  })
}

export const sendCommandAck = (wsClient: WebSocketClient, commandId: string, result: boolean, message?: string): void => {
  wsClient.send('command_ack', {
    commandId,
    result,
    message
  })
}
