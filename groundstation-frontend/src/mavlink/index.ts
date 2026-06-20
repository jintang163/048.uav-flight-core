export const MAV_CMD = {
  NAV_WAYPOINT: 16,
  NAV_LOITER_UNLIM: 17,
  NAV_LOITER_TURNS: 18,
  NAV_LOITER_TIME: 19,
  NAV_RETURN_TO_LAUNCH: 20,
  NAV_LAND: 21,
  NAV_TAKEOFF: 22,
  NAV_DELAY: 93,
  NAV_SCRIPT_TIME: 9005,
  CONDITION_DELAY: 112,
  CONDITION_DISTANCE: 114,
  CONDITION_YAW: 115,
  DO_SET_MODE: 176,
  DO_JUMP: 177,
  DO_CHANGE_SPEED: 178,
  DO_SET_HOME: 179,
  DO_SET_PARAMETER: 230,
  DO_MOUNT_CONTROL: 205,
  DO_SET_CAM_TRIGG_DIST: 206,
  DO_SET_CAM_TRIGG_INTERVAL: 214,
  DO_DIGICAM_CONTROL: 203,
  DO_DIGICAM_CONFIGURE: 202,
  DO_CAMERA_CONTROL: 150,
  DO_VIDEO_START: 2000,
  DO_VIDEO_STOP: 2001,
  DO_START_MAG_CAL: 142,
  DO_ACCEPT_MAG_CAL: 143,
  DO_CANCEL_MAG_CAL: 144,
  MISSION_START: 300,
  MISSION_STOP: 301,
  COMPONENT_ARM_DISARM: 400
}

export const MAV_MODE = {
  MANUAL: 0,
  STABILIZE: 1,
  GUIDED: 2,
  AUTO: 3,
  TEST: 4,
  CIRCLE: 7,
  ACRO: 8,
  ALT_HOLD: 16,
  POSHOLD: 17,
  BRAKE: 17,
  THROW: 18,
  AVOID_ADSB: 19,
  LOITER: 19,
  RTL: 21,
  SMART_RTL: 22,
  LAND: 23,
  FOLLOW: 24,
  ZIGZAG: 25,
  SYSTEMID: 26,
  AUTOROTATE: 27,
  NEW_MODE: 29,
  AUTO_RTL: 30,
  TURTLE: 31,
  DRIFT: 32,
  SPORT: 33,
  FLIP: 34,
  AUTOTUNE: 35,
  POSITION: 36
}

export const MAV_STATE = {
  UNINIT: 0,
  BOOT: 1,
  CALIBRATING: 2,
  STANDBY: 3,
  ACTIVE: 4,
  CRITICAL: 5,
  EMERGENCY: 6,
  POWEROFF: 7,
  FLIGHT_TERMINATION: 8
}

export const MAV_SEVERITY = {
  EMERGENCY: 0,
  ALERT: 1,
  CRITICAL: 2,
  ERROR: 3,
  WARNING: 4,
  NOTICE: 5,
  INFO: 6,
  DEBUG: 7
}

export const GPS_FIX_TYPE = {
  NO_GPS: 0,
  NO_FIX: 1,
  FIX_2D: 2,
  FIX_3D: 3,
  DGPS: 4,
  RTK_FLOAT: 5,
  RTK_FIXED: 6,
  STATIC: 7,
  PPP: 8
}

export const MAV_FRAME = {
  GLOBAL: 0,
  LOCAL_NED: 1,
  MISSION: 2,
  GLOBAL_RELATIVE_ALT: 3,
  LOCAL_ENU: 4,
  GLOBAL_INT: 5,
  GLOBAL_RELATIVE_ALT_INT: 6,
  LOCAL_OFFSET_NED: 7,
  BODY_NED: 8,
  BODY_OFFSET_NED: 9,
  GLOBAL_TERRAIN_ALT: 10,
  GLOBAL_TERRAIN_ALT_INT: 11
}

export interface MAVLinkMessage {
  sysid: number
  compid: number
  msgid: number
  payload: Record<string, unknown>
  time_boot_ms: number
}

export interface HeartbeatMessage extends MAVLinkMessage {
  msgid: 0
  payload: {
    type: number
    autopilot: number
    base_mode: number
    custom_mode: number
    system_status: number
    mavlink_version: number
  }
}

export interface AttitudeMessage extends MAVLinkMessage {
  msgid: 30
  payload: {
    time_boot_ms: number
    roll: number
    pitch: number
    yaw: number
    rollspeed: number
    pitchspeed: number
    yawspeed: number
  }
}

export interface GPSRawIntMessage extends MAVLinkMessage {
  msgid: 24
  payload: {
    time_usec: number
    fix_type: number
    lat: number
    lon: number
    alt: number
    eph: number
    epv: number
    vel: number
    cog: number
    satellites_visible: number
  }
}

export interface BatteryStatusMessage extends MAVLinkMessage {
  msgid: 147
  payload: {
    id: number
    battery_function: number
    type: number
    temperature: number
    voltages: number[]
    current_battery: number
    current_consumed: number
    energy_consumed: number
    battery_remaining: number
    time_remaining: number
    charge_state: number
  }
}

export interface RCChannelsMessage extends MAVLinkMessage {
  msgid: 65
  payload: {
    time_boot_ms: number
    chancount: number
    chan1_raw: number
    chan2_raw: number
    chan3_raw: number
    chan4_raw: number
    chan5_raw: number
    chan6_raw: number
    chan7_raw: number
    chan8_raw: number
    chan9_raw: number
    chan10_raw: number
    chan11_raw: number
    chan12_raw: number
    chan13_raw: number
    chan14_raw: number
    chan15_raw: number
    chan16_raw: number
    rssi: number
  }
}

export interface VFRHUDMessage extends MAVLinkMessage {
  msgid: 74
  payload: {
    airspeed: number
    groundspeed: number
    throttle: number
    alt: number
    climb: number
  }
}

export interface SystemStatusMessage extends MAVLinkMessage {
  msgid: 1
  payload: {
    load: number
    voltage_battery: number
    current_battery: number
    battery_remaining: number
    drop_rate_comm: number
    errors_comm: number
    errors_count1: number
    errors_count2: number
    errors_count3: number
    errors_count4: number
  }
}

export interface MissionItemMessage extends MAVLinkMessage {
  msgid: 39
  payload: {
    target_system: number
    target_component: number
    seq: number
    frame: number
    command: number
    current: number
    autocontinue: number
    param1: number
    param2: number
    param3: number
    param4: number
    x: number
    y: number
    z: number
    mission_type: number
  }
}

export interface MissionCurrentMessage extends MAVLinkMessage {
  msgid: 42
  payload: {
    seq: number
  }
}

export interface MissionItemReachedMessage extends MAVLinkMessage {
  msgid: 46
  payload: {
    seq: number
  }
}

export interface StatusTextMessage extends MAVLinkMessage {
  msgid: 253
  payload: {
    severity: number
    text: string
    id: number
    chunk_seq: number
  }
}

export const parseMAVLinkMessage = (data: ArrayBuffer): MAVLinkMessage | null => {
  try {
    const view = new DataView(data)
    const payloadLength = view.getUint8(1)
    const msgId = view.getUint8(5)
    const sysId = view.getUint8(3)
    const compId = view.getUint8(4)
    
    const payload: Record<string, unknown> = {}
    
    switch (msgId) {
      case 0:
        payload.type = view.getUint8(8)
        payload.autopilot = view.getUint8(9)
        payload.base_mode = view.getUint8(10)
        payload.custom_mode = view.getUint32(11, true)
        payload.system_status = view.getUint8(15)
        payload.mavlink_version = view.getUint8(16)
        break
      case 30:
        payload.time_boot_ms = view.getUint32(8, true)
        payload.roll = view.getFloat32(12, true)
        payload.pitch = view.getFloat32(16, true)
        payload.yaw = view.getFloat32(20, true)
        payload.rollspeed = view.getFloat32(24, true)
        payload.pitchspeed = view.getFloat32(28, true)
        payload.yawspeed = view.getFloat32(32, true)
        break
      case 74:
        payload.airspeed = view.getFloat32(8, true)
        payload.groundspeed = view.getFloat32(12, true)
        payload.throttle = view.getInt16(16, true)
        payload.alt = view.getFloat32(18, true)
        payload.climb = view.getFloat32(22, true)
        break
    }
    
    return {
      sysid: sysId,
      compid: compId,
      msgid: msgId,
      payload,
      time_boot_ms: Date.now()
    }
  } catch (error) {
    console.error('Failed to parse MAVLink message:', error)
    return null
  }
}

export const encodeMAVLinkMessage = (msg: MAVLinkMessage): ArrayBuffer => {
  const buffer = new ArrayBuffer(255)
  const view = new DataView(buffer)
  
  view.setUint8(0, 0xFD)
  view.setUint8(1, 0)
  view.setUint8(2, 0)
  view.setUint8(3, msg.sysid)
  view.setUint8(4, msg.compid)
  view.setUint8(5, msg.msgid)
  view.setUint8(6, msg.msgid >> 8)
  view.setUint8(7, msg.msgid >> 16)
  
  return buffer
}

export const mavModeToUAVMode = (mode: number): string => {
  const modeMap: Record<number, string> = {
    [MAV_MODE.STABILIZE]: 'stabilize',
    [MAV_MODE.ALT_HOLD]: 'alt_hold',
    [MAV_MODE.LOITER]: 'loiter',
    [MAV_MODE.AUTO]: 'auto',
    [MAV_MODE.GUIDED]: 'guided',
    [MAV_MODE.RTL]: 'rtl',
    [MAV_MODE.LAND]: 'land',
    [MAV_MODE.CIRCLE]: 'circle',
    [MAV_MODE.MANUAL]: 'manual'
  }
  return modeMap[mode] || 'manual'
}

export const uavModeToMAVMode = (mode: string): number => {
  const modeMap: Record<string, number> = {
    'stabilize': MAV_MODE.STABILIZE,
    'alt_hold': MAV_MODE.ALT_HOLD,
    'loiter': MAV_MODE.LOITER,
    'auto': MAV_MODE.AUTO,
    'guided': MAV_MODE.GUIDED,
    'rtl': MAV_MODE.RTL,
    'land': MAV_MODE.LAND,
    'circle': MAV_MODE.CIRCLE,
    'manual': MAV_MODE.MANUAL
  }
  return modeMap[mode] ?? MAV_MODE.STABILIZE
}

export const mavSeverityToAlertSeverity = (severity: number): string => {
  const severityMap: Record<number, string> = {
    [MAV_SEVERITY.EMERGENCY]: 'critical',
    [MAV_SEVERITY.ALERT]: 'critical',
    [MAV_SEVERITY.CRITICAL]: 'critical',
    [MAV_SEVERITY.ERROR]: 'error',
    [MAV_SEVERITY.WARNING]: 'warning',
    [MAV_SEVERITY.NOTICE]: 'info',
    [MAV_SEVERITY.INFO]: 'info',
    [MAV_SEVERITY.DEBUG]: 'info'
  }
  return severityMap[severity] || 'info'
}
