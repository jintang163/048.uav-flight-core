package webrtc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"groundstation-backend/internal/middleware"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
)

type StreamProxy struct {
	upstreamURLs map[uint64]string
	upstreamMu   sync.RWMutex
	signaling    *SignalingManager
	cancelMap    map[uint64]context.CancelFunc
	cancelMu     sync.Mutex
}

var streamProxy *StreamProxy
var proxyOnce sync.Once

func NewStreamProxy(sm *SignalingManager) *StreamProxy {
	proxyOnce.Do(func() {
		streamProxy = &StreamProxy{
			upstreamURLs: make(map[uint64]string),
			signaling:    sm,
			cancelMap:    make(map[uint64]context.CancelFunc),
		}
	})
	return streamProxy
}

func (p *StreamProxy) RegisterStream(uavID uint64, upstreamURL string) {
	p.upstreamMu.Lock()
	p.upstreamURLs[uavID] = upstreamURL
	p.upstreamMu.Unlock()
}

func (p *StreamProxy) RemoveStream(uavID uint64) {
	p.upstreamMu.Lock()
	delete(p.upstreamURLs, uavID)
	p.upstreamMu.Unlock()

	p.cancelMu.Lock()
	if cancel, ok := p.cancelMap[uavID]; ok {
		cancel()
		delete(p.cancelMap, uavID)
	}
	p.cancelMu.Unlock()
}

func (p *StreamProxy) StartForwarding(ctx context.Context, uavID uint64) error {
	p.upstreamMu.RLock()
	upstreamURL, ok := p.upstreamURLs[uavID]
	p.upstreamMu.RUnlock()

	if !ok {
		return fmt.Errorf("未注册上游流: uav_%d", uavID)
	}

	fwdCtx, cancel := context.WithCancel(ctx)
	p.cancelMu.Lock()
	if existing, exists := p.cancelMap[uavID]; exists {
		existing()
	}
	p.cancelMap[uavID] = cancel
	p.cancelMu.Unlock()

	go p.forwardWithRetry(fwdCtx, uavID, upstreamURL)

	return nil
}

func (p *StreamProxy) forwardWithRetry(ctx context.Context, uavID uint64, upstreamURL string) {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := p.forwardStream(ctx, uavID, upstreamURL)
		if err == nil {
			return
		}

		middleware.Logger.Warn("上游流转发失败，准备重试",
			zap.Uint64("uav_id", uavID),
			zap.String("url", upstreamURL),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (p *StreamProxy) forwardStream(ctx context.Context, uavID uint64, upstreamURL string) error {
	videoTrack := p.signaling.GetVideoTrack(uavID)
	if videoTrack == nil {
		return fmt.Errorf("视频轨道不存在: uav_%d", uavID)
	}

	if isRTSPURL(upstreamURL) {
		return p.forwardFromRTSP(ctx, uavID, upstreamURL, videoTrack)
	}

	return p.forwardFromRTP(ctx, uavID, upstreamURL, videoTrack)
}

func (p *StreamProxy) forwardFromRTSP(ctx context.Context, uavID uint64, rtspURL string, videoTrack *webrtc.TrackLocalStaticRTP) error {
	conn, err := net.DialTimeout("tcp", extractHost(rtspURL), 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接RTSP源失败: %w", err)
	}
	defer conn.Close()

	buf := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("读取RTSP数据失败: %w", err)
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(buf[:n]); err != nil {
			continue
		}

		if err := videoTrack.WriteRTP(packet); err != nil {
			continue
		}
	}
}

func (p *StreamProxy) forwardFromRTP(ctx context.Context, uavID uint64, rtpAddr string, videoTrack *webrtc.TrackLocalStaticRTP) error {
	_, err := net.ResolveUDPAddr("udp", rtpAddr)
	if err != nil {
		return fmt.Errorf("解析RTP地址失败: %w", err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		return fmt.Errorf("监听UDP失败: %w", err)
	}
	defer conn.Close()

	sendAddr, err := net.ResolveUDPAddr("udp", rtpAddr)
	if err != nil {
		return fmt.Errorf("解析RTP目标地址失败: %w", err)
	}

	_, err = conn.WriteToUDP([]byte("PING"), sendAddr)
	if err != nil {
		return fmt.Errorf("发送RTP注册失败: %w", err)
	}

	buf := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("读取RTP数据失败: %w", err)
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(buf[:n]); err != nil {
			continue
		}

		if err := videoTrack.WriteRTP(packet); err != nil {
			continue
		}
	}
}

func isRTSPURL(url string) bool {
	return len(url) >= 7 && (url[:7] == "rtsp://" || url[:7] == "rtsps://")
}

func extractHost(rtspURL string) string {
	start := 7
	if len(rtspURL) >= 8 && rtspURL[:8] == "rtsps://" {
		start = 8
	}
	rest := rtspURL[start:]
	for i, c := range rest {
		if c == '/' || c == '?' {
			return rest[:i]
		}
	}
	return rest
}

func (p *StreamProxy) GetUpstreamURL(uavID uint64) (string, bool) {
	p.upstreamMu.RLock()
	defer p.upstreamMu.RUnlock()
	url, ok := p.upstreamURLs[uavID]
	return url, ok
}

func (p *StreamProxy) Close() {
	p.cancelMu.Lock()
	for uavID, cancel := range p.cancelMap {
		cancel()
		delete(p.cancelMap, uavID)
	}
	p.cancelMu.Unlock()
}
