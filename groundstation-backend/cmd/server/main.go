package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"groundstation-backend/internal/config"
	"groundstation-backend/internal/handler"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/nsq"
	"groundstation-backend/internal/service"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/tts"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	viper.SetDefault("webrtc.media_server_host", "127.0.0.1")
	viper.SetDefault("webrtc.media_server_port", 8888)
	viper.SetDefault("webrtc.whip_endpoint", "/whip")
	viper.SetDefault("webrtc.whep_endpoint", "/whep")

	if _, err := config.LoadConfig(configPath); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := config.InitDB(); err != nil {
		fmt.Printf("Failed to init database: %v\n", err)
		os.Exit(1)
	}
	if err := config.InitRedis(); err != nil {
		fmt.Printf("Failed to init redis: %v\n", err)
		os.Exit(1)
	}

	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.UAV{},
		&models.FlightStatus{},
		&models.MissionTemplate{},
		&models.MissionWaypoint{},
		&models.FlightMission{},
		&models.Geofence{},
		&models.GeofenceCoordinate{},
		&models.GeofenceViolationLog{},
		&models.TemporaryUnlocking{},
		&models.AlertEvent{},
		&models.Firmware{},
		&models.FirmwareUpdate{},
		&models.BlackboxLog{},
		&models.Formation{},
		&models.FormationMember{},
		&models.FormationLightConfig{},
		&models.FormationCollisionWarning{},
		&models.DetectionTarget{},
		&models.TrackingTask{},
		&models.PayloadDevice{},
		&models.CameraStatus{},
		&models.SpeakerAudio{},
		&models.SprayerStatus{},
		&models.PayloadTelemetry{},
		&models.OrbitMission{},
		&models.OrthoMission{},
		&models.OrthoWaypoint{},
		&models.TextToSpeechTask{},
		&models.MotorStatus{},
		&models.MotorFailureEvent{},
		&models.LinkStatus{},
		&models.Battery{},
		&models.BatteryUsageRecord{},
		&models.BatteryCellData{},
		&models.ChargingStation{},
		&models.ChargingSlot{},
		&models.ChargingRecord{},
		&models.BatteryMaintenanceAlert{},
		&models.ObstacleAvoidanceConfig{},
		&models.ObstacleDetectionLog{},
		&models.ObstacleAvoidanceEvent{},
		&models.ObstacleHeatmapPoint{},
		&models.ThrustLearningStatus{},
		&models.ThrustCurvePoint{},
		&models.PIDGainProfile{},
		&models.ThrustLearningSample{},
		&models.CockpitSession{},
		&models.VideoStreamSession{},
		&models.CockpitLinkSnapshot{},
		&models.NetworkMetricsLog{},
		&models.FlightControlLog{},
		&models.WeatherData{},
		&models.FlightWeatherLog{},
		&models.WeatherAlertEvent{},
		&models.CollisionAlert{},
		&models.RouteIntersection{},
		&models.LandingPoint{},
		&models.LandingSession{},
		&models.LandingTrajectoryPoint{},
		&models.ForcedLandingEvent{},
		&models.VisionLandingData{},
		&models.RTKPositionData{},
	); err != nil {
		fmt.Printf("Failed to migrate database: %v\n", err)
		os.Exit(1)
	}

	initDefaultUser()

	if err := nsq.InitProducer(); err != nil {
		fmt.Printf("Warning: Failed to init NSQ producer: %v\n", err)
	}

	nsqConsumerMgr := nsq.NewConsumerManager()
	if err := nsqConsumerMgr.StartConsumers(); err != nil {
		fmt.Printf("Warning: Failed to start NSQ consumers: %v\n", err)
	}

	mavlinkMgr := mavlink.NewCommandManager()
	if err := mavlinkMgr.Start(); err != nil {
		fmt.Printf("Warning: Failed to start MAVLink server: %v\n", err)
	}

	metricsService := service.NewMetricsService()
	metricsService.StartCollector(30 * time.Second)

	batteryService := service.NewBatteryService()
	batteryService.StartMaintenanceScheduler(1*time.Hour, 7)

	ttsConfig := tts.TTSServiceConfig{
		AudioDir:     "./data/tts_audio",
		AudioURLBase: "/api/v1/static/tts",
	}
	if len(config.AppConfig.ExternalServices) > 0 {
		if url, ok := config.AppConfig.ExternalServices["edge_tts_url"]; ok {
			ttsConfig.EdgeTTSURL = url
		}
	}
	_ = tts.NewTTSService(ttsConfig)

	gin.SetMode(gin.ReleaseMode)
	if config.AppConfig.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RateLimitMiddleware(100, time.Minute))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/refresh", handler.RefreshToken)
			auth.POST("/logout", middleware.JWTAuth(), handler.Logout)
			auth.GET("/me", middleware.JWTAuth(), handler.GetCurrentUser)
			auth.PUT("/password", middleware.JWTAuth(), handler.ChangePassword)
		}

		admin := api.Group("", middleware.JWTAuth(), middleware.RoleAuth(models.UserRoleAdmin))
		{
			admin.GET("/users", handler.ListUsers)
			admin.GET("/users/:id", handler.GetUser)
			admin.PUT("/users/:id", handler.UpdateUser)
			admin.DELETE("/users/:id", handler.DeleteUser)
		}

		uav := api.Group("/uavs", middleware.JWTAuth())
		{
			uav.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateUAV)
			uav.GET("", handler.ListUAVs)
			uav.GET("/statistics", handler.GetUAVStatistics)
			uav.GET("/:id", handler.GetUAV)
			uav.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateUAV)
			uav.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteUAV)
			uav.GET("/:id/status", handler.GetUAVStatus)
			uav.GET("/:id/history", handler.GetUAVFlightHistory)
		}

		flight := api.Group("/flights", middleware.JWTAuth())
		{
			flight.GET("/uav/:uav_id/latest", handler.GetLatestFlightStatus)
			flight.GET("/uav/:uav_id/history", handler.GetFlightHistory)
			flight.GET("/uav/:uav_id/realtime", handler.GetRealTimeTelemetry)
			flight.GET("/realtime", handler.GetAllUAVsRealtime)
			flight.GET("/uav/:uav_id/statistics", handler.GetFlightStatistics)
			flight.GET("/uav/:uav_id/track", handler.GetFlightTrack)
		}

		mission := api.Group("/missions", middleware.JWTAuth())
		{
			templates := mission.Group("/templates")
			{
				templates.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateTemplate)
				templates.GET("", handler.ListTemplates)
				templates.GET("/:id", handler.GetTemplate)
				templates.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateTemplate)
				templates.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteTemplate)
			}

			mission.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateMission)
			mission.GET("", handler.ListMissions)
			mission.GET("/:id", handler.GetMission)
			mission.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateMission)
			mission.POST("/:id/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartMission)
			mission.POST("/:id/pause", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PauseMission)
			mission.POST("/:id/resume", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResumeMission)
			mission.POST("/:id/resume-breakpoint", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResumeMissionFromBreakpoint)
			mission.POST("/:id/abort", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AbortMission)
			mission.POST("/:id/waypoint/current", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCurrentWaypoint)
			mission.POST("/:id/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AbortMission)
		}

		geofence := api.Group("/geofences", middleware.JWTAuth())
		{
			geofence.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateGeofence)
			geofence.GET("", handler.ListGeofences)
			geofence.GET("/:id", handler.GetGeofence)
			geofence.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateGeofence)
			geofence.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteGeofence)
			geofence.GET("/uav/:uav_id", handler.GetUAVGeofences)
			geofence.GET("/uav/:uav_id/check", handler.CheckViolation)
			geofence.GET("/uav/:uav_id/takeoff-check", handler.CheckTakeoffPermission)
			geofence.POST("/national/import", middleware.RoleAuth(models.UserRoleAdmin), handler.ImportNationalGeofences)
		}

		violation := api.Group("/geofence-violations", middleware.JWTAuth())
		{
			violation.GET("", handler.ListViolations)
			violation.GET("/statistics", handler.GetViolationStatistics)
			violation.GET("/:id", handler.GetViolation)
			violation.POST("/:id/resolve", handler.ResolveViolation)
			violation.POST("/batch/resolve", handler.BatchResolveViolations)
		}

		unlocking := api.Group("/temporary-unlockings", middleware.JWTAuth())
		{
			unlocking.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ApplyUnlocking)
			unlocking.GET("", handler.ListUnlockings)
			unlocking.GET("/:id", handler.GetUnlocking)
			unlocking.POST("/:id/approve", middleware.RoleAuth(models.UserRoleAdmin), handler.ApproveUnlocking)
			unlocking.POST("/:id/reject", middleware.RoleAuth(models.UserRoleAdmin), handler.RejectUnlocking)
			unlocking.POST("/:id/cancel", handler.CancelUnlocking)
			unlocking.GET("/uav/:uav_id/active", handler.GetActiveUnlockings)
		}

		alert := api.Group("/alerts", middleware.JWTAuth())
		{
			alert.GET("", handler.ListAlerts)
			alert.GET("/statistics", handler.GetAlertStatistics)
			alert.GET("/:id", handler.GetAlert)
			alert.POST("/:id/acknowledge", handler.AcknowledgeAlert)
			alert.POST("/:id/resolve", handler.ResolveAlert)
			alert.POST("/batch/acknowledge", handler.BatchAcknowledge)
			alert.POST("/test-notification", middleware.RoleAuth(models.UserRoleAdmin), handler.SendTestNotification)
		}

		ota := api.Group("/ota", middleware.JWTAuth())
		{
			ota.POST("/firmware", middleware.RoleAuth(models.UserRoleAdmin), handler.UploadFirmware)
			ota.GET("/firmware", handler.ListFirmwares)
			ota.GET("/firmware/latest", handler.GetLatestFirmware)
			ota.GET("/firmware/:id", handler.GetFirmware)
			ota.PUT("/firmware/:id/status", middleware.RoleAuth(models.UserRoleAdmin), handler.UpdateFirmwareStatus)
			ota.DELETE("/firmware/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteFirmware)
			ota.GET("/firmware/:id/download", handler.DownloadFirmware)

			ota.POST("/update", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartUpdate)
			ota.GET("/update", handler.ListUpdates)
			ota.GET("/update/:id", handler.GetUpdate)
			ota.GET("/update/uav/:uav_id/active", handler.GetActiveUpdate)
		}

		blackbox := api.Group("/blackbox", middleware.JWTAuth())
		{
			blackbox.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UploadBlackbox)
			blackbox.POST("/auto-upload", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AutoUploadBlackbox)
			blackbox.GET("", handler.ListBlackboxes)
			blackbox.GET("/statistics", handler.GetBlackboxStatistics)
			blackbox.GET("/:id", handler.GetBlackbox)
			blackbox.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateBlackbox)
			blackbox.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteBlackbox)
			blackbox.GET("/:id/download", handler.DownloadBlackbox)
			blackbox.GET("/:id/parse", handler.ParseBlackboxLog)
			blackbox.GET("/:id/analysis", handler.GetAnalysisReport)
			blackbox.POST("/:id/analyze", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AnalyzeBlackbox)
			blackbox.GET("/:id/export/csv", handler.ExportBlackboxCSV)
			blackbox.GET("/:id/export/report", handler.ExportBlackboxReport)
			blackbox.GET("/:id/reports", handler.GetBlackboxReports)
		}

		formation := api.Group("/formations", middleware.JWTAuth())
		{
			formation.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateFormation)
			formation.GET("", handler.ListFormations)
			formation.GET("/active", handler.GetActiveFormations)
			formation.GET("/:id", handler.GetFormation)
			formation.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateFormation)
			formation.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteFormation)
			formation.GET("/:id/members", handler.GetFormationMembers)
			formation.POST("/:id/members", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AddFormationMember)
			formation.DELETE("/:id/members/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.RemoveFormationMember)
			formation.POST("/:id/leader/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetFormationLeader)
			formation.POST("/:id/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartFormation)
			formation.POST("/:id/pause", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PauseFormation)
			formation.POST("/:id/resume", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResumeFormation)
			formation.POST("/:id/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopFormation)
			formation.GET("/:id/collisions", handler.GetCollisionWarnings)
			formation.POST("/:id/light", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetFormationLight)
			formation.POST("/:id/sync-waypoints", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SyncFormationWaypoints)
			formation.POST("/:id/takeoff", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.MultiTakeoff)
		}

		preflight := api.Group("/preflight", middleware.JWTAuth())
		{
			preflight.POST("/run", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.RunPreflightCheck)
			preflight.POST("/batch", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.BatchRunPreflightCheck)
			preflight.GET("/thresholds", handler.GetPreflightThresholds)
		}

		tracking := api.Group("/tracking", middleware.JWTAuth())
		{
			tracking.POST("/lock", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.LockTarget)
			tracking.POST("/:id/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopTracking)
			tracking.GET("", handler.ListTrackingTasks)
			tracking.GET("/:id", handler.GetTrackingTask)
			tracking.GET("/uav/:uav_id/active", handler.GetActiveTracking)
			tracking.GET("/uav/:uav_id/detections", handler.ListDetections)
			tracking.POST("/uav/:uav_id/detect", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.DetectImage)
		}

		payloads := api.Group("/payloads", middleware.JWTAuth())
		{
			payloads.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreatePayload)
			payloads.GET("", handler.ListPayloads)
			payloads.GET("/uav/:uav_id", handler.ListUAVPayloads)
			payloads.GET("/:id", handler.GetPayload)
			payloads.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdatePayload)
			payloads.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeletePayload)

			payloads.GET("/:id/camera/status", handler.GetCameraStatus)
			payloads.POST("/:id/camera/photo", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.TakePhoto)
			payloads.POST("/:id/camera/recording/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartVideoRecording)
			payloads.POST("/:id/camera/recording/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopVideoRecording)
			payloads.POST("/:id/camera/mode", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCameraMode)
			payloads.POST("/:id/camera/zoom", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCameraZoom)
			payloads.POST("/:id/camera/settings", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCameraSettings)

			payloads.GET("/:id/sprayer/status", handler.GetSprayerStatus)
			payloads.POST("/:id/sprayer/flow", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetSprayerFlowRate)
			payloads.POST("/:id/sprayer/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartSpraying)
			payloads.POST("/:id/sprayer/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopSpraying)

			payloads.POST("/speaker/audios", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateSpeakerAudio)
			payloads.GET("/speaker/audios", handler.ListSpeakerAudios)
			payloads.GET("/speaker/audios/:id", handler.GetSpeakerAudio)
			payloads.DELETE("/speaker/audios/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteSpeakerAudio)
			payloads.POST("/:id/speaker/play/:audio_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PlaySpeakerAudio)
			payloads.POST("/:id/speaker/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopSpeaker)
		}

		payloadMissions := api.Group("/payload-missions", middleware.JWTAuth())
		{
			orbit := payloadMissions.Group("/orbit")
			{
				orbit.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateOrbitMission)
				orbit.GET("", handler.ListOrbitMissions)
				orbit.GET("/:id", handler.GetOrbitMission)
				orbit.POST("/:id/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartOrbitMission)
				orbit.POST("/:id/pause", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PauseOrbitMission)
				orbit.POST("/:id/resume", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResumeOrbitMission)
				orbit.POST("/:id/abort", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AbortOrbitMission)
				orbit.POST("/:id/progress", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateOrbitProgress)
			}

			ortho := payloadMissions.Group("/ortho")
			{
				ortho.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateOrthoMission)
				ortho.GET("", handler.ListOrthoMissions)
				ortho.GET("/:id", handler.GetOrthoMission)
				ortho.POST("/:id/plan", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PlanOrthoMission)
				ortho.POST("/:id/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartOrthoMission)
				ortho.POST("/:id/pause", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PauseOrthoMission)
				ortho.POST("/:id/resume", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResumeOrthoMission)
				ortho.POST("/:id/abort", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AbortOrthoMission)
			}

			tts := payloadMissions.Group("/tts")
			{
				tts.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateTTS)
				tts.GET("", handler.ListTTSTasks)
				tts.GET("/:id", handler.GetTTSTask)
			}
		}

		metrics := api.Group("/metrics")
		{
			metrics.GET("", gin.WrapH(promhttp.Handler()))
			metrics.GET("/summary", middleware.JWTAuth(), func(c *gin.Context) {
				summary, err := metricsService.GetSummary()
				if err != nil {
					utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
					return
				}
				utils.SuccessResponse(c, "获取成功", summary)
			})
		}

		motor := api.Group("/motor-protection", middleware.JWTAuth())
		{
			motor.GET("/uav/:uav_id/status", handler.GetMotorStatuses)
			motor.GET("/uav/:uav_id/failure-state", handler.GetMotorFailureState)
			motor.POST("/uav/:uav_id/pid-adjust", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ManualPIDAdjustment)
			motor.POST("/uav/:uav_id/emergency-rth", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.EmergencyRTH)
			motor.POST("/uav/:uav_id/emergency-land", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.EmergencyLand)
			motor.POST("/uav/:uav_id/resolve/:motor_index", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResolveMotorFailure)
		}

		link := api.Group("/link", middleware.JWTAuth())
		{
			link.POST("/status", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ReportLinkStatus)
			link.GET("/:uav_id/latest", handler.GetLinkStatus)
			link.GET("/:uav_id/history", handler.GetLinkHistory)
			link.GET("/statistics", handler.GetLinkStatistics)
		}

		battery := api.Group("/batteries", middleware.JWTAuth())
		{
			battery.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateBattery)
			battery.GET("", handler.ListBatteries)
			battery.GET("/statistics", handler.GetBatteryStatistics)
			battery.GET("/identify", handler.IdentifyBattery)
			battery.GET("/maintenance/alerts", handler.GetMaintenanceAlerts)
			battery.GET("/maintenance/alerts/unacknowledged-count", handler.GetUnacknowledgedMaintenanceCount)
			battery.POST("/maintenance/check", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CheckMaintenanceReminders)
			battery.GET("/:id", handler.GetBattery)
			battery.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateBattery)
			battery.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteBattery)
			battery.GET("/:id/usage-records", handler.GetBatteryUsageRecords)
			battery.GET("/:id/cell-data", handler.GetBatteryCellData)
			battery.POST("/:id/telemetry", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateBatteryTelemetry)
			battery.POST("/:id/register-use", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.RegisterBatteryUse)
			battery.PUT("/:id/soh", middleware.RoleAuth(models.UserRoleAdmin), handler.UpdateBatterySOH)
			battery.POST("/maintenance/alerts/:id/acknowledge", handler.AcknowledgeMaintenanceAlert)
			battery.POST("/maintenance/alerts/:id/resolve", handler.ResolveMaintenanceAlert)
		}

		charging := api.Group("/charging", middleware.JWTAuth())
		{
			stations := charging.Group("/stations")
			{
				stations.POST("", middleware.RoleAuth(models.UserRoleAdmin), handler.CreateChargingStation)
				stations.GET("", handler.ListChargingStations)
				stations.GET("/statistics", handler.GetChargingStatistics)
				stations.GET("/:id", handler.GetChargingStation)
				stations.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.UpdateChargingStation)
				stations.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteChargingStation)
				stations.GET("/:id/slots", handler.GetChargingStationSlots)
				stations.POST("/:id/heartbeat", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StationHeartbeat)
				stations.GET("/:id/records", handler.GetStationChargingRecords)
			}

			slots := charging.Group("/slots")
			{
				slots.GET("/:slot_id", handler.GetChargingSlot)
				slots.POST("/:slot_id/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartCharging)
				slots.POST("/:slot_id/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopCharging)
				slots.POST("/:slot_id/telemetry", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateSlotTelemetry)
				slots.POST("/:slot_id/assign", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AssignBatteryToSlot)
				slots.POST("/:slot_id/remove", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.RemoveBatteryFromSlot)
				slots.POST("/:slot_id/fault", middleware.RoleAuth(models.UserRoleAdmin), handler.SetSlotFault)
			}

			records := charging.Group("/records")
			{
				records.GET("", handler.GetChargingRecords)
				records.GET("/:id", handler.GetChargingRecord)
				records.GET("/battery/:battery_id", handler.GetBatteryChargingRecords)
			}
		}

		obstacleAvoidance := api.Group("/obstacle-avoidance", middleware.JWTAuth())
		{
			obstacleAvoidance.GET("/config/:uav_id", handler.GetObstacleAvoidanceConfig)
			obstacleAvoidance.PUT("/config/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateObstacleAvoidanceConfig)
			obstacleAvoidance.GET("/heatmap", handler.GetObstacleHeatmap)
			obstacleAvoidance.DELETE("/heatmap", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ClearObstacleHeatmap)
			obstacleAvoidance.GET("/logs", handler.GetObstacleAvoidanceLogs)
			obstacleAvoidance.GET("/statistics", handler.GetObstacleAvoidanceStatistics)
			obstacleAvoidance.GET("/events", handler.GetAvoidanceEvents)
			obstacleAvoidance.GET("/events/:id", handler.GetAvoidanceEventDetail)
			obstacleAvoidance.POST("/test/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.TriggerAvoidanceTest)
		}

		thrustLearning := api.Group("/thrust-learning", middleware.JWTAuth())
		{
			thrustLearning.GET("/status/:uav_id", handler.GetThrustLearningStatus)
			thrustLearning.POST("/trigger/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.TriggerThrustLearning)
			thrustLearning.GET("/curve/:uav_id", handler.GetThrustCurve)
			thrustLearning.PUT("/curve/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateThrustCurve)
			thrustLearning.GET("/pid/:uav_id", handler.GetPIDGains)
			thrustLearning.PUT("/pid/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdatePIDGains)
			thrustLearning.POST("/pid/apply/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ApplyAutoTunedPID)
			thrustLearning.GET("/samples/:uav_id", handler.GetThrustLearningSamples)
			thrustLearning.POST("/optimize/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.OptimizeThrustModel)
		}

		cockpit := api.Group("/remote-cockpit", middleware.JWTAuth())
		{
			cockpit.POST("/sessions", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartCockpitSession)
			cockpit.DELETE("/sessions/:uavId", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.EndCockpitSession)
			cockpit.GET("/sessions/:uavId", handler.GetCockpitSession)
			cockpit.GET("/uavs", handler.GetAvailableCockpitUAVs)
			cockpit.POST("/switch", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SwitchCockpitUAV)

			cockpit.POST("/video/:uavId/start", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartCockpitVideoStream)
			cockpit.POST("/video/:uavId/stop", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StopCockpitVideoStream)
			cockpit.GET("/video/:uavId", handler.GetCockpitVideoStream)
			cockpit.POST("/video/:uavId/quality", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AdjustCockpitVideoQuality)
			cockpit.GET("/video/:uavId/url", handler.GetCockpitStreamURL)

			cockpit.GET("/link/:uavId", handler.GetCockpitLinkStatus)
			cockpit.POST("/link/:uavId/primary", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCockpitPrimaryLink)
			cockpit.POST("/link/:uavId/failover", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCockpitFailoverEnabled)
			cockpit.POST("/link/:uavId/fallback", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SetCockpitAutoMissionFallback)
			cockpit.POST("/link/:uavId/fallback/trigger", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.TriggerCockpitAutoMissionFallback)

			cockpit.POST("/control/:uavId", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SendCockpitFlightControl)

			cockpit.GET("/metrics/:uavId/network", handler.GetCockpitNetworkMetrics)
			cockpit.GET("/metrics/:uavId/control-logs", handler.GetCockpitFlightControlLogs)

			cockpit.POST("/video/:uavId/sdp", handler.HandleCockpitSDPOffer)
			cockpit.GET("/video/:uavId/webrtc-stats", handler.GetCockpitWebRTCStats)
		}

		weather := api.Group("/weather", middleware.JWTAuth())
		{
			weather.GET("/uav/:uav_id/latest", handler.GetLatestWeather)
			weather.GET("/uav/:uav_id/history", handler.GetWeatherHistory)
			weather.GET("/uav/:uav_id/alerts/active", handler.GetActiveWeatherAlerts)
			weather.GET("/uav/:uav_id/takeoff-check", handler.CheckTakeoffWeather)
			weather.POST("/sensor", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ReportWeatherSensorData)
			weather.GET("/fetch", handler.FetchWeatherFromAPI)
			weather.GET("/alerts", handler.GetWeatherAlerts)
			weather.POST("/alerts/:id/resolve", handler.ResolveWeatherAlert)
			weather.GET("/thresholds", handler.GetWeatherThresholds)
			weather.PUT("/thresholds", middleware.RoleAuth(models.UserRoleAdmin), handler.UpdateWeatherThresholds)
			weather.GET("/flight/:flight_id", handler.GetFlightWeatherLog)
		}

		collision := api.Group("/collision", middleware.JWTAuth())
		{
			collision.GET("/status", handler.GetCollisionAvoidanceStatus)
			collision.PUT("/enabled", middleware.RoleAuth(models.UserRoleAdmin), handler.ToggleCollisionAvoidance)
			collision.GET("/alerts/active", handler.GetActiveCollisionAlerts)
			collision.GET("/alerts", handler.ListCollisionAlerts)
			collision.POST("/alerts/:id/resolve", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ResolveCollisionAlert)
			collision.GET("/positions", handler.GetAllUAVPositions)
			collision.GET("/intersections", handler.GetRouteIntersections)
			collision.POST("/intersections/detect", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.DetectRouteIntersections)
			collision.POST("/manual", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.ManualCollisionAvoidance)
			collision.GET("/uav/:uav_id/speed-factor", handler.GetUAVSpeedFactor)
			collision.GET("/uav/:uav_id/cmd-status", handler.GetAvoidCommandStatus)
			collision.GET("/stats", handler.GetCollisionStats)
		}

		landing := api.Group("/landing", middleware.JWTAuth())
		{
			landingPoints := landing.Group("/points")
			{
				landingPoints.POST("", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.CreateLandingPoint)
				landingPoints.GET("", handler.ListLandingPoints)
				landingPoints.GET("/:id", handler.GetLandingPoint)
				landingPoints.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateLandingPoint)
				landingPoints.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteLandingPoint)
			}

			landing.POST("/plan/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.PlanLanding)
			landing.POST("/start/:uav_id/:session_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.StartLanding)
			landing.POST("/abort/:uav_id/:session_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.AbortLanding)
			landing.POST("/switch-alternate/:uav_id/:session_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.SwitchToAlternateLanding)
			landing.GET("/session/active/:uav_id", handler.GetActiveLandingSession)
			landing.GET("/session/:id", handler.GetLandingSession)
			landing.GET("/sessions", handler.ListLandingSessions)
			landing.GET("/trajectory/:session_id", handler.GetLandingTrajectory)
			landing.GET("/statistics", handler.GetLandingStatistics)

			landing.POST("/trajectory/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.RecordTrajectoryPoint)
			landing.POST("/vision/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateVisionLandingData)
			landing.POST("/rtk/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateRTKPositionData)
			landing.POST("/moving-platform/:uav_id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateMovingPlatformPosition)

			forcedLanding := landing.Group("/forced")
			{
				forcedLanding.POST("/:uav_id", middleware.RoleAuth(models.UserRoleAdmin), handler.TriggerForcedLanding)
				forcedLanding.GET("/active/:uav_id", handler.GetActiveForcedLandingEvent)
				forcedLanding.GET("/events", handler.ListForcedLandingEvents)
				forcedLanding.GET("/event/:id", handler.GetForcedLandingEvent)
				forcedLanding.POST("/event/:id/resolve", middleware.RoleAuth(models.UserRoleAdmin), handler.ResolveForcedLanding)
			}
		}
	}

	api.Static("/static/tts", "./data/tts_audio")

	ws := r.Group("/ws", middleware.JWTAuth())
	{
		ws.GET("", websocket.ServeWS)
	}

	api.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, "ok", gin.H{
			"status":    "running",
			"timestamp": time.Now().Unix(),
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler: r,
	}

	go func() {
		middleware.Logger.Info("Server starting", zap.Int("port", config.AppConfig.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	mavlinkMgr.Stop()
	nsqConsumerMgr.StopConsumers()
	nsq.StopProducer()

	sqlDB, _ := config.DB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	if config.Redis != nil {
		config.Redis.Close()
	}

	fmt.Println("Server exited properly")
}

func initDefaultUser() {
	authService := service.NewAuthService()
	_, err := authService.GetUserByUsername("admin")
	if err == nil {
		return
	}

	admin := &models.User{
		Username: "admin",
		Password: "admin123",
		Email:    "admin@groundstation.com",
		FullName: "系统管理员",
		Role:     models.UserRoleAdmin,
		Status:   models.UserStatusActive,
	}

	if _, err := authService.Register(admin); err != nil {
		fmt.Printf("Warning: Failed to create default admin user: %v\n", err)
	} else {
		fmt.Println("Default admin user created: admin/admin123")
	}
}
