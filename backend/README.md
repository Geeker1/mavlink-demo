# MAVLink WebSocket Relay Server

This service listens to an AWS SQS queue for MAVLink telemetry messages and streams only **GLOBAL_POSITION_INT (Message ID = 33)** packets to connected WebSocket clients.

It is intended for real-time drone/vehicle location display.

---

## How it works

1. A WebSocket client connects to `GET /ws`.
2. The backend continuously polls an SQS queue.
3. Incoming messages are decoded from hex → raw MAVLink payload.
4. Messages that are not type **33** are ignored.
5. Valid packets are transformed into a JSON object:

```json
{
  "sysid": 1,
  "timestamp": 1732700000,
  "lat": 49.1234567,
  "lon": -121.9876543
}
```

6. Each JSON payload is pushed to the WebSocket client.

Messages are deleted in batch after processing.

---

## Environment variables

| Name                    | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| `QUEUE_URL`             | AWS SQS queue URL to read MAVLink messages from          |
| `ALLOWED_DOMAIN`        | WebSocket client origin allowed by CORS                  |
| `AWS_ACCESS_KEY_ID`     | AWS IAM access key ID for authentication                 |
| `AWS_SECRET_ACCESS_KEY` | AWS IAM secret access key for authentication             |
| `AWS_REGION`            | AWS region where SQS and other AWS services are deployed |

`.env` example:

```
QUEUE_URL=https://sqs.us-east-1.amazonaws.com/1234567890/mav_queue.fifo
ALLOWED_DOMAIN=http://localhost:5173
AWS_ACCESS_KEY_ID=xxxxxxxxx
AWS_SECRET_ACCESS_KEY=xxxxxxx
AWS_REGION=us-west-2
```


---

## WebSocket endpoint

```
ws://localhost:1323/ws
```

Clients should connect from the origin defined in `ALLOWED_DOMAIN`.

---

## AWS credentials

This uses **AWS SDK v2** default credential resolution:

1. Environment variables
2. `~/.aws/credentials` or mounted path in Docker
3. IAM role (EC2 metadata)

Make sure credentials exist in one of those places.

---

## Build & run

### Local

```bash
go mod tidy
go run .
```

### Docker (typical pattern)

```yaml
services:
  mav_backend:
    image: yourimage
    ports:
      - "1323:1323"
    environment:
      - QUEUE_URL=${QUEUE_URL}
      - ALLOWED_DOMAIN=${ALLOWED_DOMAIN}
    volumes:
      - ~/.aws:/root/.aws:ro
```

---

## Message handling rules

* Extracts payload length from byte position `1`.
* Verifies MAVLink message ID via byte indices `7:10`.
* Only processes packets where ID = `33`.
* Latitude & longitude are little-endian `int32` scaled by `1e7`.

---

## Use cases

* UAV/UGV position trackers
* Telemetry dashboards
* Field operations monitoring
* Demo integrations using MAVLink → WebSockets
