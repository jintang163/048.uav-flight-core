package nsq

import (
	"encoding/json"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/internal/websocket"
	"sync"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

type ConsumerManager struct {
	consumers   map[string]*nsq.Consumer
	alertService *service.AlertService
	flightService *service.FlightService
	missionService *service.MissionService
}

var consumerManager *ConsumerManager
var consumerOnce sync.Once

func NewConsumerManager() *ConsumerManager {
	consumerOnce.Do(func() {
		consumerManager = &ConsumerManager{
			consumers:      make(map[string]*nsq.Consumer),
			alertService:   service.NewAlertService(),
			flightService:  service.NewFlightService(),
			missionService: service.NewMissionService(),
		}
	})
	return consumerManager
}

func (cm *ConsumerManager) StartConsumers() error {
	if err := cm.startTelemetryConsumer(); err != nil {
		return err
	}
	if err := cm.startAlertConsumer(); err != nil {
		return err
	}
	if err := cm.startMAVLinkCommandConsumer(); err != nil {
		return err
	}
	if err := cm.startGeofenceViolationConsumer(); err != nil {
		return err
	}
	return nil
}

func (cm *ConsumerManager) startTelemetryConsumer() error {
	cfg := config.AppConfig.NSQ
	config := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(TopicTelemetryData, ChannelTelemetryProcessor, config)
	if err != nil {
		return err
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var data service.TelemetryData
		if err := json.Unmarshal(message.Body, &data); err != nil {
			return nil
		}

		_ = cm.flightService.ProcessTelemetry(&data)
		return nil
	}))

	err = consumer.ConnectToNSQLookupds(cfg.LookupdAddresses)
	if err != nil {
		return err
	}

	cm.consumers[TopicTelemetryData] = consumer
	middleware.Logger.Info("NSQ consumer started", zap.String("topic", TopicTelemetryData))
	return nil
}

func (cm *ConsumerManager) startAlertConsumer() error {
	cfg := config.AppConfig.NSQ
	config := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(TopicAlertEvent, ChannelAlertNotifier, config)
	if err != nil {
		return err
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var alert models.AlertEvent
		if err := json.Unmarshal(message.Body, &alert); err != nil {
			return nil
		}

		_ = cm.alertService.SendNotifications(&alert)
		websocket.NewHub().BroadcastAlert(&alert)
		return nil
	}))

	err = consumer.ConnectToNSQLookupds(cfg.LookupdAddresses)
	if err != nil {
		return err
	}

	cm.consumers[TopicAlertEvent] = consumer
	middleware.Logger.Info("NSQ consumer started", zap.String("topic", TopicAlertEvent))
	return nil
}

func (cm *ConsumerManager) startMAVLinkCommandConsumer() error {
	cfg := config.AppConfig.NSQ
	config := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(TopicMAVLinkCommand, ChannelMAVLinkHandler, config)
	if err != nil {
		return err
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var cmd struct {
			UAVID uint64 `json:"uav_id"`
			Data  []byte `json:"data"`
		}
		if err := json.Unmarshal(message.Body, &cmd); err != nil {
			return nil
		}

		mavlinkMgr := mavlink.NewCommandManager()
		_ = mavlinkMgr.SendCommand(cmd.UAVID, cmd.Data)
		return nil
	}))

	err = consumer.ConnectToNSQLookupds(cfg.LookupdAddresses)
	if err != nil {
		return err
	}

	cm.consumers[TopicMAVLinkCommand] = consumer
	middleware.Logger.Info("NSQ consumer started", zap.String("topic", TopicMAVLinkCommand))
	return nil
}

func (cm *ConsumerManager) startGeofenceViolationConsumer() error {
	cfg := config.AppConfig.NSQ
	config := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(TopicGeofenceViolation, ChannelAlertNotifier, config)
	if err != nil {
		return err
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var violation service.GeofenceViolation
		if err := json.Unmarshal(message.Body, &violation); err != nil {
			return nil
		}

		alert := &models.AlertEvent{
			UAVID:      violation.UAVID,
			Type:       models.AlertTypeGeofenceViolation,
			Level:      models.AlertLevelCritical,
			Title:      fmt.Sprintf("电子围栏越界 - %s", violation.GeofenceName),
			Message:    fmt.Sprintf("无人机越界: 围栏[%s], 类型[%s], 距离[%.2fm]",
				violation.GeofenceName, violation.ViolationType, violation.Distance),
			Location: &models.Location{
				Latitude:  violation.Latitude,
				Longitude: violation.Longitude,
				Altitude:  violation.Altitude,
			},
		}

		_ = cm.alertService.Create(alert)
		return nil
	}))

	err = consumer.ConnectToNSQLookupds(cfg.LookupdAddresses)
	if err != nil {
		return err
	}

	cm.consumers[TopicGeofenceViolation] = consumer
	middleware.Logger.Info("NSQ consumer started", zap.String("topic", TopicGeofenceViolation))
	return nil
}

func (cm *ConsumerManager) StopConsumers() {
	for _, consumer := range cm.consumers {
		consumer.Stop()
	}
}
