import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import {
  getFormationList,
  getFormationDetail,
  createFormation,
  updateFormation,
  deleteFormation,
  getActiveFormations,
  startFormation,
  pauseFormation,
  resumeFormation,
  stopFormation
} from '@/api/formation'
import type {
  Formation,
  FormationStatus,
  FormationType,
  FormationCollisionWarning,
  FormationLightConfig,
  CreateFormationRequest,
  UpdateFormationRequest
} from '@/types'

interface FormationState {
  formations: Formation[]
  currentFormation: Formation | null
  selectedFormationId: string | null
  activeFormations: Formation[]
  collisionWarnings: FormationCollisionWarning[]
  lightConfig: FormationLightConfig | null
  loading: boolean
  listLoading: boolean
  error: string | null
  total: number
}

const initialState: FormationState = {
  formations: [],
  currentFormation: null,
  selectedFormationId: null,
  activeFormations: [],
  collisionWarnings: [],
  lightConfig: null,
  loading: false,
  listLoading: false,
  error: null,
  total: 0
}

export const fetchFormationList = createAsyncThunk<
  { list: Formation[]; total: number },
  { page?: number; pageSize?: number }
>('formation/fetchList', async (params, { rejectWithValue }) => {
  try {
    return await getFormationList(params)
  } catch (error) {
    return rejectWithValue(error instanceof Error ? error.message : '获取编队列表失败')
  }
})

export const fetchFormationDetail = createAsyncThunk<Formation, string>(
  'formation/fetchDetail',
  async (id, { rejectWithValue }) => {
    try {
      return await getFormationDetail(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取编队详情失败')
    }
  }
)

export const fetchActiveFormations = createAsyncThunk<Formation[]>(
  'formation/fetchActive',
  async (_, { rejectWithValue }) => {
    try {
      return await getActiveFormations()
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取活动编队失败')
    }
  }
)

export const createNewFormation = createAsyncThunk<Formation, CreateFormationRequest>(
  'formation/create',
  async (data, { rejectWithValue }) => {
    try {
      return await createFormation(data)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '创建编队失败')
    }
  }
)

export const updateExistingFormation = createAsyncThunk<
  Formation,
  { id: string; data: UpdateFormationRequest }
>('formation/update', async ({ id, data }, { rejectWithValue }) => {
  try {
    return await updateFormation(id, data)
  } catch (error) {
    return rejectWithValue(error instanceof Error ? error.message : '更新编队失败')
  }
})

export const deleteFormationById = createAsyncThunk<void, string>(
  'formation/delete',
  async (id, { rejectWithValue }) => {
    try {
      await deleteFormation(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '删除编队失败')
    }
  }
)

export const startFormationById = createAsyncThunk<void, string>(
  'formation/start',
  async (id, { rejectWithValue }) => {
    try {
      await startFormation(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '启动编队失败')
    }
  }
)

export const pauseFormationById = createAsyncThunk<void, string>(
  'formation/pause',
  async (id, { rejectWithValue }) => {
    try {
      await pauseFormation(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '暂停编队失败')
    }
  }
)

export const resumeFormationById = createAsyncThunk<void, string>(
  'formation/resume',
  async (id, { rejectWithValue }) => {
    try {
      await resumeFormation(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '恢复编队失败')
    }
  }
)

export const stopFormationById = createAsyncThunk<void, string>(
  'formation/stop',
  async (id, { rejectWithValue }) => {
    try {
      await stopFormation(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '停止编队失败')
    }
  }
)

const formationSlice = createSlice({
  name: 'formation',
  initialState,
  reducers: {
    selectFormation: (state, action: PayloadAction<string | null>) => {
      state.selectedFormationId = action.payload
    },
    updateFormationStatus: (
      state,
      action: PayloadAction<{ id: string; status: FormationStatus }>
    ) => {
      const formation = state.formations.find(f => f.id === action.payload.id)
      if (formation) {
        formation.status = action.payload.status
      }
      if (state.currentFormation && state.currentFormation.id === action.payload.id) {
        state.currentFormation.status = action.payload.status
      }
      const activeFormation = state.activeFormations.find(f => f.id === action.payload.id)
      if (activeFormation) {
        activeFormation.status = action.payload.status
      }
    },
    updateFormationRealtime: (
      state,
      action: PayloadAction<Partial<Formation> & { id: string }>
    ) => {
      if (state.currentFormation && state.currentFormation.id === action.payload.id) {
        state.currentFormation = { ...state.currentFormation, ...action.payload }
      }
      const formation = state.formations.find(f => f.id === action.payload.id)
      if (formation) {
        Object.assign(formation, action.payload)
      }
    },
    addCollisionWarning: (state, action: PayloadAction<FormationCollisionWarning>) => {
      state.collisionWarnings.unshift(action.payload)
      if (state.collisionWarnings.length > 100) {
        state.collisionWarnings = state.collisionWarnings.slice(0, 100)
      }
    },
    setLightConfig: (state, action: PayloadAction<FormationLightConfig>) => {
      state.lightConfig = action.payload
    },
    clearFormationError: state => {
      state.error = null
    }
  },
  extraReducers: builder => {
    builder
      .addCase(fetchFormationList.pending, state => {
        state.listLoading = true
      })
      .addCase(fetchFormationList.fulfilled, (state, action) => {
        state.listLoading = false
        state.formations = action.payload.list
        state.total = action.payload.total
      })
      .addCase(fetchFormationList.rejected, (state, action) => {
        state.listLoading = false
        state.error = action.payload as string
      })
      .addCase(fetchFormationDetail.pending, state => {
        state.loading = true
      })
      .addCase(fetchFormationDetail.fulfilled, (state, action) => {
        state.loading = false
        state.currentFormation = action.payload
      })
      .addCase(fetchFormationDetail.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(fetchActiveFormations.fulfilled, (state, action) => {
        state.activeFormations = action.payload
      })
      .addCase(createNewFormation.fulfilled, (state, action) => {
        state.formations.unshift(action.payload)
        state.total += 1
      })
      .addCase(updateExistingFormation.fulfilled, (state, action) => {
        const index = state.formations.findIndex(f => f.id === action.payload.id)
        if (index !== -1) {
          state.formations[index] = action.payload
        }
        if (state.currentFormation && state.currentFormation.id === action.payload.id) {
          state.currentFormation = action.payload
        }
      })
      .addCase(deleteFormationById.fulfilled, (state, action) => {
        state.formations = state.formations.filter(f => f.id !== action.meta.arg)
        state.total -= 1
        if (state.selectedFormationId === action.meta.arg) {
          state.selectedFormationId = null
          state.currentFormation = null
        }
      })
      .addCase(startFormationById.fulfilled, state => {
        if (state.currentFormation) {
          state.currentFormation.status = FormationStatus.EXECUTING
        }
      })
      .addCase(pauseFormationById.fulfilled, state => {
        if (state.currentFormation) {
          state.currentFormation.status = FormationStatus.PAUSED
        }
      })
      .addCase(resumeFormationById.fulfilled, state => {
        if (state.currentFormation) {
          state.currentFormation.status = FormationStatus.EXECUTING
        }
      })
      .addCase(stopFormationById.fulfilled, state => {
        if (state.currentFormation) {
          state.currentFormation.status = FormationStatus.IDLE
        }
      })
  }
})

export const {
  selectFormation,
  updateFormationStatus,
  updateFormationRealtime,
  addCollisionWarning,
  setLightConfig,
  clearFormationError
} = formationSlice.actions

export default formationSlice.reducer
