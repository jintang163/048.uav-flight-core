import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type {
  ObstacleDetection,
  ObstacleAvoidanceEvent,
  ObstacleAvoidanceConfig,
  ObstacleHeatmapPoint,
  ObstacleAvoidanceLog,
  ObstacleAvoidanceStatistics,
  AvoidanceSensitivity,
  AvoidanceStrategy
} from '@/types/obstacle-avoidance'

interface ObstacleAvoidanceState {
  currentDetections: ObstacleDetection[]
  activeAvoidanceEvent: ObstacleAvoidanceEvent | null
  recentAvoidanceEvents: ObstacleAvoidanceEvent[]
  heatmapPoints: ObstacleHeatmapPoint[]
  config: ObstacleAvoidanceConfig | null
  logs: ObstacleAvoidanceLog[]
  statistics: ObstacleAvoidanceStatistics | null
  enabled: boolean
  sensitivity: AvoidanceSensitivity
  strategy: AvoidanceStrategy
  loading: boolean
  error: string | null
}

const initialState: ObstacleAvoidanceState = {
  currentDetections: [],
  activeAvoidanceEvent: null,
  recentAvoidanceEvents: [],
  heatmapPoints: [],
  config: null,
  logs: [],
  statistics: null,
  enabled: true,
  sensitivity: 'medium',
  strategy: 'ascend_bypass',
  loading: false,
  error: null
}

const obstacleAvoidanceSlice = createSlice({
  name: 'obstacleAvoidance',
  initialState,
  reducers: {
    updateDetections(state, action: PayloadAction<ObstacleDetection[]>) {
      state.currentDetections = action.payload
    },
    addDetection(state, action: PayloadAction<ObstacleDetection>) {
      state.currentDetections.push(action.payload)
      if (state.currentDetections.length > 50) {
        state.currentDetections = state.currentDetections.slice(-50)
      }
    },
    clearDetections(state) {
      state.currentDetections = []
    },
    setActiveAvoidanceEvent(state, action: PayloadAction<ObstacleAvoidanceEvent | null>) {
      state.activeAvoidanceEvent = action.payload
    },
    updateAvoidanceEvent(state, action: PayloadAction<Partial<ObstacleAvoidanceEvent>>) {
      if (state.activeAvoidanceEvent) {
        state.activeAvoidanceEvent = { ...state.activeAvoidanceEvent, ...action.payload }
      }
    },
    addAvoidanceEvent(state, action: PayloadAction<ObstacleAvoidanceEvent>) {
      state.recentAvoidanceEvents.unshift(action.payload)
      if (state.recentAvoidanceEvents.length > 100) {
        state.recentAvoidanceEvents = state.recentAvoidanceEvents.slice(0, 100)
      }
      state.activeAvoidanceEvent = action.payload
    },
    completeAvoidanceEvent(state, action: PayloadAction<{ id: string; completedAt: number }>) {
      if (state.activeAvoidanceEvent && state.activeAvoidanceEvent.id === action.payload.id) {
        state.activeAvoidanceEvent.status = 'completed'
        state.activeAvoidanceEvent.completedAt = action.payload.completedAt
      }
      const idx = state.recentAvoidanceEvents.findIndex(e => e.id === action.payload.id)
      if (idx !== -1) {
        state.recentAvoidanceEvents[idx].status = 'completed'
        state.recentAvoidanceEvents[idx].completedAt = action.payload.completedAt
      }
    },
    updateHeatmapPoints(state, action: PayloadAction<ObstacleHeatmapPoint[]>) {
      state.heatmapPoints = action.payload
    },
    addHeatmapPoint(state, action: PayloadAction<ObstacleHeatmapPoint>) {
      const existing = state.heatmapPoints.find(
        p => Math.abs(p.lat - action.payload.lat) < 0.00001 && Math.abs(p.lng - action.payload.lng) < 0.00001
      )
      if (existing) {
        existing.triggerCount += action.payload.triggerCount
        existing.lastTriggerTime = action.payload.lastTriggerTime
        existing.intensity = Math.min(1, existing.triggerCount / 20)
        existing.avgDistance = (existing.avgDistance + action.payload.avgDistance) / 2
        existing.minDistance = Math.min(existing.minDistance, action.payload.minDistance)
      } else {
        state.heatmapPoints.push(action.payload)
      }
    },
    setConfig(state, action: PayloadAction<ObstacleAvoidanceConfig>) {
      state.config = action.payload
      state.enabled = action.payload.enabled
      state.sensitivity = action.payload.sensitivity
      state.strategy = action.payload.strategy
    },
    updateConfig(state, action: PayloadAction<Partial<ObstacleAvoidanceConfig>>) {
      if (state.config) {
        state.config = { ...state.config, ...action.payload }
      }
      if (action.payload.enabled !== undefined) state.enabled = action.payload.enabled
      if (action.payload.sensitivity !== undefined) state.sensitivity = action.payload.sensitivity
      if (action.payload.strategy !== undefined) state.strategy = action.payload.strategy
    },
    setEnabled(state, action: PayloadAction<boolean>) {
      state.enabled = action.payload
      if (state.config) {
        state.config.enabled = action.payload
      }
    },
    setSensitivity(state, action: PayloadAction<AvoidanceSensitivity>) {
      state.sensitivity = action.payload
      if (state.config) {
        state.config.sensitivity = action.payload
      }
    },
    setStrategy(state, action: PayloadAction<AvoidanceStrategy>) {
      state.strategy = action.payload
      if (state.config) {
        state.config.strategy = action.payload
      }
    },
    setLogs(state, action: PayloadAction<ObstacleAvoidanceLog[]>) {
      state.logs = action.payload
    },
    addLog(state, action: PayloadAction<ObstacleAvoidanceLog>) {
      state.logs.unshift(action.payload)
      if (state.logs.length > 200) {
        state.logs = state.logs.slice(0, 200)
      }
    },
    setStatistics(state, action: PayloadAction<ObstacleAvoidanceStatistics>) {
      state.statistics = action.payload
    },
    setLoading(state, action: PayloadAction<boolean>) {
      state.loading = action.payload
    },
    setError(state, action: PayloadAction<string | null>) {
      state.error = action.payload
    }
  }
})

export const {
  updateDetections,
  addDetection,
  clearDetections,
  setActiveAvoidanceEvent,
  updateAvoidanceEvent,
  addAvoidanceEvent,
  completeAvoidanceEvent,
  updateHeatmapPoints,
  addHeatmapPoint,
  setConfig,
  updateConfig,
  setEnabled,
  setSensitivity,
  setStrategy,
  setLogs,
  addLog,
  setStatistics,
  setLoading,
  setError
} = obstacleAvoidanceSlice.actions

export default obstacleAvoidanceSlice.reducer
