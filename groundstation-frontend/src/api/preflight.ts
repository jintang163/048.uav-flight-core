import { http } from './http'
import type {
  PreflightCheckResult,
  PreflightCheckThresholds,
  RunPreflightRequest,
  BatchRunPreflightRequest,
  BatchPreflightResponse
} from '@/types'

export const runPreflightCheck = (data: RunPreflightRequest) =>
  http.post<PreflightCheckResult>('/preflight/run', data)

export const batchRunPreflightCheck = (data: BatchRunPreflightRequest) =>
  http.post<BatchPreflightResponse>('/preflight/batch', data)

export const getPreflightThresholds = () =>
  http.get<PreflightCheckThresholds>('/preflight/thresholds')
