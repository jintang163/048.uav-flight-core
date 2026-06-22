-- UAV Flight Control System Database Schema
-- Version: 1.0.0

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ======================================
-- 无人机基础信息表
-- ======================================
CREATE TABLE IF NOT EXISTS `uavs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` VARCHAR(64) NOT NULL COMMENT '无人机唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '无人机名称',
  `model` VARCHAR(64) DEFAULT NULL COMMENT '型号',
  `firmware_version` VARCHAR(32) DEFAULT NULL COMMENT '固件版本',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态:0-离线,1-在线,2-飞行中,3-故障',
  `battery_level` DECIMAL(5,2) DEFAULT 100.00 COMMENT '电量百分比',
  `last_heartbeat` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `home_latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '返航点纬度',
  `home_longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '返航点经度',
  `home_altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '返航点海拔',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uav_id` (`uav_id`),
  KEY `idx_status` (`status`),
  KEY `idx_last_heartbeat` (`last_heartbeat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='无人机基础信息表';

-- ======================================
-- 飞行状态表
-- ======================================
CREATE TABLE IF NOT EXISTS `flight_status` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` VARCHAR(64) NOT NULL COMMENT '无人机标识',
  `flight_mode` VARCHAR(32) DEFAULT NULL COMMENT '飞行模式:MANUAL,ALT_HOLD,POS_HOLD,AUTO,RTL,LAND',
  `latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '当前纬度',
  `longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '当前经度',
  `altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '当前海拔(米)',
  `relative_altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '相对高度(米)',
  `roll` DECIMAL(8,4) DEFAULT NULL COMMENT '横滚角(度)',
  `pitch` DECIMAL(8,4) DEFAULT NULL COMMENT '俯仰角(度)',
  `yaw` DECIMAL(8,4) DEFAULT NULL COMMENT '航向角(度)',
  `ground_speed` DECIMAL(8,2) DEFAULT NULL COMMENT '地速(m/s)',
  `air_speed` DECIMAL(8,2) DEFAULT NULL COMMENT '空速(m/s)',
  `climb_rate` DECIMAL(8,2) DEFAULT NULL COMMENT '爬升率(m/s)',
  `throttle` DECIMAL(5,2) DEFAULT NULL COMMENT '油门百分比',
  `voltage` DECIMAL(6,3) DEFAULT NULL COMMENT '电压(V)',
  `current` DECIMAL(6,3) DEFAULT NULL COMMENT '电流(A)',
  `battery_remaining` DECIMAL(5,2) DEFAULT NULL COMMENT '剩余电量(%)',
  `gps_satellites_visible` TINYINT DEFAULT NULL COMMENT 'GPS可见卫星数',
  `gps_fix_type` TINYINT DEFAULT 0 COMMENT 'GPS定位类型:0-无,1-2D,2-3D,3-DGPS,4-RTK',
  `signal_strength` TINYINT DEFAULT 100 COMMENT '信号强度(%)',
  `arm_state` TINYINT DEFAULT 0 COMMENT '解锁状态:0-上锁,1-解锁',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='飞行状态表';

-- ======================================
-- 航线模板表
-- ======================================
CREATE TABLE IF NOT EXISTS `mission_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL COMMENT '航线名称',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '描述',
  `creator_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人ID',
  `max_altitude` DECIMAL(8,2) DEFAULT 120.00 COMMENT '最大飞行高度(米)',
  `speed` DECIMAL(8,2) DEFAULT 5.00 COMMENT '巡航速度(m/s)',
  `rtl_altitude` DECIMAL(8,2) DEFAULT 30.00 COMMENT '返航高度(米)',
  `is_public` TINYINT DEFAULT 0 COMMENT '是否公开:0-否,1-是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_creator_id` (`creator_id`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='航线模板表';

-- ======================================
-- 航点表
-- ======================================
CREATE TABLE IF NOT EXISTS `mission_waypoints` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `mission_id` BIGINT UNSIGNED NOT NULL COMMENT '航线模板ID',
  `waypoint_index` INT NOT NULL COMMENT '航点序号',
  `latitude` DECIMAL(10,7) NOT NULL COMMENT '纬度',
  `longitude` DECIMAL(10,7) NOT NULL COMMENT '经度',
  `altitude` DECIMAL(8,2) NOT NULL COMMENT '海拔(米)',
  `action_type` VARCHAR(32) DEFAULT 'WAYPOINT' COMMENT '动作类型:WAYPOINT,HOVER,TAKE_PHOTO,TAKE_VIDEO,CONDITION_DELAY,LOITER_TURNS,LAND',
  `action_param1` DECIMAL(10,4) DEFAULT NULL COMMENT '动作参数1',
  `action_param2` DECIMAL(10,4) DEFAULT NULL COMMENT '动作参数2',
  `action_param3` DECIMAL(10,4) DEFAULT NULL COMMENT '动作参数3',
  `speed` DECIMAL(8,2) DEFAULT NULL COMMENT '该段速度(m/s)',
  `hold_time` INT DEFAULT 0 COMMENT '停留时间(秒)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mission_index` (`mission_id`, `waypoint_index`),
  CONSTRAINT `fk_waypoints_mission` FOREIGN KEY (`mission_id`) REFERENCES `mission_templates` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='航点表';

-- ======================================
-- 飞行任务表
-- ======================================
CREATE TABLE IF NOT EXISTS `flight_missions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` VARCHAR(64) NOT NULL COMMENT '无人机标识',
  `mission_template_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '航线模板ID',
  `mission_name` VARCHAR(128) NOT NULL COMMENT '任务名称',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态:0-待执行,1-执行中,2-已完成,3-已取消,4-失败',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `current_waypoint` INT DEFAULT 0 COMMENT '当前航点',
  `total_waypoints` INT DEFAULT 0 COMMENT '总航点数',
  `distance_traveled` DECIMAL(10,2) DEFAULT 0.00 COMMENT '已飞行距离(米)',
  `max_altitude_reached` DECIMAL(8,2) DEFAULT 0.00 COMMENT '达到的最大高度(米)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_status` (`status`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='飞行任务表';

-- ======================================
-- 飞行轨迹表
-- ======================================
CREATE TABLE IF NOT EXISTS `flight_trajectories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `flight_mission_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '飞行任务ID',
  `uav_id` VARCHAR(64) NOT NULL COMMENT '无人机标识',
  `latitude` DECIMAL(10,7) NOT NULL COMMENT '纬度',
  `longitude` DECIMAL(10,7) NOT NULL COMMENT '经度',
  `altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '海拔',
  `heading` DECIMAL(8,4) DEFAULT NULL COMMENT '航向',
  `timestamp` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_mission_id` (`flight_mission_id`),
  KEY `idx_uav_time` (`uav_id`, `timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='飞行轨迹表';

-- ======================================
-- 电子围栏表
-- ======================================
CREATE TABLE IF NOT EXISTS `geofences` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL COMMENT '围栏名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型:1-禁飞区,2-限飞区,3-允许区',
  `shape` VARCHAR(16) NOT NULL DEFAULT 'POLYGON' COMMENT '形状:POLYGON,CIRCLE',
  `center_latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '中心点纬度(圆形)',
  `center_longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '中心点经度(圆形)',
  `radius` DECIMAL(10,2) DEFAULT NULL COMMENT '半径(米,圆形)',
  `boundary_coordinates` JSON DEFAULT NULL COMMENT '多边形边界坐标',
  `max_altitude` DECIMAL(8,2) DEFAULT 120.00 COMMENT '最大高度(米)',
  `min_altitude` DECIMAL(8,2) DEFAULT 0.00 COMMENT '最低高度(米)',
  `is_active` TINYINT DEFAULT 1 COMMENT '是否启用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='电子围栏表';

-- ======================================
-- 告警事件表
-- ======================================
CREATE TABLE IF NOT EXISTS `alert_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` VARCHAR(64) DEFAULT NULL COMMENT '无人机标识',
  `alert_type` VARCHAR(32) NOT NULL COMMENT '告警类型:LOW_BATTERY,RTH,NO_SIGNAL,GEOFENCE,OBSTACLE,WEATHER,ERROR',
  `severity` TINYINT NOT NULL DEFAULT 1 COMMENT '严重程度:1-提示,2-警告,3-严重,4-紧急',
  `title` VARCHAR(128) NOT NULL COMMENT '告警标题',
  `message` VARCHAR(512) DEFAULT NULL COMMENT '详细信息',
  `latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '发生位置纬度',
  `longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '发生位置经度',
  `is_acknowledged` TINYINT DEFAULT 0 COMMENT '是否已确认',
  `acknowledged_at` DATETIME DEFAULT NULL COMMENT '确认时间',
  `acknowledged_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '确认人',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_alert_type` (`alert_type`),
  KEY `idx_severity` (`severity`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警事件表';

-- ======================================
-- 用户表
-- ======================================
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `real_name` VARCHAR(64) DEFAULT NULL COMMENT '真实姓名',
  `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
  `role` VARCHAR(32) DEFAULT 'USER' COMMENT '角色:ADMIN,OPERATOR,USER',
  `is_active` TINYINT DEFAULT 1 COMMENT '是否启用',
  `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ======================================
-- 固件表
-- ======================================
CREATE TABLE IF NOT EXISTS `firmware_versions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `version` VARCHAR(32) NOT NULL COMMENT '版本号',
  `hardware_model` VARCHAR(64) NOT NULL COMMENT '硬件型号',
  `file_path` VARCHAR(512) NOT NULL COMMENT '文件路径',
  `file_size` BIGINT UNSIGNED DEFAULT NULL COMMENT '文件大小(字节)',
  `md5_hash` VARCHAR(64) DEFAULT NULL COMMENT 'MD5校验',
  `release_notes` TEXT COMMENT '更新说明',
  `is_active` TINYINT DEFAULT 1 COMMENT '是否启用',
  `created_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '上传人',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_version_hardware` (`version`, `hardware_model`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='固件表';

-- ======================================
-- 黑匣子日志表
-- ======================================
CREATE TABLE IF NOT EXISTS `blackbox_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` VARCHAR(64) NOT NULL COMMENT '无人机标识',
  `flight_mission_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '飞行任务ID',
  `log_file_path` VARCHAR(512) NOT NULL COMMENT '日志文件路径',
  `log_file_size` BIGINT UNSIGNED DEFAULT NULL COMMENT '文件大小',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `duration` INT DEFAULT NULL COMMENT '飞行时长(秒)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_mission_id` (`flight_mission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='黑匣子日志表';

-- ======================================
-- 初始化数据
-- ======================================
INSERT INTO `users` (`username`, `password_hash`, `real_name`, `role`, `phone`, `email`) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '系统管理员', 'ADMIN', '13800138000', 'admin@uav.com')
ON DUPLICATE KEY UPDATE `username` = `username`;

INSERT INTO `geofences` (`name`, `type`, `shape`, `center_latitude`, `center_longitude`, `radius`, `max_altitude`, `is_active`) VALUES
('机场周边禁飞区', 1, 'CIRCLE', 39.8651, 116.3074, 5000, 120.00, 1),
('天安门禁飞区', 1, 'CIRCLE', 39.9087, 116.3975, 3000, 120.00, 1)
ON DUPLICATE KEY UPDATE `name` = `name`;

SET FOREIGN_KEY_CHECKS = 1;

-- ======================================
-- 编队表
-- ======================================
CREATE TABLE IF NOT EXISTS `formations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uuid` VARCHAR(36) NOT NULL COMMENT '编队唯一标识',
  `name` VARCHAR(100) NOT NULL COMMENT '编队名称',
  `type` VARCHAR(20) NOT NULL DEFAULT 'line' COMMENT '队形类型:line,triangle,circle',
  `status` VARCHAR(20) NOT NULL DEFAULT 'idle' COMMENT '状态:idle,ready,executing,paused,completed',
  `leader_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '长机UAV ID',
  `spacing` DECIMAL(8,2) DEFAULT 5.00 COMMENT '编队间距(米)',
  `description` TEXT COMMENT '描述',
  `owner_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_status` (`status`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='编队表';

-- ======================================
-- 编队成员表
-- ======================================
CREATE TABLE IF NOT EXISTS `formation_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `formation_id` BIGINT UNSIGNED NOT NULL COMMENT '编队ID',
  `uav_id` BIGINT UNSIGNED NOT NULL COMMENT '无人机ID',
  `position_index` INT DEFAULT 0 COMMENT '位置序号',
  `offset_x` DECIMAL(10,3) DEFAULT 0.000 COMMENT 'X偏移(米)',
  `offset_y` DECIMAL(10,3) DEFAULT 0.000 COMMENT 'Y偏移(米)',
  `offset_z` DECIMAL(10,3) DEFAULT 0.000 COMMENT 'Z偏移(米)',
  `is_leader` TINYINT DEFAULT 0 COMMENT '是否长机',
  `status` VARCHAR(20) DEFAULT 'idle' COMMENT '成员状态',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_formation_id` (`formation_id`),
  KEY `idx_uav_id` (`uav_id`),
  CONSTRAINT `fk_fm_formation` FOREIGN KEY (`formation_id`) REFERENCES `formations` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='编队成员表';

-- ======================================
-- 编队灯光配置表
-- ======================================
CREATE TABLE IF NOT EXISTS `formation_light_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `red` TINYINT UNSIGNED DEFAULT 0 COMMENT '红色通道',
  `green` TINYINT UNSIGNED DEFAULT 0 COMMENT '绿色通道',
  `blue` TINYINT UNSIGNED DEFAULT 0 COMMENT '蓝色通道',
  `effect` VARCHAR(20) DEFAULT 'static' COMMENT '灯效:static,blink,rainbow,breathing',
  `owner_id` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='编队灯光配置表';

-- ======================================
-- 编队碰撞预警表
-- ======================================
CREATE TABLE IF NOT EXISTS `formation_collision_warnings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `formation_id` BIGINT UNSIGNED NOT NULL COMMENT '编队ID',
  `uav_id_1` BIGINT UNSIGNED NOT NULL COMMENT '无人机1 ID',
  `uav_id_2` BIGINT UNSIGNED NOT NULL COMMENT '无人机2 ID',
  `distance` DECIMAL(10,3) DEFAULT NULL COMMENT '实时距离(米)',
  `warning_level` VARCHAR(20) DEFAULT 'warning' COMMENT '预警级别:warning,critical',
  `timestamp` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发生时间',
  `resolved` TINYINT DEFAULT 0 COMMENT '是否已解决',
  `resolved_at` DATETIME DEFAULT NULL COMMENT '解决时间',
  PRIMARY KEY (`id`),
  KEY `idx_formation_id` (`formation_id`),
  KEY `idx_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='编队碰撞预警表';

-- ======================================
-- AI目标检测结果表
-- ======================================
CREATE TABLE IF NOT EXISTS `detection_targets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` BIGINT UNSIGNED NOT NULL COMMENT '无人机ID',
  `class` VARCHAR(32) NOT NULL COMMENT '目标类别:person,car,truck,bus,motorcycle,bicycle,dog,unknown',
  `class_name` VARCHAR(64) DEFAULT NULL COMMENT '类别中文名',
  `confidence` DECIMAL(5,4) DEFAULT NULL COMMENT '识别置信度',
  `bbox_x` DECIMAL(8,4) DEFAULT NULL COMMENT '边界框左上角X(像素)',
  `bbox_y` DECIMAL(8,4) DEFAULT NULL COMMENT '边界框左上角Y(像素)',
  `bbox_width` DECIMAL(8,4) DEFAULT NULL COMMENT '边界框宽度(像素)',
  `bbox_height` DECIMAL(8,4) DEFAULT NULL COMMENT '边界框高度(像素)',
  `frame_width` INT DEFAULT NULL COMMENT '画面宽度(像素)',
  `frame_height` INT DEFAULT NULL COMMENT '画面高度(像素)',
  `latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '检测时无人机纬度',
  `longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '检测时无人机经度',
  `altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '检测时无人机高度(米)',
  `image_path` VARCHAR(512) DEFAULT NULL COMMENT '检测截图路径',
  `track_id` VARCHAR(64) DEFAULT NULL COMMENT '追踪ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_class` (`class`),
  KEY `idx_track_id` (`track_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI目标检测结果表';

-- ======================================
-- 目标追踪任务表
-- ======================================
CREATE TABLE IF NOT EXISTS `tracking_tasks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` BIGINT UNSIGNED NOT NULL COMMENT '无人机ID',
  `name` VARCHAR(128) DEFAULT NULL COMMENT '追踪任务名称',
  `target_class` VARCHAR(32) DEFAULT NULL COMMENT '目标类别',
  `status` VARCHAR(20) NOT NULL DEFAULT 'idle' COMMENT '状态:idle,locking,tracking,searching,lost,completed',
  `initial_bbox_x` DECIMAL(8,4) DEFAULT NULL COMMENT '初始框左上角X(像素)',
  `initial_bbox_y` DECIMAL(8,4) DEFAULT NULL COMMENT '初始框左上角Y(像素)',
  `initial_bbox_width` DECIMAL(8,4) DEFAULT NULL COMMENT '初始框宽度(像素)',
  `initial_bbox_height` DECIMAL(8,4) DEFAULT NULL COMMENT '初始框高度(像素)',
  `current_bbox_x` DECIMAL(8,4) DEFAULT NULL COMMENT '当前框左上角X(像素)',
  `current_bbox_y` DECIMAL(8,4) DEFAULT NULL COMMENT '当前框左上角Y(像素)',
  `current_bbox_width` DECIMAL(8,4) DEFAULT NULL COMMENT '当前框宽度(像素)',
  `current_bbox_height` DECIMAL(8,4) DEFAULT NULL COMMENT '当前框高度(像素)',
  `center_offset_x` DECIMAL(8,4) DEFAULT NULL COMMENT '目标中心相对画面中心X偏移(归一化)',
  `center_offset_y` DECIMAL(8,4) DEFAULT NULL COMMENT '目标中心相对画面中心Y偏移(归一化)',
  `search_radius` DECIMAL(8,2) NOT NULL DEFAULT 10.0 COMMENT '当前搜索半径(米)',
  `max_search_radius` DECIMAL(8,2) NOT NULL DEFAULT 50.0 COMMENT '最大搜索半径(米)',
  `confidence` DECIMAL(5,4) DEFAULT NULL COMMENT '当前识别置信度',
  `frames_visible` INT NOT NULL DEFAULT 0 COMMENT '目标连续可见帧数',
  `frames_lost` INT NOT NULL DEFAULT 0 COMMENT '目标连续丢失帧数',
  `target_latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '目标估算纬度',
  `target_longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '目标估算经度',
  `start_time` DATETIME DEFAULT NULL COMMENT '追踪开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '追踪结束时间',
  `created_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='目标追踪任务表';

-- ======================================
-- 降落点表
-- ======================================
CREATE TABLE IF NOT EXISTS `landing_points` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL COMMENT '降落点名称',
  `type` VARCHAR(20) NOT NULL DEFAULT 'alternate' COMMENT '类型:primary-主降落点,alternate-备用降落点,emergency-紧急降落点',
  `latitude` DECIMAL(10,7) NOT NULL COMMENT '纬度',
  `longitude` DECIMAL(10,7) NOT NULL COMMENT '经度',
  `altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '海拔高度(米)',
  `has_markers` TINYINT DEFAULT 0 COMMENT '是否有降落标志',
  `marker_type` VARCHAR(20) DEFAULT NULL COMMENT '标志类型:QR_CODE,H_MARKER,APRIL_TAG,ARUCO',
  `is_moving_platform` TINYINT DEFAULT 0 COMMENT '是否为移动平台',
  `platform_identifier` VARCHAR(64) DEFAULT NULL COMMENT '移动平台标识符',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '描述',
  `is_active` TINYINT DEFAULT 1 COMMENT '是否启用',
  `priority` INT DEFAULT 1 COMMENT '优先级(数字越小优先级越高)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_active` (`is_active`),
  KEY `idx_moving_platform` (`is_moving_platform`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='降落点表';

-- ======================================
-- 降落会话表
-- ======================================
CREATE TABLE IF NOT EXISTS `landing_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` BIGINT UNSIGNED NOT NULL COMMENT '无人机ID',
  `landing_point_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '目标降落点ID',
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态:pending-待降落,approaching-接近阶段,descending-下降阶段,precision-精准阶段,touchdown-触地阶段,landed-已降落,aborted-已中止,failed-失败',
  `rtk_enabled` TINYINT DEFAULT 1 COMMENT '是否启用RTK',
  `vision_enabled` TINYINT DEFAULT 1 COMMENT '是否启用视觉降落',
  `is_moving_platform` TINYINT DEFAULT 0 COMMENT '是否为移动平台降落',
  `platform_velocity_x` DECIMAL(8,4) DEFAULT NULL COMMENT '平台X方向速度(m/s)',
  `platform_velocity_y` DECIMAL(8,4) DEFAULT NULL COMMENT '平台Y方向速度(m/s)',
  `landing_error` DECIMAL(6,3) DEFAULT NULL COMMENT '着陆误差(米)',
  `approach_altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '接近高度(米)',
  `final_altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '最终高度(米)',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `abort_reason` VARCHAR(255) DEFAULT NULL COMMENT '中止原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_status` (`status`),
  KEY `idx_landing_point` (`landing_point_id`),
  KEY `idx_uav_status` (`uav_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='降落会话表';

-- ======================================
-- 降落轨迹点表
-- ======================================
CREATE TABLE IF NOT EXISTS `landing_trajectory_points` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL COMMENT '降落会话ID',
  `sequence` INT NOT NULL DEFAULT 0 COMMENT '序列编号',
  `phase` VARCHAR(20) NOT NULL COMMENT '降落阶段:approach,descend,precision,touchdown',
  `latitude` DECIMAL(10,7) NOT NULL COMMENT '纬度',
  `longitude` DECIMAL(10,7) NOT NULL COMMENT '经度',
  `altitude` DECIMAL(8,2) NOT NULL COMMENT '高度(米)',
  `velocity_x` DECIMAL(8,4) DEFAULT NULL COMMENT 'X方向速度(m/s)',
  `velocity_y` DECIMAL(8,4) DEFAULT NULL COMMENT 'Y方向速度(m/s)',
  `velocity_z` DECIMAL(8,4) DEFAULT NULL COMMENT 'Z方向速度(m/s)',
  `heading` DECIMAL(8,4) DEFAULT NULL COMMENT '航向角(度)',
  `target_latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '目标纬度',
  `target_longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '目标经度',
  `target_altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '目标高度(米)',
  `horizontal_error` DECIMAL(6,3) DEFAULT NULL COMMENT '水平误差(米)',
  `vertical_error` DECIMAL(6,3) DEFAULT NULL COMMENT '垂直误差(米)',
  `timestamp` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_session_sequence` (`session_id`, `sequence`),
  KEY `idx_session_phase` (`session_id`, `phase`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='降落轨迹点表';

-- ======================================
-- 强制降落事件表
-- ======================================
CREATE TABLE IF NOT EXISTS `forced_landing_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uav_id` BIGINT UNSIGNED NOT NULL COMMENT '无人机ID',
  `session_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联的降落会话ID',
  `trigger_type` VARCHAR(32) NOT NULL COMMENT '触发类型:EMERGENCY_SWITCH,LOW_BATTERY,COMM_LOSS,MOTOR_FAILURE,SYSTEM_FAILURE,USER_FORCED',
  `trigger_source` VARCHAR(32) NOT NULL COMMENT '触发来源:GROUND_STATION,ONBOARD,AUTOPILOT',
  `triggered_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '触发人ID(地面站触发时)',
  `status` VARCHAR(20) NOT NULL DEFAULT 'triggered' COMMENT '状态:triggered-已触发,executing-执行中,completed-已完成,resolved-已解除',
  `arm_locked` TINYINT DEFAULT 0 COMMENT '机臂是否锁定',
  `latitude` DECIMAL(10,7) DEFAULT NULL COMMENT '触发时纬度',
  `longitude` DECIMAL(10,7) DEFAULT NULL COMMENT '触发时经度',
  `altitude` DECIMAL(8,2) DEFAULT NULL COMMENT '触发时高度(米)',
  `reason` VARCHAR(255) DEFAULT NULL COMMENT '触发原因',
  `triggered_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '触发时间',
  `resolved_at` DATETIME DEFAULT NULL COMMENT '解除时间',
  `resolved_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '解除人ID',
  `resolve_note` VARCHAR(255) DEFAULT NULL COMMENT '解除备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uav_id` (`uav_id`),
  KEY `idx_status` (`status`),
  KEY `idx_trigger_type` (`trigger_type`),
  KEY `idx_uav_status` (`uav_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='强制降落事件表';

-- ======================================
-- 视觉降落数据表
-- ======================================
CREATE TABLE IF NOT EXISTS `vision_landing_data` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL COMMENT '降落会话ID',
  `marker_detected` TINYINT NOT NULL DEFAULT 0 COMMENT '是否检测到标志',
  `marker_type` VARCHAR(20) DEFAULT NULL COMMENT '标志类型',
  `marker_id` VARCHAR(64) DEFAULT NULL COMMENT '标志ID',
  `confidence` DECIMAL(5,4) DEFAULT NULL COMMENT '识别置信度',
  `offset_x` DECIMAL(8,4) DEFAULT NULL COMMENT 'X方向偏移(米)',
  `offset_y` DECIMAL(8,4) DEFAULT NULL COMMENT 'Y方向偏移(米)',
  `offset_z` DECIMAL(8,4) DEFAULT NULL COMMENT 'Z方向偏移(米)',
  `yaw_offset` DECIMAL(8,4) DEFAULT NULL COMMENT '偏航角偏移(度)',
  `image_width` INT DEFAULT NULL COMMENT '画面宽度(像素)',
  `image_height` INT DEFAULT NULL COMMENT '画面高度(像素)',
  `marker_corners` JSON DEFAULT NULL COMMENT '标志角点坐标',
  `timestamp` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_marker_detected` (`marker_detected`),
  KEY `idx_session_timestamp` (`session_id`, `timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视觉降落数据表';

-- ======================================
-- RTK位置数据表
-- ======================================
CREATE TABLE IF NOT EXISTS `rtk_position_data` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `session_id` BIGINT UNSIGNED NOT NULL COMMENT '降落会话ID',
  `fix_type` TINYINT NOT NULL DEFAULT 0 COMMENT '定位类型:0-NONE,1-SINGLE,2-DGPS,3-PPP,4-RTK_FIXED,5-RTK_FLOAT',
  `satellites_visible` INT DEFAULT NULL COMMENT '可见卫星数',
  `satellites_used` INT DEFAULT NULL COMMENT '使用卫星数',
  `latitude` DECIMAL(12,9) NOT NULL COMMENT 'RTK纬度',
  `longitude` DECIMAL(12,9) NOT NULL COMMENT 'RTK经度',
  `altitude` DECIMAL(10,3) NOT NULL COMMENT 'RTK高度(米)',
  `latitude_stdev` DECIMAL(8,4) DEFAULT NULL COMMENT '纬度标准差(米)',
  `longitude_stdev` DECIMAL(8,4) DEFAULT NULL COMMENT '经度标准差(米)',
  `altitude_stdev` DECIMAL(8,4) DEFAULT NULL COMMENT '高度标准差(米)',
  `baseline_length` DECIMAL(10,3) DEFAULT NULL COMMENT '基线长度(米)',
  `age_of_correction` DECIMAL(8,3) DEFAULT NULL COMMENT '校正数据延迟(秒)',
  `timestamp` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_fix_type` (`fix_type`),
  KEY `idx_session_timestamp` (`session_id`, `timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='RTK位置数据表';
