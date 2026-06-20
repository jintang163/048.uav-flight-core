import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import { getUAVList, getUAVDetail } from '@/api/uav'
import type { UAV, UAVListItem, UAVStatus, UAVMode } from '@/types'

interface UAVState {
  currentUAV: UAV | null
  uavList: UAVListItem[]
  selectedUAVId: string | null
  loading: boolean
  listLoading: boolean
  error: string | null
  total: number
}

const initialState: UAVState = {
  currentUAV: null,
  uavList: [],
  selectedUAVId: null,
  loading: false,
  listLoading: false,
  error: null,
  total: 0
}

export const fetchUAVList = createAsyncThunk<{ list: UAVListItem[]; total: number }, { page?: number; pageSize?: number; keyword?: string; status?: string }>(
  'uav/fetchList',
  async (params, { rejectWithValue }) => {
    try {
      return await getUAVList(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取无人机列表失败')
    }
  }
)

export const fetchUAVDetail = createAsyncThunk<UAV, string>(
  'uav/fetchDetail',
  async (id, { rejectWithValue }) => {
    try {
      return await getUAVDetail(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取无人机详情失败')
    }
  }
)

const uavSlice = createSlice({
  name: 'uav',
  initialState,
  reducers: {
    selectUAV: (state, action: PayloadAction<string | null>) => {
      state.selectedUAVId = action.payload
    },
    updateUAVStatus: (state, action: PayloadAction<{ id: string; status: UAVStatus }>) => {
      const uav = state.uavList.find(u => u.id === action.payload.id)
      if (uav) {
        uav.status = action.payload.status
      }
      if (state.currentUAV && state.currentUAV.id === action.payload.id) {
        state.currentUAV.status = action.payload.status
      }
    },
    updateUAVMode: (state, action: PayloadAction<{ id: string; mode: UAVMode }>) => {
      if (state.currentUAV && state.currentUAV.id === action.payload.id) {
        state.currentUAV.mode = action.payload.mode
      }
    },
    updateUAVRealtime: (state, action: PayloadAction<Partial<UAV> & { id: string }>) => {
      if (state.currentUAV && state.currentUAV.id === action.payload.id) {
        state.currentUAV = { ...state.currentUAV, ...action.payload }
      }
    },
    updateUAVBattery: (state, action: PayloadAction<{ id: string; remaining: number; voltage: number }>) => {
      const uav = state.uavList.find(u => u.id === action.payload.id)
      if (uav) {
        uav.battery = action.payload.remaining
      }
    },
    clearUAVError: (state) => {
      state.error = null
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchUAVList.pending, (state) => {
        state.listLoading = true
      })
      .addCase(fetchUAVList.fulfilled, (state, action) => {
        state.listLoading = false
        state.uavList = action.payload.list
        state.total = action.payload.total
      })
      .addCase(fetchUAVList.rejected, (state, action) => {
        state.listLoading = false
        state.error = action.payload as string
      })
      .addCase(fetchUAVDetail.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchUAVDetail.fulfilled, (state, action) => {
        state.loading = false
        state.currentUAV = action.payload
      })
      .addCase(fetchUAVDetail.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
  }
})

export const { selectUAV, updateUAVStatus, updateUAVMode, updateUAVRealtime, updateUAVBattery, clearUAVError } = uavSlice.actions
export default uavSlice.reducer
