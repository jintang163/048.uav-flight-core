import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type { TelemetryData, TelemetryHistoryPoint } from '@/types'

interface TelemetryState {
  currentData: TelemetryData | null
  history: TelemetryHistoryPoint[]
  maxHistoryPoints: number
  connected: boolean
  uavId: string | null
}

const initialState: TelemetryState = {
  currentData: null,
  history: [],
  maxHistoryPoints: 1000,
  connected: false,
  uavId: null
}

const telemetrySlice = createSlice({
  name: 'telemetry',
  initialState,
  reducers: {
    updateTelemetry: (state, action: PayloadAction<TelemetryData>) => {
      state.currentData = action.payload
      state.uavId = action.payload.uavId
      
      const historyPoint: TelemetryHistoryPoint = {
        timestamp: action.payload.timestamp,
        altitude: action.payload.position.alt,
        speed: action.payload.velocity.groundSpeed,
        throttle: 0,
        battery: action.payload.battery.remaining
      }
      
      state.history.push(historyPoint)
      
      if (state.history.length > state.maxHistoryPoints) {
        state.history = state.history.slice(-state.maxHistoryPoints)
      }
    },
    setTelemetryConnected: (state, action: PayloadAction<boolean>) => {
      state.connected = action.payload
    },
    clearTelemetryHistory: (state) => {
      state.history = []
      state.currentData = null
    },
    setUAVId: (state, action: PayloadAction<string | null>) => {
      state.uavId = action.payload
    }
  }
})

export const { updateTelemetry, setTelemetryConnected, clearTelemetryHistory, setUAVId } = telemetrySlice.actions
export default telemetrySlice.reducer
