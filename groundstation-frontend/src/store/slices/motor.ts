import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit'
import type { MotorStatus, MotorFailureAlert, MotorFailureState } from '@/types'
import * as motorApi from '@/api/motor'

interface MotorState {
  motorStatuses: Record<string, MotorStatus[]>
  failureStates: Record<string, MotorFailureState | null>
  activeAlerts: MotorFailureAlert[]
  loading: boolean
  error: string | null
}

const initialState: MotorState = {
  motorStatuses: {},
  failureStates: {},
  activeAlerts: [],
  loading: false,
  error: null
}

export const fetchMotorStatuses = createAsyncThunk(
  'motor/fetchMotorStatuses',
  async (uavId: string) => {
    const res = await motorApi.getMotorStatuses(uavId)
    return { uavId, statuses: res }
  }
)

export const fetchMotorFailureState = createAsyncThunk(
  'motor/fetchMotorFailureState',
  async (uavId: string) => {
    const res = await motorApi.getMotorFailureState(uavId)
    return { uavId, state: res }
  }
)

const motorSlice = createSlice({
  name: 'motor',
  initialState,
  reducers: {
    updateMotorStatus: (state, action: PayloadAction<{ uavId: string; motor: MotorStatus }>) => {
      const { uavId, motor } = action.payload
      if (!state.motorStatuses[uavId]) {
        state.motorStatuses[uavId] = []
      }
      const idx = state.motorStatuses[uavId].findIndex(m => m.motor_index === motor.motor_index)
      if (idx >= 0) {
        state.motorStatuses[uavId][idx] = { ...state.motorStatuses[uavId][idx], ...motor }
      } else {
        state.motorStatuses[uavId].push(motor)
      }
    },
    addMotorFailureAlert: (state, action: PayloadAction<MotorFailureAlert>) => {
      const exists = state.activeAlerts.find(a => a.motorIndex === action.payload.motorIndex && a.uavId === action.payload.uavId && !a.resolved)
      if (!exists) {
        state.activeAlerts.unshift(action.payload)
      }
    },
    dismissMotorAlert: (state, action: PayloadAction<string>) => {
      state.activeAlerts = state.activeAlerts.filter(a => a.id !== action.payload)
    },
    clearMotorAlerts: (state) => {
      state.activeAlerts = []
    },
    updateFailureState: (state, action: PayloadAction<{ uavId: string; state: MotorFailureState | null }>) => {
      state.failureStates[action.payload.uavId] = action.payload.state
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchMotorStatuses.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchMotorStatuses.fulfilled, (state, action) => {
        state.motorStatuses[action.payload.uavId] = action.payload.statuses
        state.loading = false
      })
      .addCase(fetchMotorStatuses.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to fetch motor statuses'
      })
      .addCase(fetchMotorFailureState.fulfilled, (state, action) => {
        state.failureStates[action.payload.uavId] = action.payload.state
      })
  }
})

export const {
  updateMotorStatus,
  addMotorFailureAlert,
  dismissMotorAlert,
  clearMotorAlerts,
  updateFailureState
} = motorSlice.actions

export default motorSlice.reducer
