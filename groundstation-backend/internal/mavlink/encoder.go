package mavlink

import (
	"encoding/binary"
	"math"
)

func EncodeCommandLong(uavID uint64, command uint16, params ...float32) []byte {
	payload := make([]byte, 33)

	param1 := float32(0)
	param2 := float32(0)
	param3 := float32(0)
	param4 := float32(0)
	param5 := float32(0)
	param6 := float32(0)
	param7 := float32(0)

	if len(params) > 0 {
		param1 = params[0]
	}
	if len(params) > 1 {
		param2 = params[1]
	}
	if len(params) > 2 {
		param3 = params[2]
	}
	if len(params) > 3 {
		param4 = params[3]
	}
	if len(params) > 4 {
		param5 = params[4]
	}
	if len(params) > 5 {
		param6 = params[5]
	}
	if len(params) > 6 {
		param7 = params[6]
	}

	binary.LittleEndian.PutUint32(payload[0:4], math.Float32bits(param1))
	binary.LittleEndian.PutUint32(payload[4:8], math.Float32bits(param2))
	binary.LittleEndian.PutUint32(payload[8:12], math.Float32bits(param3))
	binary.LittleEndian.PutUint32(payload[12:16], math.Float32bits(param4))
	binary.LittleEndian.PutUint32(payload[16:20], math.Float32bits(param5))
	binary.LittleEndian.PutUint32(payload[20:24], math.Float32bits(param6))
	binary.LittleEndian.PutUint32(payload[24:28], math.Float32bits(param7))
	binary.LittleEndian.PutUint16(payload[28:30], command)
	payload[30] = 1
	payload[31] = 1
	payload[32] = 0

	return EncodeMessage(76, payload)
}

func EncodeMissionCount(count uint16) []byte {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], count)
	payload[2] = 1
	payload[3] = 0
	return EncodeMessage(44, payload)
}

func EncodeMissionItem(seq uint16, lat, lng, alt float64) []byte {
	payload := make([]byte, 37)

	binary.LittleEndian.PutUint32(payload[0:4], math.Float32bits(float32(0)))
	binary.LittleEndian.PutUint32(payload[4:8], math.Float32bits(float32(0)))
	binary.LittleEndian.PutUint32(payload[8:12], math.Float32bits(float32(0)))
	binary.LittleEndian.PutUint32(payload[12:16], math.Float32bits(float32(alt)))
	binary.LittleEndian.PutUint32(payload[16:20], math.Float32bits(float32(lat)))
	binary.LittleEndian.PutUint32(payload[20:24], math.Float32bits(float32(lng)))
	binary.LittleEndian.PutUint16(payload[24:26], seq)
	payload[26] = 1
	payload[27] = 0
	binary.LittleEndian.PutUint16(payload[28:30], 16)
	payload[30] = 1
	payload[31] = 0
	payload[32] = 0
	payload[33] = 0
	payload[34] = 0
	payload[35] = 1
	payload[36] = 0

	return EncodeMessage(39, payload)
}

func EncodeMessage(messageID uint32, payload []byte) []byte {
	length := len(payload)
	buf := make([]byte, 12+length+2)

	buf[0] = MAVLINK_STX
	buf[1] = byte(length)
	buf[2] = 0
	buf[3] = 0
	buf[4] = 0
	buf[5] = 1
	buf[6] = 1
	buf[7] = byte(messageID & 0xFF)
	buf[8] = byte((messageID >> 8) & 0xFF)
	buf[9] = byte((messageID >> 16) & 0xFF)

	copy(buf[12:12+length], payload)

	checksum := calculateChecksum(buf[1 : 12+length])
	binary.LittleEndian.PutUint16(buf[12+length:], checksum)

	return buf
}

func calculateChecksum(data []byte) uint16 {
	var crc uint16
	for _, b := range data {
		tmp := uint16(b) ^ (crc & 0xFF)
		tmp ^= (tmp << 4) & 0xFF
		crc = (crc >> 8) ^ (tmp << 8) ^ (tmp << 3) ^ (tmp >> 4)
	}
	return crc
}

func EncodeHeartbeat() []byte {
	payload := make([]byte, 9)
	payload[0] = 2
	payload[1] = 12
	payload[2] = 0
	binary.LittleEndian.PutUint32(payload[3:7], 0)
	payload[7] = 4
	payload[8] = 3

	return EncodeMessage(HEARTBEAT, payload)
}

func EncodeSetMode(customMode uint32) []byte {
	return EncodeCommandLong(1, CMD_DO_SET_MODE, 1, float32(customMode), 0, 0, 0, 0, 0)
}

func EncodeArmDisarm(arm bool) []byte {
	param1 := float32(0)
	if arm {
		param1 = 1
	}
	return EncodeCommandLong(1, CMD_COMPONENT_ARM_DISARM, param1, 0, 0, 0, 0, 0, 0)
}
