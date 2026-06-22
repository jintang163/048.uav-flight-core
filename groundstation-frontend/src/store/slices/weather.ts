import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import { getLatestWeather, getActiveWeatherAlerts, getWeatherThresholds, checkTakeoffWeather } from '@/api/weather'
import type { WeatherData, WeatherAlertEvent, WeatherThresholds, WeatherCheckResult } from '@/types'

interface WeatherState {
  currentWeather: Record<number, WeatherData>
  activeAlerts: Record<number, WeatherAlertEvent[]>
  thresholds: WeatherThresholds
  takeoffCheck: Record<number, WeatherCheckResult>
  loading: boolean
  alertsLoading: boolean
  error: string | null
}

const initialState: WeatherState = {
  currentWeather: {},
  activeAlerts: {},
  thresholds: {
    wind_speed_return_ms: 5.0,
    gust_protect_ms: 12.0,
    wind_adapt_ms: 8.0,
    low_temp_c: -10.0,
    thunderstorm_reject: true,
  },
  takeoffCheck: {},
  loading: false,
  alertsLoading: false,
  error: null,
}

export const fetchLatestWeather = createAsyncThunk<WeatherData, number>(
  'weather/fetchLatest',
  async (uavId, { rejectWithValue }) => {
    try {
      return await getLatestWeather(uavId)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取气象数据失败')
    }
  }
)

export const fetchActiveAlerts = createAsyncThunk<WeatherAlertEvent[], number>(
  'weather/fetchActiveAlerts',
  async (uavId, { rejectWithValue }) => {
    try {
      return await getActiveWeatherAlerts(uavId)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取气象预警失败')
    }
  }
)

export const fetchThresholds = createAsyncThunk<WeatherThresholds>(
  'weather/fetchThresholds',
  async (_, { rejectWithValue }) => {
    try {
      return await getWeatherThresholds()
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取阈值配置失败')
    }
  }
)

export const fetchTakeoffCheck = createAsyncThunk<WeatherCheckResult, number>(
  'weather/fetchTakeoffCheck',
  async (uavId, { rejectWithValue }) => {
    try {
      return await checkTakeoffWeather(uavId)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '起飞气象检查失败')
    }
  }
)

const weatherSlice = createSlice({
  name: 'weather',
  initialState,
  reducers: {
    updateWeatherData: (state, action: PayloadAction<WeatherData>) => {
      state.currentWeather[action.payload.uav_id] = action.payload
    },
    addWeatherAlert: (state, action: PayloadAction<{ uavId: number; alert: WeatherAlertEvent }>) => {
      const { uavId, alert } = action.payload
      if (!state.activeAlerts[uavId]) {
        state.activeAlerts[uavId] = []
      }
      const exists = state.activeAlerts[uavId].some(a => a.id === alert.id)
      if (!exists) {
        state.activeAlerts[uavId].unshift(alert)
      }
    },
    removeWeatherAlert: (state, action: PayloadAction<{ uavId: number; alertId: number }>) => {
      const { uavId, alertId } = action.payload
      if (state.activeAlerts[uavId]) {
        state.activeAlerts[uavId] = state.activeAlerts[uavId].filter(a => a.id !== alertId)
      }
    },
    updateThresholds: (state, action: PayloadAction<WeatherThresholds>) => {
      state.thresholds = action.payload
    },
    clearWeatherError: (state) => {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchLatestWeather.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchLatestWeather.fulfilled, (state, action) => {
        state.loading = false
        state.currentWeather[action.payload.uav_id] = action.payload
      })
      .addCase(fetchLatestWeather.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(fetchActiveAlerts.pending, (state) => {
        state.alertsLoading = true
      })
      .addCase(fetchActiveAlerts.fulfilled, (state, action) => {
        state.alertsLoading = false
      })
      .addCase(fetchActiveAlerts.rejected, (state, action) => {
        state.alertsLoading = false
        state.error = action.payload as string
      })
      .addCase(fetchThresholds.fulfilled, (state, action) => {
        state.thresholds = action.payload
      })
      .addCase(fetchTakeoffCheck.fulfilled, (state, action) => {
        const weatherData = action.payload.weather_data
        if (weatherData) {
          state.takeoffCheck[weatherData.uav_id] = action.payload
        }
      })
  },
})

export const {
  updateWeatherData,
  addWeatherAlert,
  removeWeatherAlert,
  updateThresholds,
  clearWeatherError,
} = weatherSlice.actions

export default weatherSlice.reducer
