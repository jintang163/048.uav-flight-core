import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import { getMissionList, getMissionDetail, createMission, updateMission, deleteMission, addWaypoint, updateWaypoint, deleteWaypoint, reorderWaypoints } from '@/api/mission'
import type { Mission, Waypoint, WaypointAction, MissionExecutionState, MissionStatus, PageResult } from '@/types'

interface MissionState {
  missions: Mission[]
  currentMission: Mission | null
  selectedMissionId: string | null
  executionState: MissionExecutionState | null
  loading: boolean
  listLoading: boolean
  error: string | null
  total: number
  waypoints: Waypoint[]
}

const initialState: MissionState = {
  missions: [],
  currentMission: null,
  selectedMissionId: null,
  executionState: null,
  loading: false,
  listLoading: false,
  error: null,
  total: 0,
  waypoints: []
}

export const fetchMissions = createAsyncThunk<PageResult<Mission>, { page?: number; pageSize?: number; keyword?: string; status?: string }>(
  'mission/fetchList',
  async (params, { rejectWithValue }) => {
    try {
      return await getMissionList(params)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取任务列表失败')
    }
  }
)

export const fetchMissionDetail = createAsyncThunk<Mission, string>(
  'mission/fetchDetail',
  async (id, { rejectWithValue }) => {
    try {
      return await getMissionDetail(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取任务详情失败')
    }
  }
)

export const createNewMission = createAsyncThunk<Mission, Partial<Mission>>(
  'mission/create',
  async (data, { rejectWithValue }) => {
    try {
      return await createMission(data)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '创建任务失败')
    }
  }
)

export const updateExistingMission = createAsyncThunk<Mission, { id: string; data: Partial<Mission> }>(
  'mission/update',
  async ({ id, data }, { rejectWithValue }) => {
    try {
      return await updateMission(id, data)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '更新任务失败')
    }
  }
)

export const deleteMissionById = createAsyncThunk<void, string>(
  'mission/delete',
  async (id, { rejectWithValue }) => {
    try {
      await deleteMission(id)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '删除任务失败')
    }
  }
)

const missionSlice = createSlice({
  name: 'mission',
  initialState,
  reducers: {
    selectMission: (state, action: PayloadAction<string | null>) => {
      state.selectedMissionId = action.payload
    },
    setCurrentMission: (state, action: PayloadAction<Mission | null>) => {
      state.currentMission = action.payload
      if (action.payload) {
        state.waypoints = action.payload.waypoints
      }
    },
    addWaypointToMission: (state, action: PayloadAction<Waypoint>) => {
      state.waypoints.push(action.payload)
      if (state.currentMission) {
        state.currentMission.waypoints = [...state.waypoints]
      }
    },
    updateWaypointInMission: (state, action: PayloadAction<Waypoint>) => {
      const index = state.waypoints.findIndex(w => w.id === action.payload.id)
      if (index !== -1) {
        state.waypoints[index] = action.payload
        if (state.currentMission) {
          state.currentMission.waypoints = [...state.waypoints]
        }
      }
    },
    removeWaypointFromMission: (state, action: PayloadAction<string>) => {
      state.waypoints = state.waypoints.filter(w => w.id !== action.payload.id)
      if (state.currentMission) {
        state.currentMission.waypoints = [...state.waypoints]
      }
    },
    reorderWaypointsInMission: (state, action: PayloadAction<string[]>) => {
      const reordered: Waypoint[] = []
      action.payload.forEach((id, index) => {
        const waypoint = state.waypoints.find(w => w.id === id)
        if (waypoint) {
          reordered.push({ ...waypoint, sequence: index + 1 })
        }
      })
      state.waypoints = reordered
      if (state.currentMission) {
        state.currentMission.waypoints = [...state.waypoints]
      }
    },
    setWaypointAction: (state, action: PayloadAction<{ waypointId: string; actionType: WaypointAction }>) => {
      const waypoint = state.waypoints.find(w => w.id === action.payload.waypointId)
      if (waypoint) {
        waypoint.action = action.payload.actionType
      }
    },
    setMissionStatus: (state, action: PayloadAction<{ missionId: string; status: MissionStatus }>) => {
      const mission = state.missions.find(m => m.id === action.payload.missionId)
      if (mission) {
        mission.status = action.payload.status
      }
      if (state.currentMission && state.currentMission.id === action.payload.missionId) {
        state.currentMission.status = action.payload.status
      }
    },
    updateExecutionState: (state, action: PayloadAction<MissionExecutionState>) => {
      state.executionState = action.payload
    },
    setCurrentWaypoint: (state, action: PayloadAction<number>) => {
      state.waypoints.forEach((w, index) => {
        w.isCurrent = index === action.payload
      })
    },
    setWaypointReached: (state, action: PayloadAction<number>) => {
      if (state.waypoints[action.payload]) {
        state.waypoints[action.payload].isReached = true
      }
    },
    clearMissionError: (state) => {
      state.error = null
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchMissions.pending, (state) => {
        state.listLoading = true
      })
      .addCase(fetchMissions.fulfilled, (state, action) => {
        state.listLoading = false
        state.missions = action.payload.list
        state.total = action.payload.total
      })
      .addCase(fetchMissions.rejected, (state, action) => {
        state.listLoading = false
        state.error = action.payload as string
      })
      .addCase(fetchMissionDetail.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchMissionDetail.fulfilled, (state, action) => {
        state.loading = false
        state.currentMission = action.payload
        state.waypoints = action.payload.waypoints
      })
      .addCase(fetchMissionDetail.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(createNewMission.fulfilled, (state, action) => {
        state.missions.unshift(action.payload)
        state.currentMission = action.payload
        state.waypoints = action.payload.waypoints
      })
      .addCase(updateExistingMission.fulfilled, (state, action) => {
        const index = state.missions.findIndex(m => m.id === action.payload.id)
        if (index !== -1) {
          state.missions[index] = action.payload
        }
        state.currentMission = action.payload
        state.waypoints = action.payload.waypoints
      })
      .addCase(deleteMissionById.fulfilled, (state, action) => {
        state.missions = state.missions.filter(m => m.id !== action.meta.arg)
        if (state.currentMission?.id === action.meta.arg) {
          state.currentMission = null
          state.waypoints = []
        }
      })
  }
})

export const {
  selectMission,
  setCurrentMission,
  addWaypointToMission,
  updateWaypointInMission,
  removeWaypointFromMission,
  reorderWaypointsInMission,
  setWaypointAction,
  setMissionStatus,
  updateExecutionState,
  setCurrentWaypoint,
  setWaypointReached,
  clearMissionError
} = missionSlice.actions

export default missionSlice.reducer
