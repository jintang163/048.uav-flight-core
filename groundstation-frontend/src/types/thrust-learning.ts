export type LearningState = 'idle' | 'weight_estimation' | 'data_collecting' | 'model_optimizing' | 'applied';

export interface ThrustLearningStatus {
  uav_id: number;
  state: LearningState;
  estimated_weight_kg: number;
  hover_throttle: number;
  sample_count: number;
  progress_pct: number;
  started_at?: string;
  completed_at?: string;
}

export interface ThrustCurvePoint {
  throttle: number;
  thrust_n: number;
  motor_rpm_avg: number;
  sample_count: number;
}

export interface PIDGainProfile {
  uav_id: number;
  profile_name: string;
  is_auto_tuned: boolean;
  roll_kp: number; roll_ki: number; roll_kd: number;
  pitch_kp: number; pitch_ki: number; pitch_kd: number;
  yaw_kp: number; yaw_ki: number; yaw_kd: number;
  rate_roll_kp: number; rate_roll_ki: number; rate_roll_kd: number;
  rate_pitch_kp: number; rate_pitch_ki: number; rate_pitch_kd: number;
  rate_yaw_kp: number; rate_yaw_ki: number; rate_yaw_kd: number;
  alt_kp: number; alt_ki: number; alt_kd: number;
}

export interface ThrustLearningSample {
  id: number;
  throttle: number;
  accel_z: number;
  altitude: number;
  vz: number;
  motor_pwm: number[];
  voltage: number;
  timestamp: number;
}

export const LEARNING_STATE_LABELS: Record<LearningState, string> = {
  idle: '空闲',
  weight_estimation: '重量估算中',
  data_collecting: '数据采集中',
  model_optimizing: '模型优化中',
  applied: '已应用'
};

export const LEARNING_STATE_COLORS: Record<LearningState, string> = {
  idle: '#999',
  weight_estimation: '#1890ff',
  data_collecting: '#52c41a',
  model_optimizing: '#faad14',
  applied: '#52c41a'
};
