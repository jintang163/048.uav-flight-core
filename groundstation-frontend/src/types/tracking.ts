export enum DetectionClass {
  PERSON = 'person',
  CAR = 'car',
  TRUCK = 'truck',
  BUS = 'bus',
  MOTORCYCLE = 'motorcycle',
  BICYCLE = 'bicycle',
  DOG = 'dog',
  UNKNOWN = 'unknown'
}

export enum TrackingStatus {
  IDLE = 'idle',
  LOCKING = 'locking',
  TRACKING = 'tracking',
  SEARCHING = 'searching',
  LOST = 'lost',
  COMPLETED = 'completed'
}

export interface DetectionTarget {
  id: string
  uav_id: string
  class: DetectionClass
  class_name: string
  confidence: number
  bbox_x: number
  bbox_y: number
  bbox_width: number
  bbox_height: number
  frame_width: number
  frame_height: number
  latitude?: number
  longitude?: number
  altitude?: number
  image_path?: string
  track_id?: string
  created_at: string
}

export interface TrackingTask {
  id: string
  uav_id: string
  name?: string
  target_class: DetectionClass
  status: TrackingStatus
  initial_bbox_x: number
  initial_bbox_y: number
  initial_bbox_width: number
  initial_bbox_height: number
  current_bbox_x?: number
  current_bbox_y?: number
  current_bbox_width?: number
  current_bbox_height?: number
  center_offset_x?: number
  center_offset_y?: number
  search_radius: number
  max_search_radius: number
  confidence?: number
  frames_visible: number
  frames_lost: number
  target_latitude?: number
  target_longitude?: number
  start_time?: string
  end_time?: string
  created_by?: string
  created_at: string
  updated_at: string
  uav?: {
    id: string
    name: string
    status: string
  }
}

export interface LockTargetRequest {
  uav_id: string
  bbox_x: number
  bbox_y: number
  bbox_width: number
  bbox_height: number
  frame_width?: number
  frame_height?: number
  target_class?: DetectionClass
  name?: string
  search_radius?: number
  max_radius?: number
}

export interface BoundingBox {
  x: number
  y: number
  width: number
  height: number
}

export const DetectionClassLabels: Record<DetectionClass, string> = {
  [DetectionClass.PERSON]: '人员',
  [DetectionClass.CAR]: '轿车',
  [DetectionClass.TRUCK]: '卡车',
  [DetectionClass.BUS]: '公交车',
  [DetectionClass.MOTORCYCLE]: '摩托车',
  [DetectionClass.BICYCLE]: '自行车',
  [DetectionClass.DOG]: '狗',
  [DetectionClass.UNKNOWN]: '未知'
}

export const DetectionClassColors: Record<DetectionClass, string> = {
  [DetectionClass.PERSON]: '#52c41a',
  [DetectionClass.CAR]: '#1890ff',
  [DetectionClass.TRUCK]: '#722ed1',
  [DetectionClass.BUS]: '#eb2f96',
  [DetectionClass.MOTORCYCLE]: '#fa8c16',
  [DetectionClass.BICYCLE]: '#faad14',
  [DetectionClass.DOG]: '#a0d911',
  [DetectionClass.UNKNOWN]: '#8c8c8c'
}

export const TrackingStatusLabels: Record<TrackingStatus, string> = {
  [TrackingStatus.IDLE]: '待命',
  [TrackingStatus.LOCKING]: '锁定中',
  [TrackingStatus.TRACKING]: '追踪中',
  [TrackingStatus.SEARCHING]: '搜索中',
  [TrackingStatus.LOST]: '丢失',
  [TrackingStatus.COMPLETED]: '已完成'
}

export const TrackingStatusColors: Record<TrackingStatus, string> = {
  [TrackingStatus.IDLE]: 'default',
  [TrackingStatus.LOCKING]: 'processing',
  [TrackingStatus.TRACKING]: 'success',
  [TrackingStatus.SEARCHING]: 'warning',
  [TrackingStatus.LOST]: 'error',
  [TrackingStatus.COMPLETED]: 'default'
}
