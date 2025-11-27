package main

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
	"syscall"
	"os"
	"os/signal"
)

// MAVLink2 constants
const (
	stx          = 0xFD // MAVLink2 magic byte
	msgID        = 33   // GLOBAL_POSITION_INT
	compID byte  = 1
)

// Extra CRC for GLOBAL_POSITION_INT
const msg33ExtraCRC = 104

var DOMAIN = os.Getenv("DOMAIN")
var PORT = os.Getenv("PORT")

var ENDPOINT = DOMAIN + ":" + PORT

type InitialData struct {
	id int64
	lat float64
	lon float64
	drift float64
	direction float64
}

var mu sync.Mutex

func directionSignFloat(degrees int64) float64 {
    // Convert degrees to radians
    if degrees >= 0 && degrees <= 180 {
        return 1
    }
    return -1
}

func droneHandler(data InitialData, conn *net.Conn, stopChan <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done()
	lat, lon, drift, direction := data.lat, data.lon, data.drift, data.direction
	sysID := data.id

	for {
		select {
        case <-stopChan:
            log.Printf("Worker %d: shutting down cleanly\n", sysID)
            return nil
        default:
			packet := buildGlobalPositionInt(
				sysID,
				lat * 1e7,   // latitude * 1E7
				lon * 1e7,   // longitude * 1E7
				20*1000,          // altitude mm
				0,                // relative alt mm
				0, 0, 0,          // vx, vy, vz (cm/s)
				9000,             // heading (centidegrees)
			)

			// log.Printf("Packet [%d]: %v ", sysID, packet)

			mu.Lock()
			_, err := (*conn).Write(packet)
			if err != nil {
				log.Println("write error:", err)
				return err
			}
			mu.Unlock()

			time.Sleep(1000 * time.Millisecond)

			// Small delta: +/- 0.001 degrees (~111 meters)
			lat += (rand.Float64() + direction) * drift
			lon += (rand.Float64() + direction) * drift
		}
	}
}

func main() {
	stopChan := make(chan struct{})
	conn, err := net.Dial("udp", ENDPOINT)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Define drone devices and initial data
	// Pass to goroutine to stream simulated drone data

	initialDataList := []InitialData{
		// London (3)
		{id: 1, lat: 51.5084, lon: -0.1278, drift: 0.0015, direction: directionSignFloat(90)},
		{id: 2, lat: 51.5100, lon: -0.1407, drift: 0.0015, direction: directionSignFloat(180)},
		{id: 3, lat: 51.5033, lon: -0.1195, drift: 0.0015, direction: directionSignFloat(270)},

		// Africa (1)
		{id: 4, lat: 6.5244, lon: 3.3792, drift: 0.0015, direction: directionSignFloat(135)},   // Lagos, Nigeria

		// Americas (2)
		{id: 5, lat: 40.7128, lon: -74.0060, drift: 0.0015, direction: directionSignFloat(0)},  // New York
		{id: 6, lat: 34.0522, lon: -118.2437, drift: 0.0015, direction: directionSignFloat(180)}, // Los Angeles, USA
	}

	var wg sync.WaitGroup

	for _, v := range initialDataList {
		wg.Add(1)
		go droneHandler(v, &conn, stopChan, &wg)
	}

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Wait for signal
    <-sigChan
    log.Println("Main: received shutdown signal")

    close(stopChan)

    // Wait for workers
    wg.Wait()
    log.Println("Main: all workers stopped, exiting")
}

func buildGlobalPositionInt(sysID int64, lat, lon, alt, relAlt float64, vx, vy, vz, hdg int32) []byte {
	payload := make([]byte, 28)
	
	// log.Println("Latitude: ", int32(lat))

	// Payload format (GLOBAL_POSITION_INT msg 33)
	putInt32(payload, 0, int32(time.Now().UnixMilli()))
	putInt32(payload, 4, int32(lat))
	putInt32(payload, 8, int32(lon))
	putInt32(payload, 12, int32(alt))
	putInt32(payload, 16, int32(relAlt))
	putInt16(payload, 20, int16(vx))
	putInt16(payload, 22, int16(vy))
	putInt16(payload, 24, int16(vz))
	putUint16(payload, 26, uint16(hdg))

	frame := buildMAVLink2Frame(sysID, payload, msgID)
	return frame
}

func buildMAVLink2Frame(sysID int64, payload []byte, msgID int) []byte {
	header := []byte{
		stx,
		byte(len(payload)),
		0x00, // incompat flags
		0x00, // compat flags
		0x01, // sequence
		byte(sysID),
		compID,
		byte(msgID), byte(msgID >> 8), byte(msgID >> 16),
	}

	frame := append(header, payload...)
	checksum := crcMAV(frame[1:])         // exclude STX
	checksum = crcAccumulateExtra(checksum) // add extra CRC

	frame = append(frame, byte(checksum), byte(checksum>>8))
	return frame
}

// --------------------------------
// Helpers
// --------------------------------

func putInt32(b []byte, o int, v int32) {
	b[o] = byte(v)
	b[o+1] = byte(v >> 8)
	b[o+2] = byte(v >> 16)
	b[o+3] = byte(v >> 24)
}

func putInt16(b []byte, o int, v int16) {
	b[o] = byte(v)
	b[o+1] = byte(v >> 8)
}

func putUint16(b []byte, o int, v uint16) {
	b[o] = byte(v)
	b[o+1] = byte(v >> 8)
}

// MAVLink CRC
func crcMAV(buf []byte) uint16 {
	var crc uint16
	for _, b := range buf {
		tmp := b ^ byte(crc&0xFF)
		tmp ^= tmp << 4
		crc = (crc >> 8) ^ (uint16(tmp)<<8) ^ (uint16(tmp)<<3) ^ (uint16(tmp)>>4)
	}
	return crc
}

func crcAccumulateExtra(crc uint16) uint16 {
	extra := msg33ExtraCRC
	tmp := byte(extra) ^ byte(crc&0xFF)
	tmp ^= tmp << 4
	return (crc >> 8) ^ (uint16(tmp)<<8) ^ (uint16(tmp)<<3) ^ (uint16(tmp)>>4)
}


