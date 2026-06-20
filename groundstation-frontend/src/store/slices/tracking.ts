import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import {
  lockTarget as lockTargetAPI,
  stopTracking as stopTrackingAPI,
  getTrackingTask as getTrackingTaskAPI,
  listTrackingTasks as listTrackingTasksAPI,
  getActiveTracking as getActiveTrackingAPI,
  listDetections as listDetectionsAPI
} from '@/api/tracking'
import type {
  LockTargetRequest,
  TrackingTask,
  DetectionTarget,
  TrackingStatus,
  BoundingBox
} from '@/types'

interface TrackingState {
  tasks: TrackingTask[]
  activeTask: TrackingTask | null
  detections: DetectionTarget[]
  selectedBbox: BoundingBox | null
  isDrawing: boolean
  loading: boolean
  listLoading: boolean
  detectionLoading: boolean
  total: number
  error: string | null
}

const initialState: TrackingState = {
  tasks: [],
  activeTask: null,
  detections: [],
  selectedBbox: null,
  isDrawing: false,
  loading: false,
  listLoading: false,
  detectionLoading: false,
  total: 0,
  error: null
}

export const fetchTrackingTasks = createAsyncThunk(
  'tracking/fetchTasks',
  async (params: { page?: number; pageSize?: number; uav_id?: string; status?: string }) => {
    const result = await listTrackingTasksAPI(params)
    return result
  }
)

export const fetchActiveTracking = createAsyncThunk(
  'tracking/fetchActive',
  async (uavId: string) => {
    const result = await getActiveTrackingAPI(uavId)
    return result
  }
)

export const fetchTrackingDetail = createAsyncThunk(
  'tracking/fetchDetail',
  async (id: string) => {
    const result = await getTrackingTaskAPI(id)
    return result
  }
)

export const lockTarget = createAsyncThunk(
  'tracking/lockTarget',
  async (req: LockTargetRequest) => {
    const result = await lockTargetAPI(req)
    return result
  }
)

export const stopTracking = createAsyncThunk(
  'tracking/stopTracking',
  async (id: string) => {
    await stopTrackingAPI(id)
    return id
  }
)

export const fetchDetections = createAsyncThunk(
  'tracking/fetchDetections',
  async ({ uavId, page, pageSize }: { uavId: string; page?: number; pageSize?: number }) => {
    const result = await listDetectionsAPI(uavId, { page, pageSize })
    return result
  }
)

const trackingSlice = createSlice({
  name: 'tracking',
  initialState,
  reducers: {
    setSelectedBbox: (state, action: PayloadAction<BoundingBox | null>) => {
      state.selectedBbox = action.payload
    },
    setIsDrawing: (state, action: PayloadAction<boolean>) => {
      state.isDrawing = action.payload
    },
    updateDetections: (state, action: PayloadAction<DetectionTarget[]>) => {
      state.detections = action.payload
    },
    updateActiveTask: (state, action: PayloadAction<TrackingTask | null>) => {
      state.activeTask = action.payload
    },
    resetTracking: () => initialState
  },
  extraReducers: builder => {
    builder
      .addCase(fetchTrackingTasks.pending, state => {
        state.listLoading = true
        state.error = null
      })
      .addCase(fetchTrackingTasks.fulfilled, (state, action) => {
        state.listLoading = false
        state.tasks = action.payload.list
        state.total = action.payload.total
      })
      .addCase(fetchTrackingTasks.rejected, (state, action) => {
        state.listLoading = false
        state.error = action.error.message || '获取任务列表失败'
      })

      .addCase(fetchActiveTracking.pending, state => {
        state.loading = true
      })
      .addCase(fetchActiveTracking.fulfilled, (state, action) => {
        state.loading = false
        state.activeTask = action.payload
      })
      .addCase(fetchActiveTracking.rejected, state => {
        state.loading = false
        state.activeTask = null
      })

      .addCase(lockTarget.pending, state => {
        state.loading = true
        state.error = null
      })
      .addCase(lockTarget.fulfilled, (state, action) => {
        state.loading = false
        state.activeTask = action.payload
        state.selectedBbox = null
        state.isDrawing = false
      })
      .addCase(lockTarget.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || '锁定目标失败'
      })

      .addCase(stopTracking.fulfilled, state => {
        state.activeTask = null
      })

      .addCase(fetchDetections.pending, state => {
        state.detectionLoading = true
      })
      .addCase(fetchDetections.fulfilled, (state, action) => {
        state.detectionLoading = false
        state.detections = action.payload.list
      })
      .addCase(fetchDetections.rejected, state => {
        state.detectionLoading = false
      })
  }
})

export const {
  setSelectedBbox,
  setIsDrawing,
  updateDetections,
  updateActiveTask,
  resetTracking
} = trackingSlice.actions

export default trackingSlice.reducer
