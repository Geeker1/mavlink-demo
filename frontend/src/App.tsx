import './App.css'

import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'
import { useEffect, useReducer, useState, useMemo } from 'react'
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"

import type { State, Action } from "@/utils/types"

import L from "leaflet";

// Custom icon
const droneIcon = L.icon({
  iconUrl: "/drone.png", // path to your custom image
  iconSize: [100, 100],         // size of the icon [width, height]
  iconAnchor: [20, 20],       // point of the icon which will correspond to marker's location
  popupAnchor: [0, -20],      // point from which the popup should open relative to the iconAnchor
});

const WS_URL = "ws://localhost:1323/ws"

const initialState: State = {
  current: {},
  history: {},
  hidden: new Set()
};

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "ADD_POSITION": {
      const { sysid } = action.payload;

      // Update history
      const newHistory = { ...state.history };
      if (!newHistory[sysid]) {
        newHistory[sysid] = [];
      }
      newHistory[sysid].push(action.payload);

      // Update current
      const newCurrent = { ...state.current, [sysid]: action.payload };

      return {
        current: newCurrent,
        history: newHistory,
        hidden: state.hidden
      };
    }

    case "HIDE_DEVICE": {
      state.hidden.add(action.payload);
      return state
    }

    case "SHOW_DEVICE": {
      state.hidden.delete(action.payload);
      return state;
    }

    case "RESET":
      return initialState;

    default:
      return state;
  }
}

function App() {
  const [state, dispatch] = useReducer(reducer, initialState);
  const [map, setMap] = useState<any>(null);

  const hideDevice = (id: number) => {
    dispatch({ type: "HIDE_DEVICE", payload: id });
  }

  const showDevice = (id: number) => {
    dispatch({ type: "SHOW_DEVICE", payload: id });
  }

  useEffect(()=> {
    const ws = new WebSocket(WS_URL)

    ws.onopen = function() {
    }

    ws.onmessage = function(evt) {
      dispatch({ type: "ADD_POSITION", payload: JSON.parse(evt.data) });
    }

    return () => {
      if (ws && ws.readyState !== WebSocket.CLOSED) {
        ws.close();
      }
    }
  }, [])

  const displayMap = useMemo(
    () => (
      <MapContainer center={[51.505, -0.09]} zoom={13} scrollWheelZoom={false} ref={setMap}>
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {Object.entries(state.current).filter(([sysid, _])=>{
          return !state.hidden.has(Number(sysid));
        }).map(([sysid, pos]) => (
          <Marker key={sysid} position={[pos.lat, pos.lon]} icon={droneIcon}>
            <Popup>
              Drone #{sysid}<br />
              {pos.lat.toFixed(6)}, {pos.lon.toFixed(6)}
            </Popup>
          </Marker>
        ))}
      </MapContainer>
    ),
    [state],
  )

  return (
    <SidebarProvider>

    <AppSidebar state={state} map={map} hideDevice={hideDevice} showDevice={showDevice} />
    <main className='flex-1'>
      <div className='flex py-4'>
        <SidebarTrigger />
        <h1 className='mx-6 font-bold text-lg text-gray-800'>Proxylity Drone Simulation</h1>
      </div>
      {displayMap}
    </main>
    </SidebarProvider>
  )
}

export default App
