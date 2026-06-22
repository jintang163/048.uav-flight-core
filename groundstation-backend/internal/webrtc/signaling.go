package webrtc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/websocket"

	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
)

type SignalingManager struct {
	peers      map[uint64]*PeerConnectionContext
	peersMu    sync.RWMutex
	mediaEngine *webrtc.MediaEngine
	api         *webrtc.API
}

type PeerConnectionContext struct {
	PC          *webrtc.PeerConnection
	VideoTrack  *webrtc.TrackLocalStaticRTP
	UAVID       uint64
	LastStats   *StreamStats
	StatsMu     sync.RWMutex
	CancelStats context.CancelFunc
}

type StreamStats struct {
	BytesSent    uint64  `json:"bytes_sent"`
	PacketsSent  uint64  `json:"packets_sent"`
	PacketLoss   float64 `json:"packet_loss"`
	FramesSent   int64   `json:"frames_sent"`
	Timestamp    int64   `json:"timestamp"`
}

var signalingMgr *SignalingManager
var signalingOnce sync.Once

func NewSignalingManager() *SignalingManager {
	signalingOnce.Do(func() {
		me := &webrtc.MediaEngine{}
		if err := me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    webrtc.MimeTypeH265,
				ClockRate:   90000,
				Channels:    0,
				SDPFmtpLine: "level-id-asymmetry-allowed=1;profile-space=1",
			},
			PayloadType: 125,
		}, webrtc.RTPCodecTypeVideo); err != nil {
			middleware.Logger.Error("注册H265编解码器失败", zapError(err)...)
		}

		if err := me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeH264,
				ClockRate: 90000,
			},
			PayloadType: 96,
		}, webrtc.RTPCodecTypeVideo); err != nil {
			middleware.Logger.Error("注册H264编解码器失败", zapError(err)...)
		}

		if err := me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: 48000,
				Channels:  2,
			},
			PayloadType: 111,
		}, webrtc.RTPCodecTypeAudio); err != nil {
			middleware.Logger.Error("注册Opus编解码器失败", zapError(err)...)
		}

		api := webrtc.NewAPI(webrtc.WithMediaEngine(me))

		signalingMgr = &SignalingManager{
			peers:      make(map[uint64]*PeerConnectionContext),
			mediaEngine: me,
			api:         api,
		}
	})
	return signalingMgr
}

func (sm *SignalingManager) HandleSDPOffer(uavID uint64, sdpOffer string) (string, error) {
	sm.peersMu.Lock()
	if existing, ok := sm.peers[uavID]; ok {
		existing.CancelStats()
		existing.PC.Close()
		delete(sm.peers, uavID)
	}
	sm.peersMu.Unlock()

	pc, err := sm.api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", fmt.Errorf("创建PeerConnection失败: %w", err)
	}

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH265, ClockRate: 90000},
		"video", "uav_video",
	)
	if err != nil {
		pc.Close()
		return "", fmt.Errorf("创建视频轨道失败: %w", err)
	}

	if err := pc.AddTrack(videoTrack); err != nil {
		pc.Close()
		return "", fmt.Errorf("添加视频轨道失败: %w", err)
	}

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		middleware.Logger.Info("ICE连接状态变更",
			zap.Uint64("uav_id", uavID),
			zap.String("state", state.String()),
		)
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateDisconnected {
			sm.ClosePeerConnection(uavID)
		}
	})

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdpOffer,
	}
	if err := pc.SetRemoteDescription(offer); err != nil {
		pc.Close()
		return "", fmt.Errorf("设置远端描述失败: %w", err)
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		pc.Close()
		return "", fmt.Errorf("创建SDP应答失败: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(answer); err != nil {
		pc.Close()
		return "", fmt.Errorf("设置本地描述失败: %w", err)
	}

	<-gatherComplete

	ctx, cancel := context.WithCancel(context.Background())
	pcc := &PeerConnectionContext{
		PC:          pc,
		VideoTrack:  videoTrack,
		UAVID:       uavID,
		LastStats:   &StreamStats{},
		CancelStats: cancel,
	}

	sm.peersMu.Lock()
	sm.peers[uavID] = pcc
	sm.peersMu.Unlock()

	go sm.startStatsCollection(ctx, pcc)

	return pc.LocalDescription().SDP, nil
}

func (sm *SignalingManager) ClosePeerConnection(uavID uint64) {
	sm.peersMu.Lock()
	pcc, ok := sm.peers[uavID]
	if !ok {
		sm.peersMu.Unlock()
		return
	}
	delete(sm.peers, uavID)
	sm.peersMu.Unlock()

	pcc.CancelStats()
	pcc.PC.Close()
}

func (sm *SignalingManager) startStatsCollection(ctx context.Context, pcc *PeerConnectionContext) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := sm.collectStats(pcc)
			if stats != nil {
				pcc.StatsMu.Lock()
				pcc.LastStats = stats
				pcc.StatsMu.Unlock()

				sm.broadcastStats(pcc.UAVID, stats)
			}
		}
	}
}

func (sm *SignalingManager) collectStats(pcc *PeerConnectionContext) *StreamStats {
	statsReports := pcc.PC.GetStats()

	streamStats := &StreamStats{
		Timestamp: time.Now().UnixNano() / 1e6,
	}

	for _, report := range statsReports {
		if report.Type != webrtc.StatsTypeOutboundRTP {
			continue
		}

		kind, _ := report.Values["kind"].(string)
		if kind != "video" {
			continue
		}

		if bytesSent, ok := report.Values["bytesSent"]; ok {
			if bs, ok := bytesSent.(uint64); ok {
				streamStats.BytesSent = bs
			}
		}
		if packetsSent, ok := report.Values["packetsSent"]; ok {
			if ps, ok := packetsSent.(uint64); ok {
				streamStats.PacketsSent = ps
			}
		}
		if packetsLost, ok := report.Values["packetsLost"]; ok {
			if lost, ok := packetsLost.(uint64); ok {
				if streamStats.PacketsSent > 0 {
					streamStats.PacketLoss = float64(lost) / float64(streamStats.PacketsSent+lost) * 100
				}
			}
		}
		if framesSent, ok := report.Values["framesSent"]; ok {
			if fs, ok := framesSent.(uint64); ok {
				streamStats.FramesSent = int64(fs)
			}
		}
	}

	return streamStats
}

func (sm *SignalingManager) broadcastStats(uavID uint64, stats *StreamStats) {
	data := map[string]interface{}{
		"uav_id":       uavID,
		"uavId":        uavID,
		"bytes_sent":   stats.BytesSent,
		"packets_sent": stats.PacketsSent,
		"packet_loss":  stats.PacketLoss,
		"frames_sent":  stats.FramesSent,
		"timestamp":    stats.Timestamp,
	}
	websocket.BroadcastWebRTCStats(uavID, data)
}

func (sm *SignalingManager) GetStreamStats(uavID uint64) *StreamStats {
	sm.peersMu.RLock()
	pcc, ok := sm.peers[uavID]
	sm.peersMu.RUnlock()

	if !ok {
		return nil
	}

	pcc.StatsMu.RLock()
	defer pcc.StatsMu.RUnlock()
	return pcc.LastStats
}

func (sm *SignalingManager) Close() {
	sm.peersMu.Lock()
	defer sm.peersMu.Unlock()

	for uavID, pcc := range sm.peers {
		pcc.CancelStats()
		pcc.PC.Close()
		delete(sm.peers, uavID)
	}
}

func (sm *SignalingManager) GetPeerConnection(uavID uint64) *webrtc.PeerConnection {
	sm.peersMu.RLock()
	defer sm.peersMu.RUnlock()

	if pcc, ok := sm.peers[uavID]; ok {
		return pcc.PC
	}
	return nil
}

func (sm *SignalingManager) GetVideoTrack(uavID uint64) *webrtc.TrackLocalStaticRTP {
	sm.peersMu.RLock()
	defer sm.peersMu.RUnlock()

	if pcc, ok := sm.peers[uavID]; ok {
		return pcc.VideoTrack
	}
	return nil
}

func zapError(err error) []interface{} {
	if err == nil {
		return nil
	}
	return []interface{}{"error", err}
}
