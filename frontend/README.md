# MAVLINK Telemetry Frontend (Leaflet + WebSocket)

This frontend visualizes real-time drone locations streamed from a WebSocket backend.
Each drone is rendered as a custom icon on an interactive Leaflet map, with the ability to hide/show devices and inspect position history.


## Core Features

### Live telemetry

* Connects to `VITE_WEBSOCKET_URL` (e.g. `ws://localhost:1323/ws`)
* Receives `{ sysid, lat, lon, timestamp }`
* Updates drone state in real time

### Map visualization

* Uses **react-leaflet** to render a map.
* Each drone is represented by a custom icon (`drone.png`).
* Clicking a marker shows popup coords.

### Device visibility controls

The UI includes:

* `HIDE_DEVICE(sysid)`
* `SHOW_DEVICE(sysid)`
* History is preserved even if hidden.

### Persistent history

The reducer stores:

* `current`: latest position per drone
* `history`: full track per drone
* `hidden`: masked ids so the map only renders visible drones


## State Model

```ts
State {
  current: {
    [sysid]: { lat, lon, timestamp }
  }
  history: {
    [sysid]: Array<...>
  }
  hidden: Set<number>
}
```

Actions:

* `ADD_POSITION`: append to history and update current
* `HIDE_DEVICE`
* `SHOW_DEVICE`
* `RESET`

## WebSocket Flow

1. App bootstraps → opens WS connection.
2. On every message:
   * JSON parsed
   * Payload dispatched to reducer
3. Cleanup on unmount:
   * WS closed if still open


## Map Logic

* Default center: London (51.505, -0.09)
* Only devices not in `hidden` set are rendered.

Example marker rendering:

```tsx
<Marker position={[lat, lon]} icon={droneIcon}>
  <Popup>Drone #ID: lat, lon</Popup>
</Marker>
```


## Environment Variables

`.env` (Vite style):

```
VITE_WEBSOCKET_URL=ws://localhost:1323/ws
```


## How to Run

### 1. Install deps

```bash
npm install
```

or

```bash
yarn
```

### 2. Create `.env`

```bash
echo "VITE_WEBSOCKET_URL=ws://localhost:1323/ws" > .env
```

### 3. Place drone icon

`public/drone.png` — any PNG asset.

### 4. Launch

```bash
npm run dev
```


## Integration Expectations

* The backend must send MAVLink-derived JSON packets shaped like:

```json
{
  "sysid": 2,
  "lat": 51.509882,
  "lon": -0.122800,
  "timestamp": 1708892300
}
```

* Lat/Lon MUST be decimal degrees.
* The map only updates when a new message arrives.


## Notes & Caveats

* Initial map zoom is static.
* History is stored in-memory; refresh wipes it.
* Hidden devices still update silently in state — no data loss.
