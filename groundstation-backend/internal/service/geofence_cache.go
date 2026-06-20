package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
)

const (
	cacheKeyAllActive     = "geofence:all_active"
	cacheKeyUAVPrefix     = "geofence:uav:"
	cacheKeyNational       = "geofence:national"
	cacheTTL          = 5 * time.Minute
)

type GeofenceCacheService struct {
	repo       *repository.GeofenceRepository
	mu         *sync.RWMutex
	localCache map[uint64][]models.Geofence
	allActive  []models.Geofence
	national   []models.Geofence
	lastUpdate time.Time
}

func NewGeofenceCacheService() *GeofenceCacheService {
	return &GeofenceCacheService{
		repo:       repository.NewGeofenceRepository(),
		mu:         &sync.RWMutex{},
		localCache: make(map[uint64][]models.Geofence),
	}
}

func (c *GeofenceCacheService) GetUAVGeofences(uavID uint64) ([]models.Geofence, error) {
	c.mu.RLock()
	if gfs, ok := c.localCache[uavID]; ok && time.Since(c.lastUpdate) < cacheTTL {
		c.mu.RUnlock()
		return gfs, nil
	}
	c.mu.RUnlock()

	gfs, err := c.loadFromRedis(uavID)
	if err == nil && len(gfs) > 0 {
		c.mu.Lock()
		c.localCache[uavID] = gfs
		c.mu.Unlock()
		return gfs, nil
	}

	gfs, err = c.repo.GetUAVGeofences(uavID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.localCache[uavID] = gfs
	c.lastUpdate = time.Now()
	c.mu.Unlock()

	go c.saveToRedis(uavID, gfs)

	return gfs, nil
}

func (c *GeofenceCacheService) GetAllActive() ([]models.Geofence, error) {
	c.mu.RLock()
	if c.allActive != nil && time.Since(c.lastUpdate) < cacheTTL {
		c.mu.RUnlock()
		return c.allActive, nil
	}
	c.mu.RUnlock()

	gfs, err := c.loadAllFromRedis()
	if err == nil && len(gfs) > 0 {
		c.mu.Lock()
		c.allActive = gfs
		c.mu.Unlock()
		return gfs, nil
	}

	gfs, err = c.repo.GetActiveGeofences()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.allActive = gfs
	c.lastUpdate = time.Now()
	c.mu.Unlock()

	go c.saveAllToRedis(gfs)

	return gfs, nil
}

func (c *GeofenceCacheService) GetNational() ([]models.Geofence, error) {
	c.mu.RLock()
	if c.national != nil && time.Since(c.lastUpdate) < cacheTTL {
		c.mu.RUnlock()
		return c.national, nil
	}
	c.mu.RUnlock()

	gfs, err := c.loadNationalFromRedis()
	if err == nil && len(gfs) > 0 {
		c.mu.Lock()
		c.national = gfs
		c.mu.Unlock()
		return gfs, nil
	}

	pagination := &utils.Pagination{Page: 1, PageSize: 2000}
	isActive := true
	gfs, _, err = c.repo.ListFiltered(pagination, "", &isActive, models.GeofenceCategory(""), models.GeofenceSourceNational)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.national = gfs
	c.lastUpdate = time.Now()
	c.mu.Unlock()

	go c.saveNationalToRedis(gfs)

	return gfs, nil
}

func (c *GeofenceCacheService) Refresh() {
	c.mu.Lock()
	c.localCache = make(map[uint64][]models.Geofence)
	c.allActive = nil
	c.national = nil
	c.lastUpdate = time.Time{}
	c.mu.Unlock()

	if config.Redis != nil {
		ctx := context.Background()
		_ = config.Redis.Del(ctx, cacheKeyAllActive).Err()
		_ = config.Redis.Del(ctx, cacheKeyNational).Err()
		keys, _ := config.Redis.Keys(ctx, cacheKeyUAVPrefix+"*").Result()
		if len(keys) > 0 {
			_ = config.Redis.Del(ctx, keys...).Err()
		}
	}
}

func (c *GeofenceCacheService) loadFromRedis(uavID uint64) ([]models.Geofence, error) {
	if config.Redis == nil {
		return nil, errors.New("redis not available")
	}
	key := cacheKeyUAVPrefix + strconv.FormatUint(uavID, 10)
	ctx := context.Background()
	data, err := config.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return c.decompressGeofences(data)
}

func (c *GeofenceCacheService) loadAllFromRedis() ([]models.Geofence, error) {
	if config.Redis == nil {
		return nil, errors.New("redis not available")
	}
	ctx := context.Background()
	data, err := config.Redis.Get(ctx, cacheKeyAllActive).Bytes()
	if err != nil {
		return nil, err
	}
	return c.decompressGeofences(data)
}

func (c *GeofenceCacheService) loadNationalFromRedis() ([]models.Geofence, error) {
	if config.Redis == nil {
		return nil, errors.New("redis not available")
	}
	ctx := context.Background()
	data, err := config.Redis.Get(ctx, cacheKeyNational).Bytes()
	if err != nil {
		return nil, err
	}
	return c.decompressGeofences(data)
}

func (c *GeofenceCacheService) saveToRedis(uavID uint64, gfs []models.Geofence) {
	if config.Redis == nil {
		return
	}
	key := cacheKeyUAVPrefix + strconv.FormatUint(uavID, 10)
	data, err := c.compressGeofences(gfs)
	if err != nil {
		return
	}
	ctx := context.Background()
	_ = config.Redis.Set(ctx, key, data, cacheTTL).Err()
}

func (c *GeofenceCacheService) saveAllToRedis(gfs []models.Geofence) {
	if config.Redis == nil {
		return
	}
	data, err := c.compressGeofences(gfs)
	if err != nil {
		return
	}
	ctx := context.Background()
	_ = config.Redis.Set(ctx, cacheKeyAllActive, data, cacheTTL).Err()
}

func (c *GeofenceCacheService) saveNationalToRedis(gfs []models.Geofence) {
	if config.Redis == nil {
		return
	}
	data, err := c.compressGeofences(gfs)
	if err != nil {
		return
	}
	ctx := context.Background()
	_ = config.Redis.Set(ctx, cacheKeyNational, data, 24*time.Hour).Err()
}

func (c *GeofenceCacheService) compressGeofences(gfs []models.Geofence) ([]byte, error) {
	jsonData, err := json.Marshal(gfs)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *GeofenceCacheService) decompressGeofences(data []byte) ([]models.Geofence, error) {
	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	jsonData, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	var gfs []models.Geofence
	if err := json.Unmarshal(jsonData, &gfs); err != nil {
		return nil, err
	}
	return gfs, nil
}

func (c *GeofenceCacheService) InvalidateUAV(uavID uint64) {
	c.mu.Lock()
	delete(c.localCache, uavID)
	c.mu.Unlock()

	if config.Redis != nil {
		key := cacheKeyUAVPrefix + strconv.FormatUint(uavID, 10)
		ctx := context.Background()
		_ = config.Redis.Del(ctx, key).Err()
	}
}

func (c *GeofenceCacheService) GetByBounds(minLat, maxLat, minLng, maxLng float64) ([]models.Geofence, error) {
	all, err := c.GetAllActive()
	if err != nil {
		return nil, err
	}

	var result []models.Geofence
	for _, gf := range all {
		if gf.Shape == models.GeofenceShapeCircle {
			if gf.CenterLat >= minLat && gf.CenterLat <= maxLat &&
				gf.CenterLng >= minLng && gf.CenterLng <= maxLng {
				result = append(result, gf)
			}
		} else if gf.Shape == models.GeofenceShapePolygon {
			coords := c.parseCoords(gf.Coordinates)
			if len(coords) > 0 {
				for _, coord := range coords {
					if coord.Lat >= minLat && coord.Lat <= maxLat &&
						coord.Lng >= minLng && coord.Lng <= maxLng {
						result = append(result, gf)
						break
					}
				}
			}
		}
	}
	return result, nil
}

func (c *GeofenceCacheService) parseCoords(data string) []models.Coordinate {
	if data == "" {
		return nil
	}
	var coords []models.Coordinate
	if err := json.Unmarshal([]byte(data), &coords); err != nil {
		return nil
	}
	return coords
}
