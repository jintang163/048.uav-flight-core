import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type {
  ThrustLearningStatus,
  ThrustCurvePoint,
  PIDGainProfile,
  ThrustLearningSample
} from '@/types/thrust-learning'

interface ThrustLearningState {
  status: ThrustLearningStatus | null;
  thrustCurve: ThrustCurvePoint[];
  pidGains: PIDGainProfile | null;
  samples: ThrustLearningSample[];
  loading: boolean;
  error: string | null;
}

const initialState: ThrustLearningState = {
  status: null,
  thrustCurve: [],
  pidGains: null,
  samples: [],
  loading: false,
  error: null
}

const thrustLearningSlice = createSlice({
  name: 'thrustLearning',
  initialState,
  reducers: {
    setLearningStatus(state, action: PayloadAction<ThrustLearningStatus>) {
      state.status = action.payload
    },
    setThrustCurve(state, action: PayloadAction<ThrustCurvePoint[]>) {
      state.thrustCurve = action.payload
    },
    setPIDGains(state, action: PayloadAction<PIDGainProfile>) {
      state.pidGains = action.payload
    },
    setSamples(state, action: PayloadAction<ThrustLearningSample[]>) {
      state.samples = action.payload
    },
    appendSample(state, action: PayloadAction<ThrustLearningSample>) {
      state.samples.push(action.payload)
      if (state.samples.length > 500) {
        state.samples = state.samples.slice(-500)
      }
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
  setLearningStatus,
  setThrustCurve,
  setPIDGains,
  setSamples,
  appendSample,
  setLoading,
  setError
} = thrustLearningSlice.actions

export default thrustLearningSlice.reducer
