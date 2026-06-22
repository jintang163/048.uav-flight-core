package nsq

const (
	TopicTelemetryData     = "uav_telemetry"
	TopicUAVHeartbeat      = "uav_heartbeat"
	TopicAlertEvent        = "alert_event"
	TopicMAVLinkCommand    = "mavlink_command"
	TopicMAVLinkMessage    = "mavlink_message"
	TopicFlightStatus      = "flight_status"
	TopicMissionUpdate     = "mission_update"
	TopicGeofenceViolation = "geofence_violation"
	TopicLinkStatus        = "link_status"
	TopicBlackboxUpload    = "blackbox_upload"
	TopicOTAUpdate         = "ota_update"
	TopicWeatherSensor     = "weather_sensor"
	TopicPositionUpdate    = "position_update"

	ChannelTelemetryProcessor  = "telemetry_processor"
	ChannelAlertNotifier       = "alert_notifier"
	ChannelMAVLinkHandler      = "mavlink_handler"
	ChannelDataPersister       = "data_persister"
	ChannelNotifierSMS         = "notifier_sms"
	ChannelNotifierEmail       = "notifier_email"
	ChannelLinkHealthHandler   = "link_health_handler"
	ChannelWeatherSensorHandler = "weather_sensor_handler"
	ChannelCollisionAvoidance  = "collision_avoidance"
)

type NSQConfig struct {
	NSQDTCPAddress   string
	NSQDHTTPAddress  string
	LookupdTCPAddrs  []string
	LookupdHTTPAddrs []string
}
