package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"math"
	"time"
)

type LinkStatusReport struct {
	ActiveLink     uint8   `json:"active_link"`
	RadioRSSI      int8    `json:"radio_rssi"`
	RadioConnected bool    `json:"radio_connected"`
	LteRSSI        int8    `json:"lte_rssi"`
	LteConnected   bool    `json:"lte_connected"`
	LteNetworkType string  `json:"lte_network_type"`
	PacketLoss     float64 `json:"packet_loss"`
	LatencyMs      uint32  `json:"latency_ms"`
}

type LinkService struct {
	linkRepo *repository.LinkStatusRepository
	uavRepo  *repository.UAVRepository
}

func NewLinkService() *LinkService {
	return &LinkService{
		linkRepo: repository.NewLinkStatusRepository(),
		uavRepo:  repository.NewUAVRepository(),
	}
}

func (s *LinkService) ReportStatus(uavID uint64, req *LinkStatusReport) (*models.LinkStatus, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	radioState := s.determineLinkState(req.RadioConnected, req.RadioRSSI)
	lteState := s.determineLinkState(req.LteConnected, req.LteRSSI)

	status := &models.LinkStatus{
		UAVID:            uavID,
		ActiveLink:       models.LinkType(req.ActiveLink),
		RadioRSSI:        req.RadioRSSI,
		RadioState:       radioState,
		RadioConnected:   req.RadioConnected,
		LteRSSI:          req.LteRSSI,
		LteState:         lteState,
		LteConnected:     req.LteConnected,
		LteNetworkType:   req.LteNetworkType,
		PacketLoss:       req.PacketLoss,
		LatencyMs:        req.LatencyMs,
		Timestamp:        time.Now(),
		AutoSwitchEnabled: true,
	}

	if err := s.linkRepo.Create(status); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateLastSeen(uavID)

	websocket.BroadcastLinkStatus(uavID, status)

	return status, nil
}

func (s *LinkService) determineLinkState(connected bool, rssi int8) models.LinkState {
	if !connected {
		return models.LinkStateDisconnected
	}
	if rssi >= -70 {
		return models.LinkStateConnected
	}
	if rssi >= -90 {
		return models.LinkStateDegraded
	}
	return models.LinkStateConnecting
}

func (s *LinkService) GetLatestByUAVID(uavID uint64) (*models.LinkStatus, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}
	return s.linkRepo.GetLatestByUAVID(uavID)
}

func (s *LinkService) GetHistory(uavID uint64, page, pageSize int, startTime, endTime *time.Time) ([]*models.LinkStatus, int64, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, 0, errors.New("uav not found")
	}
	return s.linkRepo.ListByUAVID(uavID, page, pageSize, startTime, endTime)
}

func (s *LinkService) GetStatistics(uavID *uint64, startTime, endTime *time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	activeLinkCounts, err := s.linkRepo.GetCurrentActiveLinkCount()
	if err != nil {
		return nil, err
	}

	linkTypeStats := make(map[string]int64)
	for linkType, count := range activeLinkCounts {
		switch linkType {
		case models.LinkTypeRadio:
			linkTypeStats["radio"] = count
		case models.LinkTypeLTE:
			linkTypeStats["lte"] = count
		case models.LinkTypeDual:
			linkTypeStats["dual"] = count
		}
	}
	result["active_link_distribution"] = linkTypeStats

	if uavID != nil {
		history, _, err := s.linkRepo.ListByUAVID(*uavID, 1, 1000, startTime, endTime)
		if err != nil {
			return nil, err
		}

		if len(history) > 0 {
			var avgLatency, avgPacketLoss float64
			var minLatency, maxLatency uint32
			var totalRadioConnected, totalLteConnected int
			var radioRSSISum, lteRSSISum int

			minLatency = math.MaxUint32

			for _, h := range history {
				avgLatency += float64(h.LatencyMs)
				avgPacketLoss += h.PacketLoss
				if h.LatencyMs < minLatency {
					minLatency = h.LatencyMs
				}
				if h.LatencyMs > maxLatency {
					maxLatency = h.LatencyMs
				}
				if h.RadioConnected {
					totalRadioConnected++
				}
				if h.LteConnected {
					totalLteConnected++
				}
				radioRSSISum += int(h.RadioRSSI)
				lteRSSISum += int(h.LteRSSI)
			}

			count := float64(len(history))
			avgLatency /= count
			avgPacketLoss /= count

			result["avg_latency_ms"] = avgLatency
			result["min_latency_ms"] = minLatency
			result["max_latency_ms"] = maxLatency
			result["avg_packet_loss"] = avgPacketLoss
			result["radio_connectivity_rate"] = float64(totalRadioConnected) / count
			result["lte_connectivity_rate"] = float64(totalLteConnected) / count
			result["avg_radio_rssi"] = float64(radioRSSISum) / count
			result["avg_lte_rssi"] = float64(lteRSSISum) / count
			result["total_records"] = len(history)
		}
	}

	return result, nil
}

func (s *LinkService) CleanupOldHistory() (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -30)
	return s.linkRepo.DeleteOldRecords(cutoff)
}
