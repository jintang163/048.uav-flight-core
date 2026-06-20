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
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}
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
			blackbox.GET("", handler.ListBlackboxes)
			blackbox.GET("/statistics", handler.GetBlackboxStatistics)
			blackbox.GET("/:id", handler.GetBlackbox)
			blackbox.PUT("/:id", middleware.RoleAuth(models.UserRoleAdmin, models.UserRoleOperator), handler.UpdateBlackbox)
			blackbox.DELETE("/:id", middleware.RoleAuth(models.UserRoleAdmin), handler.DeleteBlackbox)
			blackbox.GET("/:id/download", handler.DownloadBlackbox)
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
	}

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
