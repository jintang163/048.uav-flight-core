import { configureStore, type ThunkAction, type Action } from '@reduxjs/toolkit'
import { useDispatch, useSelector, type TypedUseSelectorHook } from 'react-redux'
import authReducer from './slices/auth'
import uavReducer from './slices/uav'
import telemetryReducer from './slices/telemetry'
import missionReducer from './slices/mission'
import alertReducer from './slices/alert'
import geofenceReducer from './slices/geofence'
import formationReducer from './slices/formation'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    uav: uavReducer,
    telemetry: telemetryReducer,
    mission: missionReducer,
    alert: alertReducer,
    geofence: geofenceReducer,
    formation: formationReducer
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
        ignoredPaths: ['register', 'rehydrate']
      }
    }),
  devTools: import.meta.env.MODE !== 'production'
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
export type AppThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, Action<string>>

export const useAppDispatch: () => AppDispatch = useDispatch
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector

export default store
