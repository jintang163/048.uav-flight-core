import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type { LinkStatus } from '@/types'

interface LinkState {
  currentLinkStatus: LinkStatus | null
  linkHistory: LinkStatus[]
  switchCount: number
  maxHistoryPoints: number
  isSwitching: boolean
  lastSwitchTime: number | null
  previousActiveLink: number | null
}

const initialState: LinkState = {
  currentLinkStatus: null,
  linkHistory: [],
  switchCount: 0,
  maxHistoryPoints: 100,
  isSwitching: false,
  lastSwitchTime: null,
  previousActiveLink: null
}

const linkSlice = createSlice({
  name: 'link',
  initialState,
  reducers: {
    updateLinkStatus: (state, action: PayloadAction<LinkStatus>) => {
      const newStatus = action.payload
      const previousActiveLink = state.currentLinkStatus?.active_link ?? null

      if (previousActiveLink !== null && previousActiveLink !== newStatus.active_link) {
        state.switchCount += 1
        state.isSwitching = true
        state.lastSwitchTime = Date.now()
        state.previousActiveLink = previousActiveLink
      }

      state.currentLinkStatus = newStatus

      state.linkHistory.push(newStatus)

      if (state.linkHistory.length > state.maxHistoryPoints) {
        state.linkHistory = state.linkHistory.slice(-state.maxHistoryPoints)
      }
    },
    clearLinkStatus: (state) => {
      state.currentLinkStatus = null
      state.linkHistory = []
      state.switchCount = 0
      state.isSwitching = false
      state.lastSwitchTime = null
      state.previousActiveLink = null
    },
    setSwitchingComplete: (state) => {
      state.isSwitching = false
    }
  }
})

export const { updateLinkStatus, clearLinkStatus, setSwitchingComplete } = linkSlice.actions

export const selectCurrentLink = (state: { link: LinkState }): LinkStatus | null =>
  state.link.currentLinkStatus

export const selectLinkHistory = (state: { link: LinkState }): LinkStatus[] =>
  state.link.linkHistory

export const selectIsLinkSwitching = (state: { link: LinkState }): boolean =>
  state.link.isSwitching

export const selectLinkSwitchCount = (state: { link: LinkState }): number =>
  state.link.switchCount

export const selectLastSwitchTime = (state: { link: LinkState }): number | null =>
  state.link.lastSwitchTime

export default linkSlice.reducer
