import { getWebSocketClient } from '../client'
import { addCollisionAlert, updatePosition, updateIntersections } from '@/store/slices/collision'
import type { CollisionAlert, RouteIntersection, UAVLivePosition } from '@/types'
import store from '@/store'

export const initCollisionWebSocket = (): void => {
  const wsClient = getWebSocketClient()
  if (!wsClient) return

  wsClient.on('collision_alert', (data: unknown) => {
    const payload = data as Record<string, unknown>
    if (payload.alert_id) {
      const alert = {
        id: 0,
        alert_id: payload.alert_id as string,
        uav_id_1: payload.uav_id_1 as number,
        uav_id_2: payload.uav_id_2 as number,
        risk_level: payload.risk_level as CollisionAlert['risk_level'],
        current_distance: payload.current_distance as number,
        min_distance: payload.min_distance as number,
        time_to_collision: payload.time_to_collision as number,
        alert_type: payload.alert_type as string,
        action_taken: payload.action_taken as CollisionAlert['action_taken'],
        action_detail: payload.action_detail as string,
        is_resolved: false,
        resolved_at: null,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      store.dispatch(addCollisionAlert(alert))
    }
  })

  wsClient.on('collision_resolved', (data: unknown) => {
    // 已在 addCollisionAlert 中管理，这里可以触发刷新
  })

  wsClient.on('avoidance_decision', (_data: unknown) => {
    // 避让决策更新
  })

  wsClient.on('route_intersections', (data: unknown) => {
    const payload = data as { intersections?: RouteIntersection[]; count?: number }
    if (payload.intersections) {
      store.dispatch(updateIntersections(payload.intersections))
    }
  })
}
