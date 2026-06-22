import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import {
  getActiveCollisionAlerts,
  getAllUAVPositions,
  getCollisionStatus,
  getRouteIntersections,
  toggleCollisionAvoidance,
  resolveCollisionAlert,
} from '@/api/collision'
import type {
  CollisionAlert,
  UAVLivePosition,
  CollisionStatus,
  RouteIntersection,
  AvoidanceDecision,
} from '@/types'

interface CollisionState {
  status: CollisionStatus
  alerts: CollisionAlert[]
  positions: Record<string, UAVLivePosition>
  intersections: RouteIntersection[]
  decisions: AvoidanceDecision[]
  loading: boolean
  enabled: boolean
}

const initialState: CollisionState = {
  status: {
    enabled: false,
    active_uavs: 0,
    active_alerts: 0,
    intersections: 0,
    safe_distance_m: 50,
    warning_distance_m: 100,
  },
  alerts: [],
  positions: {},
  intersections: [],
  decisions: [],
  loading: false,
  enabled: false,
}

export const fetchCollisionStatus = createAsyncThunk<CollisionStatus>(
  'collision/fetchStatus',
  async () => {
    return await getCollisionStatus()
  }
)

export const fetchActiveAlerts = createAsyncThunk<CollisionAlert[]>(
  'collision/fetchActiveAlerts',
  async () => {
    return await getActiveCollisionAlerts()
  }
)

export const fetchAllPositions = createAsyncThunk<Record<string, UAVLivePosition>>(
  'collision/fetchPositions',
  async () => {
    return await getAllUAVPositions()
  }
)

export const fetchIntersections = createAsyncThunk<RouteIntersection[]>(
  'collision/fetchIntersections',
  async () => {
    return await getRouteIntersections()
  }
)

export const setCollisionEnabled = createAsyncThunk<{ enabled: boolean }, boolean>(
  'collision/setEnabled',
  async (enabled) => {
    return await toggleCollisionAvoidance(enabled)
  }
)

export const dismissAlert = createAsyncThunk<void, number>(
  'collision/dismissAlert',
  async (id) => {
    await resolveCollisionAlert(id)
  }
)

const collisionSlice = createSlice({
  name: 'collision',
  initialState,
  reducers: {
    addCollisionAlert: (state, action: PayloadAction<CollisionAlert>) => {
      const exists = state.alerts.some(a => a.alert_id === action.payload.alert_id)
      if (!exists) {
        state.alerts.unshift(action.payload)
      } else {
        state.alerts = state.alerts.map(a =>
          a.alert_id === action.payload.alert_id ? action.payload : a
        )
      }
      state.status.active_alerts = state.alerts.filter(a => !a.is_resolved).length
    },
    updatePosition: (state, action: PayloadAction<UAVLivePosition>) => {
      state.positions[String(action.payload.uav_id)] = action.payload
      state.status.active_uavs = Object.keys(state.positions).length
    },
    addAvoidanceDecision: (state, action: PayloadAction<AvoidanceDecision>) => {
      state.decisions = [action.payload, ...state.decisions].slice(0, 50)
    },
    updateIntersections: (state, action: PayloadAction<RouteIntersection[]>) => {
      state.intersections = action.payload
      state.status.intersections = action.payload.length
    },
    clearResolvedAlerts: (state) => {
      state.alerts = state.alerts.filter(a => !a.is_resolved)
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchCollisionStatus.fulfilled, (state, action) => {
        state.status = action.payload
        state.enabled = action.payload.enabled
      })
      .addCase(fetchActiveAlerts.fulfilled, (state, action) => {
        state.alerts = action.payload
        state.status.active_alerts = action.payload.filter(a => !a.is_resolved).length
      })
      .addCase(fetchAllPositions.fulfilled, (state, action) => {
        state.positions = action.payload
        state.status.active_uavs = Object.keys(action.payload).length
      })
      .addCase(fetchIntersections.fulfilled, (state, action) => {
        state.intersections = action.payload
        state.status.intersections = action.payload.length
      })
      .addCase(setCollisionEnabled.fulfilled, (state, action) => {
        state.enabled = action.payload.enabled
        state.status.enabled = action.payload.enabled
      })
      .addCase(dismissAlert.fulfilled, (state, action) => {
        const id = action.meta.arg
        state.alerts = state.alerts.filter(a => a.id !== id)
        state.status.active_alerts = state.alerts.filter(a => !a.is_resolved).length
      })
  },
})

export const {
  addCollisionAlert,
  updatePosition,
  addAvoidanceDecision,
  updateIntersections,
  clearResolvedAlerts,
} = collisionSlice.actions

export default collisionSlice.reducer
