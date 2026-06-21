import { get, post, put, del } from './http'
import type {
  PayloadDevice,
  CameraStatus,
  SprayerStatus,
  SpeakerAudio,
  OrbitMission,
  OrthoMission,
  TextToSpeechTask,
  OrbitCreateRequest,
  OrthoPlanRequest,
  TTSCreateRequest
} from '@/types'

export const listPayloads = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  payloadType?: string
  status?: string
  keyword?: string
}): Promise<{ list: PayloadDevice[]; total: number }> => {
  return get<{ list: PayloadDevice[]; total: number }>('/payloads', params)
}

export const listUAVPayloads = (uavId: string): Promise<PayloadDevice[]> => {
  return get<PayloadDevice[]>(`/payloads/uav/${uavId}`)
}

export const getPayloadDetail = (id: string): Promise<PayloadDevice> => {
  return get<PayloadDevice>(`/payloads/${id}`)
}

export const createPayload = (data: Partial<PayloadDevice>): Promise<PayloadDevice> => {
  return post<PayloadDevice>('/payloads', data)
}

export const updatePayload = (id: string, data: Partial<PayloadDevice>): Promise<PayloadDevice> => {
  return put<PayloadDevice>(`/payloads/${id}`, data)
}

export const deletePayload = (id: string): Promise<void> => {
  return del<void>(`/payloads/${id}`)
}

export const getCameraStatus = (payloadId: string): Promise<CameraStatus> => {
  return get<CameraStatus>(`/payloads/${payloadId}/camera/status`)
}

export const takePhoto = (payloadId: string): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/camera/photo`)
}

export const startVideoRecording = (payloadId: string): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/camera/recording/start`)
}

export const stopVideoRecording = (payloadId: string): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/camera/recording/stop`)
}

export const setCameraMode = (payloadId: string, mode: 'idle' | 'photo' | 'video'): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/camera/mode`, { mode })
}

export const setCameraZoom = (
  payloadId: string,
  options: number | { zoomType?: 'in' | 'out' | 'stop'; zoomSpeed?: number; zoomLevel?: number }
): Promise<void> => {
  const payload = typeof options === 'number'
    ? { zoomLevel: options }
    : options
  return post<void>(`/payloads/${payloadId}/camera/zoom`, payload)
}

export const setCameraSettings = (payloadId: string, settings: Record<string, any>): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/camera/settings`, settings)
}

export const getSprayerStatus = (payloadId: string): Promise<SprayerStatus> => {
  return get<SprayerStatus>(`/payloads/${payloadId}/sprayer/status`)
}

export const setSprayerFlowRate = (
  payloadId: string,
  options: number | { flowRate?: number }
): Promise<void> => {
  const payload = typeof options === 'number'
    ? { flowRate: options }
    : options
  return post<void>(`/payloads/${payloadId}/sprayer/flow`, payload)
}

export const startSpraying = (payloadId: string, flowRate?: number): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/sprayer/start`, { flowRate })
}

export const stopSpraying = (payloadId: string): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/sprayer/stop`)
}

export const listSpeakerAudios = (params?: {
  page?: number
  pageSize?: number
  payloadId?: string
  isTTS?: boolean
}): Promise<{ list: SpeakerAudio[]; total: number }> => {
  return get<{ list: SpeakerAudio[]; total: number }>('/payloads/speaker/audios', params)
}

export const createSpeakerAudio = (data: Partial<SpeakerAudio>): Promise<SpeakerAudio> => {
  return post<SpeakerAudio>('/payloads/speaker/audios', data)
}

export const getSpeakerAudio = (id: string): Promise<SpeakerAudio> => {
  return get<SpeakerAudio>(`/payloads/speaker/audios/${id}`)
}

export const deleteSpeakerAudio = (id: string): Promise<void> => {
  return del<void>(`/payloads/speaker/audios/${id}`)
}

export const playSpeakerAudio = (
  payloadId: string,
  options: string | { audioId?: string; loop?: boolean; volume?: number }
): Promise<void> => {
  if (typeof options === 'string') {
    return post<void>(`/payloads/${payloadId}/speaker/play/${options}`)
  }
  const { audioId, ...rest } = options
  const url = audioId
    ? `/payloads/${payloadId}/speaker/play/${audioId}`
    : `/payloads/${payloadId}/speaker/play`
  return post<void>(url, rest)
}

export const stopSpeaker = (payloadId: string): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/speaker/stop`)
}

export const listOrbitMissions = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  status?: string
  startTime?: string
  endTime?: string
}): Promise<{ list: OrbitMission[]; total: number }> => {
  return get<{ list: OrbitMission[]; total: number }>('/payload-missions/orbit', params)
}

export const createOrbitMission = (data: OrbitCreateRequest): Promise<OrbitMission> => {
  return post<OrbitMission>('/payload-missions/orbit', data)
}

export const getOrbitMission = (id: string): Promise<OrbitMission> => {
  return get<OrbitMission>(`/payload-missions/orbit/${id}`)
}

export const startOrbitMission = (id: string): Promise<OrbitMission> => {
  return post<OrbitMission>(`/payload-missions/orbit/${id}/start`)
}

export const pauseOrbitMission = (id: string): Promise<OrbitMission> => {
  return post<OrbitMission>(`/payload-missions/orbit/${id}/pause`)
}

export const resumeOrbitMission = (id: string): Promise<OrbitMission> => {
  return post<OrbitMission>(`/payload-missions/orbit/${id}/resume`)
}

export const abortOrbitMission = (id: string): Promise<OrbitMission> => {
  return post<OrbitMission>(`/payload-missions/orbit/${id}/abort`)
}

export const listOrthoMissions = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  status?: string
}): Promise<{ list: OrthoMission[]; total: number }> => {
  return get<{ list: OrthoMission[]; total: number }>('/payload-missions/ortho', params)
}

export const createOrthoMission = (data: Partial<OrthoMission>): Promise<OrthoMission> => {
  return post<OrthoMission>('/payload-missions/ortho', data)
}

export const getOrthoMission = (id: string): Promise<OrthoMission> => {
  return get<OrthoMission>(`/payload-missions/ortho/${id}`)
}

export const planOrthoMission = (id: string, data: OrthoPlanRequest): Promise<OrthoMission> => {
  return post<OrthoMission>(`/payload-missions/ortho/${id}/plan`, data)
}

export const startOrthoMission = (id: string): Promise<OrthoMission> => {
  return post<OrthoMission>(`/payload-missions/ortho/${id}/start`)
}

export const pauseOrthoMission = (id: string): Promise<OrthoMission> => {
  return post<OrthoMission>(`/payload-missions/ortho/${id}/pause`)
}

export const resumeOrthoMission = (id: string): Promise<OrthoMission> => {
  return post<OrthoMission>(`/payload-missions/ortho/${id}/resume`)
}

export const abortOrthoMission = (id: string): Promise<OrthoMission> => {
  return post<OrthoMission>(`/payload-missions/ortho/${id}/abort`)
}

export const listTTSTasks = (params?: {
  page?: number
  pageSize?: number
  payloadId?: string
  uavId?: string
  status?: string
}): Promise<{ list: TextToSpeechTask[]; total: number }> => {
  return get<{ list: TextToSpeechTask[]; total: number }>('/payload-missions/tts', params)
}

export const createTTSTask = (data: TTSCreateRequest): Promise<TextToSpeechTask> => {
  return post<TextToSpeechTask>('/payload-missions/tts', data)
}

export const getTTSTask = (id: string): Promise<TextToSpeechTask> => {
  return get<TextToSpeechTask>(`/payload-missions/tts/${id}`)
}

export const cameraTakePhoto = (
  payloadId: string,
  options?: { count?: number; intervalSec?: number }
): Promise<void> => {
  return post<void>(`/payloads/${payloadId}/camera/photo`, options || { count: 1 })
}

export const cameraStartRecording = (payloadId: string): Promise<void> => {
  return startVideoRecording(payloadId)
}

export const cameraStopRecording = (payloadId: string): Promise<void> => {
  return stopVideoRecording(payloadId)
}

export const sprayerStart = (payloadId: string): Promise<void> => {
  return startSpraying(payloadId)
}

export const sprayerStop = (payloadId: string): Promise<void> => {
  return stopSpraying(payloadId)
}

export const createTextToSpeechTask = (data: TTSCreateRequest): Promise<TextToSpeechTask> => {
  return createTTSTask(data)
}

export const listTextToSpeechTasks = listTTSTasks


