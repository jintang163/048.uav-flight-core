import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit'
import type {
  PayloadDevice,
  CameraStatus,
  SprayerStatus,
  OrbitMission,
  OrthoMission,
  TextToSpeechTask,
  SpeakerAudio
} from '@/types'
import * as payloadApi from '@/api/payload'

interface PayloadState {
  payloads: PayloadDevice[]
  currentPayload: PayloadDevice | null
  cameraStatusMap: Record<string, CameraStatus>
  sprayerStatusMap: Record<string, SprayerStatus>
  orbitMissions: OrbitMission[]
  currentOrbit: OrbitMission | null
  orthoMissions: OrthoMission[]
  currentOrtho: OrthoMission | null
  ttsTasks: TextToSpeechTask[]
  speakerAudios: SpeakerAudio[]
  selectedArea: { lat: number; lng: number }[] | null
  orbitCenter: { lat: number; lng: number } | null
  loading: boolean
  error: string | null
}

const initialState: PayloadState = {
  payloads: [],
  currentPayload: null,
  cameraStatusMap: {},
  sprayerStatusMap: {},
  orbitMissions: [],
  currentOrbit: null,
  orthoMissions: [],
  currentOrtho: null,
  ttsTasks: [],
  speakerAudios: [],
  selectedArea: null,
  orbitCenter: null,
  loading: false,
  error: null
}

export const fetchPayloads = createAsyncThunk(
  'payload/fetchPayloads',
  async (params?: { uavId?: string; type?: string }) => {
    const res = await payloadApi.listPayloads({
      page: 1,
      pageSize: 100,
      uavId: params?.uavId,
      payloadType: params?.type
    })
    return res.list
  }
)

export const fetchOrbitMissions = createAsyncThunk(
  'payload/fetchOrbitMissions',
  async (params?: { uavId?: string; status?: string }) => {
    const res = await payloadApi.listOrbitMissions({
      page: 1,
      pageSize: 100,
      ...params
    })
    return res.list
  }
)

export const fetchOrthoMissions = createAsyncThunk(
  'payload/fetchOrthoMissions',
  async (params?: { uavId?: string; status?: string }) => {
    const res = await payloadApi.listOrthoMissions({
      page: 1,
      pageSize: 100,
      ...params
    })
    return res.list
  }
)

const payloadSlice = createSlice({
  name: 'payload',
  initialState,
  reducers: {
    setCurrentPayload: (state, action: PayloadAction<PayloadDevice | null>) => {
      state.currentPayload = action.payload
    },
    updateCameraStatus: (state, action: PayloadAction<{ payloadId: string; status: CameraStatus }>) => {
      state.cameraStatusMap[action.payload.payloadId] = action.payload.status
    },
    updateSprayerStatus: (state, action: PayloadAction<{ payloadId: string; status: SprayerStatus }>) => {
      state.sprayerStatusMap[action.payload.payloadId] = action.payload.status
    },
    updatePayloadDevice: (state, action: PayloadAction<PayloadDevice>) => {
      const idx = state.payloads.findIndex(p => p.id === action.payload.id)
      if (idx >= 0) {
        state.payloads[idx] = { ...state.payloads[idx], ...action.payload }
      }
    },
    updateOrbitMission: (state, action: PayloadAction<OrbitMission>) => {
      const idx = state.orbitMissions.findIndex(o => o.id === action.payload.id)
      if (idx >= 0) {
        state.orbitMissions[idx] = { ...state.orbitMissions[idx], ...action.payload }
        if (state.currentOrbit?.id === action.payload.id) {
          state.currentOrbit = { ...state.currentOrbit, ...action.payload }
        }
      }
    },
    updateOrthoMission: (state, action: PayloadAction<OrthoMission>) => {
      const idx = state.orthoMissions.findIndex(o => o.id === action.payload.id)
      if (idx >= 0) {
        state.orthoMissions[idx] = { ...state.orthoMissions[idx], ...action.payload }
        if (state.currentOrtho?.id === action.payload.id) {
          state.currentOrtho = { ...state.currentOrtho, ...action.payload }
        }
      }
    },
    setSelectedArea: (state, action: PayloadAction<{ lat: number; lng: number }[] | null>) => {
      state.selectedArea = action.payload
    },
    setOrbitCenter: (state, action: PayloadAction<{ lat: number; lng: number } | null>) => {
      state.orbitCenter = action.payload
    },
    setCurrentOrbit: (state, action: PayloadAction<OrbitMission | null>) => {
      state.currentOrbit = action.payload
    },
    setCurrentOrtho: (state, action: PayloadAction<OrthoMission | null>) => {
      state.currentOrtho = action.payload
    },
    addTTSTask: (state, action: PayloadAction<TextToSpeechTask>) => {
      state.ttsTasks.unshift(action.payload)
    },
    updateTTSTask: (state, action: PayloadAction<TextToSpeechTask>) => {
      const idx = state.ttsTasks.findIndex(t => t.id === action.payload.id)
      if (idx >= 0) {
        state.ttsTasks[idx] = { ...state.ttsTasks[idx], ...action.payload }
      }
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchPayloads.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchPayloads.fulfilled, (state, action) => {
        state.payloads = action.payload
        state.loading = false
      })
      .addCase(fetchPayloads.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to fetch payloads'
      })
      .addCase(fetchOrbitMissions.fulfilled, (state, action) => {
        state.orbitMissions = action.payload
      })
      .addCase(fetchOrthoMissions.fulfilled, (state, action) => {
        state.orthoMissions = action.payload
      })
  }
})

export const {
  setCurrentPayload,
  updateCameraStatus,
  updateSprayerStatus,
  updatePayloadDevice,
  updateOrbitMission,
  updateOrthoMission,
  setSelectedArea,
  setOrbitCenter,
  setCurrentOrbit,
  setCurrentOrtho,
  addTTSTask,
  updateTTSTask,
  setError
} = payloadSlice.actions

export default payloadSlice.reducer
