import React, { useEffect, useRef, useState, useCallback } from 'react'
import AMapLoader from '@amap/amap-jsapi-loader'
import styled from 'styled-components'
import { Button, message } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined, SaveOutlined, CloseOutlined } from '@ant-design/icons'
import { useGeofence } from '@/hooks/useGeofence'
import { useMission } from '@/hooks/useMission'
import type { Waypoint, Geofence, GeofenceType } from '@/types'
import { generateId, calculateDistance } from '@/utils'

const MapContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`

const MapTools = styled.div`
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: rgba(0, 0, 0, 0.8);
  padding: 12px;
  border-radius: 8px;
`

const MapInfo = styled.div`
  position: absolute;
  bottom: 16px;
  left: 16px;
  z-index: 100;
  background: rgba(0, 0, 0, 0.8);
  padding: 12px 16px;
  border-radius: 8px;
  color: white;
  font-size: 12px;
  font-family: monospace;
`

interface FlightMapProps {
  uavPosition?: { lat: number; lng: number; alt: number; heading: number }
  trajectory?: { lat: number; lng: number }[]
  showTrajectory?: boolean
  showGeofence?: boolean
  showMission?: boolean
  editable?: boolean
  onWaypointAdd?: (waypoint: Waypoint) => void
  onWaypointUpdate?: (waypoint: Waypoint) => void
  onWaypointDelete?: (waypointId: string) => void
}

let amap: typeof AMap | null = null

const FlightMap: React.FC<FlightMapProps> = ({
  uavPosition,
  trajectory = [],
  showTrajectory = true,
  showGeofence = true,
  showMission = true,
  editable = false,
  onWaypointAdd,
  onWaypointUpdate,
  onWaypointDelete
}) => {
  const mapRef = useRef<HTMLDivElement>(null)
  const mapInstance = useRef<AMap.Map | null>(null)
  const uavMarker = useRef<AMap.Marker | null>(null)
  const trajectoryPolyline = useRef<AMap.Polyline | null>(null)
  const missionPolyline = useRef<AMap.Polyline | null>(null)
  const waypointMarkers = useRef<Map<string, AMap.Marker>>(new Map())
  const geofencePolygons = useRef<Map<string, AMap.Polygon | AMap.Circle>>(new Map())
  const drawingPolygon = useRef<AMap.Polygon | null>(null)
  const drawingPoints = useRef<{ lat: number; lng: number }[]>([])

  const { geofences, drawMode, drawType, startDraw, stopDraw, createGeofence } = useGeofence()
  const { waypoints, addWaypoint, updateWaypoint, deleteWaypoint } = useMission()
  const [mapLoaded, setMapLoaded] = useState(false)
  const [cursorPosition, setCursorPosition] = useState<{ lat: number; lng: number } | null>(null)

  useEffect(() => {
    const initMap = async () => {
      try {
        if (typeof window !== 'undefined' && !amap) {
          const securityCode = import.meta.env.VITE_AMAP_SECURITY_CODE
          if (securityCode) {
            (window as unknown as { _AMapSecurityConfig: { securityJsCode: string } })._AMapSecurityConfig = {
              securityJsCode: securityCode
            }
          }

          amap = await AMapLoader.load({
            key: import.meta.env.VITE_AMAP_KEY || '',
            version: '2.0',
            plugins: [
              'AMap.Scale',
              'AMap.ToolBar',
              'AMap.ControlBar',
              'AMap.MouseTool',
              'AMap.PolyEditor',
              'AMap.CircleEditor'
            ]
          })
        }

        if (mapRef.current && amap && !mapInstance.current) {
          mapInstance.current = new amap.Map(mapRef.current, {
            zoom: 15,
            center: [116.397428, 39.90923],
            viewMode: '3D',
            pitch: 0,
            mapStyle: 'amap://styles/dark'
          })

          mapInstance.current.addControl(new amap.Scale())
          mapInstance.current.addControl(new amap.ToolBar({ position: 'LT' }))
          mapInstance.current.addControl(new amap.ControlBar({ position: 'RT' }))

          mapInstance.current.on('mousemove', (e: { lnglat: { getLng: () => number; getLat: () => number } }) => {
            setCursorPosition({
              lng: e.lnglat.getLng(),
              lat: e.lnglat.getLat()
            })
          })

          if (editable) {
            mapInstance.current.on('click', handleMapClick)
          }

          setMapLoaded(true)
        }
      } catch (error) {
        console.error('Failed to load map:', error)
        message.error('地图加载失败')
      }
    }

    initMap()

    return () => {
      if (mapInstance.current) {
        mapInstance.current.destroy()
        mapInstance.current = null
      }
    }
  }, [editable])

  const handleMapClick = useCallback((e: { lnglat: { getLng: () => number; getLat: () => number } }) => {
    if (drawMode && drawType && amap && mapInstance.current) {
      const point = {
        lng: e.lnglat.getLng(),
        lat: e.lnglat.getLat()
      }
      drawingPoints.current.push(point)

      if (drawType === 'circle') {
        if (drawingPoints.current.length === 2) {
          const center = drawingPoints.current[0]
          const radius = calculateDistance(center.lat, center.lng, point.lat, point.lng)
          
          if (drawingPolygon.current) {
            (drawingPolygon.current as AMap.Circle).setCenter([center.lng, center.lat])
            ;(drawingPolygon.current as AMap.Circle).setRadius(radius)
          } else {
            drawingPolygon.current = new amap.Circle({
              center: [center.lng, center.lat],
              radius: radius,
              strokeColor: '#1890ff',
              strokeWeight: 2,
              fillColor: '#1890ff',
              fillOpacity: 0.2,
              map: mapInstance.current
            })
          }
          finishDrawing()
        } else if (drawingPoints.current.length === 1) {
          drawingPolygon.current = new amap.Circle({
            center: [point.lng, point.lat],
            radius: 0,
            strokeColor: '#1890ff',
            strokeWeight: 2,
            fillColor: '#1890ff',
            fillOpacity: 0.2,
            map: mapInstance.current
          })
        }
      } else if (drawType === 'polygon') {
        const path = drawingPoints.current.map(p => [p.lng, p.lat])
        if (drawingPolygon.current) {
          (drawingPolygon.current as AMap.Polygon).setPath(path as unknown as AMap.Vector[])
        } else {
          drawingPolygon.current = new amap.Polygon({
            path: path as unknown as AMap.Vector[],
            strokeColor: '#1890ff',
            strokeWeight: 2,
            fillColor: '#1890ff',
            fillOpacity: 0.2,
            map: mapInstance.current
          })
        }
      }
    } else if (editable && !drawMode && uavPosition && amap && mapInstance.current) {
      const lng = e.lnglat.getLng()
      const lat = e.lnglat.getLat()
      const alt = 50

      const newWaypoint: Waypoint = {
        id: generateId(),
        sequence: waypoints.length + 1,
        action: 'waypoint',
        lat,
        lng,
        altitude: alt,
        parameters: {}
      }

      addWaypoint(newWaypoint)
      onWaypointAdd?.(newWaypoint)
      addWaypointMarker(newWaypoint)
    }
  }, [drawMode, drawType, editable, uavPosition, waypoints.length, addWaypoint, onWaypointAdd])

  const finishDrawing = useCallback(() => {
    if (drawMode && drawingPolygon.current && drawingPoints.current.length > 0) {
      let shape: { type: GeofenceType; center?: { lat: number; lng: number }; radius?: number; points?: { lat: number; lng: number }[] }

      if (drawType === 'circle') {
        const circle = drawingPolygon.current as AMap.Circle
        const center = circle.getCenter() as AMap.LngLat
        shape = {
          type: 'circle',
          center: { lat: center.getLat(), lng: center.getLng() },
          radius: circle.getRadius()
        }
      } else if (drawType === 'polygon') {
        shape = {
          type: 'polygon',
          points: drawingPoints.current
        }
      } else {
        return
      }

      createGeofence({
        name: `围栏 ${new Date().toLocaleString()}`,
        shape: shape as never,
        action: 'warn',
        isInclusion: false,
        isEnabled: true
      })

      drawingPolygon.current.setMap(null)
      drawingPolygon.current = null
      drawingPoints.current = []
      stopDraw()
      message.success('围栏创建成功')
    }
  }, [drawMode, drawType, createGeofence, stopDraw])

  const addWaypointMarker = useCallback((waypoint: Waypoint) => {
    if (!amap || !mapInstance.current) return

    const marker = new amap.Marker({
      position: [waypoint.lng, waypoint.lat],
      map: mapInstance.current,
      content: `
        <div style="
          width: 24px;
          height: 24px;
          background: #1890ff;
          border: 2px solid white;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-size: 12px;
          font-weight: bold;
          box-shadow: 0 2px 8px rgba(0,0,0,0.5);
        ">${waypoint.sequence}</div>
      `,
      draggable: editable,
      extData: { waypointId: waypoint.id }
    })

    if (editable) {
      marker.on('dragend', (e: { lnglat: { getLng: () => number; getLat: () => number } }) => {
        const updatedWaypoint: Waypoint = {
          ...waypoint,
          lng: e.lnglat.getLng(),
          lat: e.lnglat.getLat()
        }
        updateWaypoint(updatedWaypoint)
        onWaypointUpdate?.(updatedWaypoint)
      })

      marker.on('rightclick', () => {
        if (confirm(`确定删除航点 ${waypoint.sequence}？`)) {
          marker.setMap(null)
          waypointMarkers.current.delete(waypoint.id)
          deleteWaypoint(waypoint.id)
          onWaypointDelete?.(waypoint.id)
        }
      })
    }

    waypointMarkers.current.set(waypoint.id, marker)
    updateMissionPolyline()
  }, [editable, updateWaypoint, deleteWaypoint, onWaypointUpdate, onWaypointDelete])

  const updateMissionPolyline = useCallback(() => {
    if (!amap || !mapInstance.current || waypoints.length < 2) {
      if (missionPolyline.current) {
        missionPolyline.current.setMap(null)
        missionPolyline.current = null
      }
      return
    }

    const path = waypoints.map(w => [w.lng, w.lat])

    if (missionPolyline.current) {
      missionPolyline.current.setPath(path as unknown as AMap.Vector[])
    } else {
      missionPolyline.current = new amap.Polyline({
        path: path as unknown as AMap.Vector[],
        strokeColor: '#1890ff',
        strokeWeight: 3,
        strokeStyle: 'dashed',
        strokeOpacity: 0.8,
        map: mapInstance.current,
        showDir: true
      })
    }
  }, [waypoints])

  useEffect(() => {
    if (!mapInstance.current || !amap || !uavPosition) return

    const position = new amap.LngLat(uavPosition.lng, uavPosition.lat)

    if (uavMarker.current) {
      uavMarker.current.setPosition(position)
      uavMarker.current.setAngle(uavPosition.heading)
    } else {
      const content = `
        <div style="
          width: 40px;
          height: 40px;
          background: #52c41a;
          border: 3px solid white;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          box-shadow: 0 0 15px rgba(82, 196, 26, 0.6);
          position: relative;
        ">
          <div style="
            position: absolute;
            width: 0;
            height: 0;
            border-left: 6px solid transparent;
            border-right: 6px solid transparent;
            border-bottom: 12px solid white;
            top: -8px;
          "></div>
        </div>
      `
      uavMarker.current = new amap.Marker({
        position: position,
        map: mapInstance.current,
        content: content,
        anchor: 'center',
        angle: uavPosition.heading
      })
    }

    mapInstance.current.setCenter(position)
  }, [uavPosition])

  useEffect(() => {
    if (!mapInstance.current || !amap || !showTrajectory || trajectory.length < 2) return

    const path = trajectory.map(p => [p.lng, p.lat])

    if (trajectoryPolyline.current) {
      trajectoryPolyline.current.setPath(path as unknown as AMap.Vector[])
    } else {
      trajectoryPolyline.current = new amap.Polyline({
        path: path as unknown as AMap.Vector[],
        strokeColor: '#13c2c2',
        strokeWeight: 2,
        strokeOpacity: 0.8,
        map: mapInstance.current
      })
    }
  }, [trajectory, showTrajectory])

  useEffect(() => {
    if (!mapInstance.current || !amap || !showGeofence) return

    geofencePolygons.current.forEach(polygon => polygon.setMap(null))
    geofencePolygons.current.clear()

    geofences.forEach(geofence => {
      if (!geofence.isEnabled) return

      let polygon: AMap.Polygon | AMap.Circle

      if (geofence.shape.type === 'circle') {
        const shape = geofence.shape as { type: 'circle'; center: { lat: number; lng: number }; radius: number }
        polygon = new amap.Circle({
          center: [shape.center.lng, shape.center.lat],
          radius: shape.radius,
          strokeColor: geofence.isInclusion ? '#52c41a' : '#ff4d4f',
          strokeWeight: 2,
          fillColor: geofence.isInclusion ? '#52c41a' : '#ff4d4f',
          fillOpacity: 0.1,
          map: mapInstance.current
        })
      } else if (geofence.shape.type === 'polygon') {
        const shape = geofence.shape as { type: 'polygon'; points: { lat: number; lng: number }[] }
        const path = shape.points.map(p => [p.lng, p.lat])
        polygon = new amap.Polygon({
          path: path as unknown as AMap.Vector[],
          strokeColor: geofence.isInclusion ? '#52c41a' : '#ff4d4f',
          strokeWeight: 2,
          fillColor: geofence.isInclusion ? '#52c41a' : '#ff4d4f',
          fillOpacity: 0.1,
          map: mapInstance.current
        })
      } else {
        return
      }

      geofencePolygons.current.set(geofence.id, polygon)
    })
  }, [geofences, showGeofence])

  useEffect(() => {
    if (!showMission) {
      waypointMarkers.current.forEach(marker => marker.setMap(null))
      waypointMarkers.current.clear()
      if (missionPolyline.current) {
        missionPolyline.current.setMap(null)
        missionPolyline.current = null
      }
      return
    }

    waypointMarkers.current.forEach(marker => marker.setMap(null))
    waypointMarkers.current.clear()

    waypoints.forEach(waypoint => {
      addWaypointMarker(waypoint)
    })
  }, [showMission, waypoints, addWaypointMarker])

  const handleDrawStart = (type: GeofenceType) => {
    startDraw(type)
    drawingPoints.current = []
    message.info('点击地图开始绘制')
  }

  const handleDrawCancel = () => {
    if (drawingPolygon.current) {
      drawingPolygon.current.setMap(null)
      drawingPolygon.current = null
    }
    drawingPoints.current = []
    stopDraw()
  }

  return (
    <MapContainer>
      <div ref={mapRef} style={{ width: '100%', height: '100%' }} />
      
      {editable && (
        <MapTools>
          {!drawMode ? (
            <>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleDrawStart('polygon')}
                size="small"
              >
                绘制多边形
              </Button>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleDrawStart('circle')}
                size="small"
              >
                绘制圆形
              </Button>
            </>
          ) : (
            <>
              <Button
                type="success"
                icon={<SaveOutlined />}
                onClick={finishDrawing}
                size="small"
              >
                完成
              </Button>
              <Button
                danger
                icon={<CloseOutlined />}
                onClick={handleDrawCancel}
                size="small"
              >
                取消
              </Button>
            </>
          )}
        </MapTools>
      )}

      {mapLoaded && cursorPosition && (
        <MapInfo>
          <div>鼠标位置: {cursorPosition.lng.toFixed(6)}, {cursorPosition.lat.toFixed(6)}</div>
          {uavPosition && (
            <div>无人机位置: {uavPosition.lng.toFixed(6)}, {uavPosition.lat.toFixed(6)}, {uavPosition.alt.toFixed(1)}m</div>
          )}
          {trajectory.length > 0 && (
            <div>轨迹点数: {trajectory.length}</div>
          )}
        </MapInfo>
      )}
    </MapContainer>
  )
}

export default FlightMap
