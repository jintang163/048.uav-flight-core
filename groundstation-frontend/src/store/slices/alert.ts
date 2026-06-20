import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import { getAlertList, getAlertStats, acknowledgeAlert, resolveAlert } from '@/api/alert'
import type { Alert, AlertStats, AlertSeverity, AlertStatus, AlertFilter, PageResult } from '@/types'

interface AlertState {
  alerts: Alert[]
  unreadCount: number
  stats: AlertStats
  loading: boolean
  statsLoading: boolean
  error: string | null
  total: number
  currentPage: number
  pageSize: number
  filter: AlertFilter
}

const initialState: AlertState = {
  alerts: [],
  unreadCount: 0,
  stats: {
    total: 0,
    active: 0,
    acknowledged: 0,
    resolved: 0,
    critical: 0,
    error: 0,
    warning: 0,
    info: 0
  },
  loading: false,
  statsLoading: false,
  error: null,
  total: 0,
  currentPage: 1,
  pageSize: 20,
  filter: {}
}

export const fetchAlerts = createAsyncThunk<PageResult<Alert>, AlertFilter & { page?: number; pageSize?: number }>(
  'alert/fetchList',
  async (params, { rejectWithValue }) => {
    try {
      return await getAlertList(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取告警列表失败')
    }
  }
)

export const fetchAlertStats = createAsyncThunk<AlertStats, { startTime?: number; endTime?: number; uavId?: string }>(
  'alert/fetchStats',
  async (params, { rejectWithValue }) => {
    try {
      return await getAlertStats(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取告警统计失败')
    }
  }
)

export const acknowledgeAlertById = createAsyncThunk<Alert, string>(
  'alert/acknowledge',
  async (id, { rejectWithValue }) => {
    try {
      return await acknowledgeAlert(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '确认告警失败')
    }
  }
)

export const resolveAlertById = createAsyncThunk<Alert, { id: string; notes?: string }>(
  'alert/resolve',
  async ({ id, notes }, { rejectWithValue }) => {
    try {
      return await resolveAlert(id, notes)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '解决告警失败')
    }
  }
)

const alertSlice = createSlice({
  name: 'alert',
  initialState,
  reducers: {
    addAlert: (state, action: PayloadAction<Alert>) => {
      state.alerts.unshift(action.payload)
      if (action.payload.status === 'active') {
        state.unreadCount++
      }
      state.stats.total++
      if (action.payload.status === 'active') state.stats.active++
      switch (action.payload.severity) {
        case 'critical': state.stats.critical++; break
        case 'error': state.stats.error++; break
        case 'warning': state.stats.warning++; break
        case 'info': state.stats.info++; break
      }
    },
    updateAlertStatus: (state, action: PayloadAction<{ id: string; status: AlertStatus }>) => {
      const alert = state.alerts.find(a => a.id === action.payload.id)
      if (alert) {
        if (alert.status === 'active' && action.payload.status !== 'active') {
          state.stats.active--
          state.unreadCount--
        }
        if (action.payload.status === 'acknowledged') state.stats.acknowledged++
        if (action.payload.status === 'resolved') state.stats.resolved++
        alert.status = action.payload.status
      }
    },
    updateAlertSeverity: (state, action: PayloadAction<{ id: string; severity: AlertSeverity }>) => {
      const alert = state.alerts.find(a => a.id === action.payload.id)
      if (alert) {
        alert.severity = action.payload.severity
      }
    },
    setFilter: (state, action: PayloadAction<AlertFilter>) => {
      state.filter = action.payload
    },
    setPage: (state, action: PayloadAction<number>) => {
      state.currentPage = action.payload
    },
    setPageSize: (state, action: PayloadAction<number>) => {
      state.pageSize = action.payload
    },
    decrementUnreadCount: (state) => {
      if (state.unreadCount > 0) {
        state.unreadCount--
      }
    },
    resetUnreadCount: (state) => {
      state.unreadCount = 0
    },
    clearAlertError: (state) => {
      state.error = null
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchAlerts.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchAlerts.fulfilled, (state, action) => {
        state.loading = false
        state.alerts = action.payload.list
        state.total = action.payload.total
        state.unreadCount = action.payload.list.filter(a => a.status === 'active').length
      })
      .addCase(fetchAlerts.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(fetchAlertStats.pending, (state) => {
        state.statsLoading = true
      })
      .addCase(fetchAlertStats.fulfilled, (state, action) => {
        state.statsLoading = false
        state.stats = action.payload
      })
      .addCase(fetchAlertStats.rejected, (state, action) => {
        state.statsLoading = false
        state.error = action.payload as string
      })
      .addCase(acknowledgeAlertById.fulfilled, (state, action) => {
        const index = state.alerts.findIndex(a => a.id === action.payload.id)
        if (index !== -1) {
          state.alerts[index] = action.payload
        }
      })
      .addCase(resolveAlertById.fulfilled, (state, action) => {
        const index = state.alerts.findIndex(a => a.id === action.payload.id)
        if (index !== -1) {
          state.alerts[index] = action.payload
        }
      })
  }
})

export const {
  addAlert,
  updateAlertStatus,
  updateAlertSeverity,
  setFilter,
  setPage,
  setPageSize,
  decrementUnreadCount,
  resetUnreadCount,
  clearAlertError
} = alertSlice.actions

export default alertSlice.reducer
