import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import { login as loginApi, logout as logoutApi, getCurrentUser, refreshToken as refreshTokenApi } from '@/api/auth'
import type { AuthState, LoginRequest, LoginResponse, UserInfo } from '@/types'

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  accessToken: localStorage.getItem('accessToken'),
  refreshToken: localStorage.getItem('refreshToken'),
  loading: false,
  error: null
}

export const login = createAsyncThunk<LoginResponse, LoginRequest>(
  'auth/login',
  async (credentials, { rejectWithValue }) => {
    try {
      const response = await loginApi(credentials)
      localStorage.setItem('accessToken', response.accessToken)
      localStorage.setItem('refreshToken', response.refreshToken)
      return response
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '登录失败')
    }
  }
)

export const logout = createAsyncThunk(
  'auth/logout',
  async (_, { rejectWithValue }) => {
    try {
      await logoutApi()
    } catch (error) {
      console.error('Logout API error:', error)
    } finally {
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
    }
    return null
  }
)

export const fetchCurrentUser = createAsyncThunk<UserInfo>(
  'auth/fetchCurrentUser',
  async (_, { rejectWithValue }) => {
    try {
      return await getCurrentUser()
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : '获取用户信息失败')
    }
  }
)

export const refreshAccessToken = createAsyncThunk<{ accessToken: string; refreshToken: string }, string>(
  'auth/refreshToken',
  async (token, { rejectWithValue }) => {
    try {
      const response = await refreshTokenApi(token)
      localStorage.setItem('accessToken', response.accessToken)
      localStorage.setItem('refreshToken', response.refreshToken)
      return response
    } catch (error) {
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
      return rejectWithValue(error instanceof Error ? error.message : '刷新令牌失败')
    }
  }
)

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null
    },
    setTokens: (state, action: PayloadAction<{ accessToken: string; refreshToken: string }>) => {
      state.accessToken = action.payload.accessToken
      state.refreshToken = action.payload.refreshToken
      state.isAuthenticated = true
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(login.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.accessToken = action.payload.accessToken
        state.refreshToken = action.payload.refreshToken
        state.user = action.payload.user
      })
      .addCase(login.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(logout.fulfilled, (state) => {
        state.isAuthenticated = false
        state.user = null
        state.accessToken = null
        state.refreshToken = null
        state.error = null
      })
      .addCase(fetchCurrentUser.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchCurrentUser.fulfilled, (state, action) => {
        state.loading = false
        state.user = action.payload
        state.isAuthenticated = true
      })
      .addCase(fetchCurrentUser.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
        state.isAuthenticated = false
      })
      .addCase(refreshAccessToken.fulfilled, (state, action) => {
        state.accessToken = action.payload.accessToken
        state.refreshToken = action.payload.refreshToken
      })
      .addCase(refreshAccessToken.rejected, (state) => {
        state.isAuthenticated = false
        state.user = null
        state.accessToken = null
        state.refreshToken = null
      })
  }
})

export const { clearError, setTokens } = authSlice.actions
export default authSlice.reducer
