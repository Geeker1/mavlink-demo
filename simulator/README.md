## README

### MAVLink UDP Drone Simulator

This program simulates multiple drones and streams **GLOBAL_POSITION_INT (MAVLink Message 33)** packets over **UDP** to a specified endpoint.

It is used to emulate real-time UAV telemetry without needing actual hardware.

To understand the serialization mechanism:

[MAVLINK Serialization](https://mavlink.io/en/guide/serialization.html)


## What it does

1. Loads the UDP endpoint from environment:

   * `DOMAIN`
   * `PORT`
2. Dials the UDP socket: `DOMAIN:PORT`
3. Spawns a worker goroutine for each simulated drone.
4. Each goroutine:

   * Generates a MAVLink2 packet containing GPS + attitude data.
   * Sends the packet every 1 second.
   * Randomly moves latitude/longitude to simulate movement.
5. When SIGINT/SIGTERM arrives:

   * Workers stop gracefully.
   * Program exits cleanly.


## Environment variables

| Name     | Description                         |
| -------- | ----------------------------------- |
| `DOMAIN` | Target IP/domain (e.g. `127.0.0.1`) |
| `PORT`   | Target UDP port (e.g. `14550`)      |

Example:

```
DOMAIN=127.0.0.1
PORT=14550
```


## MAVLink message details

This simulator packs MAVLink 2 **GLOBAL_POSITION_INT (MsgID 33)** manually.

Payload includes:

| Field        | Type   | Meaning                  |
| ------------ | ------ | ------------------------ |
| time_boot_ms | int32  | timestamp (ms)           |
| lat          | int32  | latitude in 1E7 degrees  |
| lon          | int32  | longitude in 1E7 degrees |
| alt          | int32  | altitude in mm           |
| relative_alt | int32  | relative altitude in mm  |
| vx,vy,vz     | int16  | speed in cm/s            |
| heading      | uint16 | centidegrees             |

* Uses STX `0xFD` (MAVLink2 header)
* Applies extra CRC value `104` for Message 33
* Generates a valid MAVLink2 frame (header + payload + checksum)


## Drone movement simulation

Each drone has:

* Starting latitude/longitude
* Drift magnitude: how much it moves each tick
* Direction sign (+1 or -1 based on quadrant)
* Independent goroutine writing packets to UDP

Values change slightly every second to create route-like movement.


## Included drones

Hardcoded initial positions:

* 3 devices around London
* 1 in Lagos (Nigeria)
* 2 in the US: New York & Los Angeles

Each runs as separate worker transmitting real-time coordinates.


## Run

```bash
go mod tidy
DOMAIN=127.0.0.1 PORT=14550 go run .
```

Recommended receiver: MAVLink ground station, telemetry listener, or custom decoder.


## Shutdown behavior

* Program listens to signals (`SIGINT`, `SIGTERM`)
* On signal:

  * Close control channel
  * Workers exit
  * UDP connection closes
  * Program terminates cleanly


## Use cases

* Test MAVLink parsing
* Simulate real UAV fleets
* Validate WebSocket / SQS ingestion pipelines
* Debug telemetry visualization apps
