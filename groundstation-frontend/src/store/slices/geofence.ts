import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import { getGeofenceList, getGeofenceDetail, createGeofence, updateGeofence, deleteGeofence, toggleGeofence, getViolationList, getRestrictionZones } from '@/api/geofence'
import type { Geofence, GeofenceViolation, FlightRestrictionZone, GeofenceType, GeofenceAction, GeofenceShape, PageResult } from '@/types'

interface GeofenceState {
  geofences: Geofence[]
  selectedGeofence: Geofence | null
  violations: GeofenceViolation[]
  restrictionZones: FlightRestrictionZone[]
  loading: boolean
  listLoading: boolean
  violationLoading: boolean
  error: string | null
  total: number
  violationTotal: number
  drawMode: boolean
  drawType: GeofenceType | null
  editingGeofence: Geofence | null
}

const initialState: GeofenceState = {
  geofences: [],
  selectedGeofence: null,
  violations: [],
  restrictionZones: [],
  loading: false,
  listLoading: false,
  violationLoading: false,
  error: null,
  total: 0,
  violationTotal: 0,
  drawMode: false,
  drawType: null,
  editingGeofence: null
}

export const fetchGeofences = createAsyncThunk<PageResult<Geofence>, { page?: number; pageSize?: number; keyword?: string; isEnabled?: boolean }>(
  'geofence/fetchList',
  async (params, { rejectWithValue }) => {
    try {
      return await getGeofenceList(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取围栏列表失败')
    }
  }
)

export const fetchGeofenceDetail = createAsyncThunk<Geofence, string>(
  'geofence/fetchDetail',
  async (id, { rejectWithValue }) => {
    try {
      return await getGeofenceDetail(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取围栏详情失败')
    }
  }
)

export const createNewGeofence = createAsyncThunk<Geofence, Partial<Geofence>>(
  'geofence/create',
  async (data, { rejectWithValue }) => {
    try {
      return await createGeofence(data)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '创建围栏失败')
    }
  }
)

export const updateExistingGeofence = createAsyncThunk<Geofence, { id: string; data: Partial<Geofence> }>(
  'geofence/update',
  async ({ id, data }, { rejectWithValue }) => {
    try {
      return await updateGeofence(id, data)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '更新围栏失败')
    }
  }
)

export const deleteGeofenceById = createAsyncThunk<void, string>(
  'geofence/delete',
  async (id, { rejectWithValue }) => {
    try {
      await deleteGeofence(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '删除围栏失败')
    }
  }
)

export const toggleGeofenceById = createAsyncThunk<void, { id: string; enabled: boolean }>(
  'geofence/toggle',
  async ({ id, enabled }, { rejectWithValue }) => {
    try {
      await toggleGeofence(id, enabled)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '切换围栏状态失败')
    }
  }
)

export const fetchViolations = createAsyncThunk<PageResult<GeofenceViolation>, { page?: number; pageSize?: number; geofenceId?: string; uavId?: string; resolved?: boolean }>(
  'geofence/fetchViolations',
  async (params, { rejectWithValue }) => {
    try {
      return await getViolationList(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取违规记录失败')
    }
  }
)

export const fetchRestrictionZones = createAsyncThunk<FlightRestrictionZone[], { lat: number; lng: number; radius: number }>(
  'geofence/fetchRestrictionZones',
  async (params, { rejectWithValue }) => {
    try {
      return await getRestrictionZones(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取禁飞区失败')
    }
  }
)

const geofenceSlice = createSlice({
  name: 'geofence',
  initialState,
  reducers: {
    selectGeofence: (state, action: PayloadAction<Geofence | null>) => {
      state.selectedGeofence = action.payload
    },
    setDrawMode: (state, action: PayloadAction<boolean>) => {
      state.drawMode = action.payload
    },
    setDrawType: (state, action: PayloadAction<GeofenceType | null>) => {
      state.drawType = action.payload
    },
    setEditingGeofence: (state, action: PayloadAction<Geofence | null>) => {
      state.editingGeofence = action.payload
    },
    updateGeofenceShape: (state, action: PayloadAction<{ id: string; shape: GeofenceShape }>) => {
      const geofence = state.geofences.find(g => g.id === action.payload.id)
      if (geofence) {
        geofence.shape = action.payload.shape
      }
      if (state.selectedGeofence?.id === action.payload.id) {
        state.selectedGeofence.shape = action.payload.shape
      }
    },
    updateGeofenceColor: (state, action: PayloadAction<{ id: string; color: string }>) => {
      const geofence = state.geofences.find(g => g.id === action.payload.id)
      if (geofence) {
        geofence.color = action.payload.color
      }
    },
    updateGeofenceAction: (state, action: PayloadAction<{ id: string; fenceAction: GeofenceAction }>) => {
      const geofence = state.geofences.find(g => g.id === action.payload.id)
      if (geofence) {
        geofence.action = action.payload.fenceAction
      }
    },
    addViolation: (state, action: PayloadAction<GeofenceViolation>) => {
      state.violations.unshift(action.payload)
    },
    resolveViolationLocally: (state, action: PayloadAction<string>) => {
      const violation = state.violations.find(v => v.id === action.payload)
      if (violation) {
        violation.resolved = true
        violation.resolvedAt = Date.now()
      }
    },
    clearGeofenceError: (state) => {
      state.error = null
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchGeofences.pending, (state) => {
        state.listLoading = true
      })
      .addCase(fetchGeofences.fulfilled, (state, action) => {
        state.listLoading = false
        state.geofences = action.payload.list
        state.total = action.payload.total
      })
      .addCase(fetchGeofences.rejected, (state, action) => {
        state.listLoading = false
        state.error = action.payload as string
      })
      .addCase(fetchGeofenceDetail.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchGeofenceDetail.fulfilled, (state, action) => {
        state.loading = false
        state.selectedGeofence = action.payload
      })
      .addCase(fetchGeofenceDetail.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(createNewGeofence.fulfilled, (state, action) => {
        state.geofences.unshift(action.payload)
        state.selectedGeofence = action.payload
        state.drawMode = false
        state.drawType = null
        state.editingGeofence = null
      })
      .addCase(updateExistingGeofence.fulfilled, (state, action) => {
        const index = state.geofences.findIndex(g => g.id === action.payload.id)
        if (index !== -1) {
          state.geofences[index] = action.payload
        }
        state.selectedGeofence = action.payload
        state.editingGeofence = null
      })
      .addCase(deleteGeofenceById.fulfilled, (state, action) => {
        state.geofences = state.geofences.filter(g => g.id !== action.meta.arg)
        if (state.selectedGeofence?.id === action.meta.arg) {
          state.selectedGeofence = null
        }
      })
      .addCase(toggleGeofenceById.fulfilled, (state, action) => {
        const geofence = state.geofences.find(g => g.id === action.meta.arg.id)
        if (geofence) {
          geofence.isEnabled = action.meta.arg.enabled
        }
        if (state.selectedGeofence?.id === action.meta.arg.id) {
          state.selectedGeofence.isEnabled = action.meta.arg.enabled
        }
      })
      .addCase(fetchViolations.pending, (state) => {
        state.violationLoading = true
      })
      .addCase(fetchViolations.fulfilled, (state, action) => {
        state.violationLoading = false
        state.violations = action.payload.list
        state.violationTotal = action.payload.total
      })
      .addCase(fetchViolations.rejected, (state, action) => {
        state.violationLoading = false
        state.error = action.payload as string
      })
      .addCase(fetchRestrictionZones.fulfilled, (state, action) => {
        state.restrictionZones = action.payload
      })
  }
})

export const {
  selectGeofence,
  setDrawMode,
  setDrawType,
  setEditingGeofence,
  updateGeofenceShape,
  updateGeofenceColor,
  updateGeofenceAction,
  addViolation,
  resolveViolationLocally,
  clearGeofenceError
} = geofenceSlice.actions

export default geofenceSlice.reducer
