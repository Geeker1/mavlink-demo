import { useEffect, useState } from "react"
import {
  Eye,
  EyeOff,
} from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

import type { MAVLinkData, State } from "@/utils/types"

type ToggleState = {
  visible: boolean
  enabled: boolean
}

export function AppSidebar({ state, map, hideDevice, showDevice }: { state: State, map: L.Map, hideDevice: (id: number) => void, showDevice: (id: number) => void }) {
  // local toggles for each vehicle (sysid)
  const [toggles, setToggles] = useState<Record<number, ToggleState>>({})

  // initialize toggles when state.current changes
  useEffect(() => {
    const next: Record<number, ToggleState> = {}
    Object.keys(state.current).forEach((k) => {
      const id = Number(k)
      next[id] = {
        visible: toggles[id]?.visible ?? true,
        enabled: toggles[id]?.enabled ?? true,
      }
    })
    setToggles((prev) => ({ ...next, ...prev }))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [Object.keys(state.current).length])

  const toggleVisible = (sysid: number, prevVisible: boolean) => {
    setToggles((t) => ({ ...t, [sysid]: { ...(t[sysid] ?? { visible: true, enabled: true }), visible: !(t[sysid]?.visible ?? true) } }))
    if(prevVisible) {
      hideDevice(sysid);
    } else {
      showDevice(sysid);
    }
  }

  // derived metrics
  const droneIds = Object.keys(state.current).map((k) => Number(k))
  const totalDrones = droneIds.length
  const totalPositions = Object.values(state.history).reduce((acc, arr) => acc + (arr?.length ?? 0), 0)

  const handleDevice = (data: MAVLinkData) => {
    // alert("Clicked device");
    map.setView([data.lat, data.lon]);
  }

  return (
    <Sidebar>
      <SidebarContent>
        {/* Metrics */}
        <SidebarGroup>
          <SidebarGroupLabel className="text-md px-0 mb-4">Live Metrics</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem key="summary">
                <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
                  <div>Drones: {totalDrones}</div>
                  <div>Events: {totalPositions}</div>
                </div>
              </SidebarMenuItem>

              <hr/>

              {droneIds.map((id) => {
                const data = state.current[id]
                const ts = data ? new Date(data.timestamp*1000).toLocaleTimeString() : "-"
                const toggle = toggles[id] ?? { visible: true, enabled: true }

                return (
                  <SidebarMenuItem className="pt-1" key={`drone-${id}`}>
                    <div className="relative z-10 cursor-pointer rounded-xl shadow-lg hover:shadow-2xl transition-shadow p-3 pt-4 bg-white" style={{ display: "flex", alignItems: "center", justifyContent: "space-between"}}>
                      <div onClick={() => {handleDevice(data)}} style={{ display: "flex", gap: 8, alignItems: "center" }}>
                        <p className="text-sm cursor-pointer" style={{ fontWeight: 600, fontSize: 12 }}>Drone #{id}</p>
                        {data ? (
                          <div style={{ fontSize: 12}}>
                            {data.lat.toFixed(4)}, {data.lon.toFixed(4)} @ {ts}
                          </div>
                        ) : (
                          <div style={{ fontSize: 12, color: "var(--muted)" }}>no data</div>
                        )}
                      </div>

                      <div style={{ display: "flex", gap: 8 }}>
                        <button
                          aria-label={`toggle-visibility-${id}`}
                          onClick={() => toggleVisible(id, toggle.visible)}
                          title={toggle.visible ? "Hide" : "Show"}
                          style={{ background: "transparent", border: "none", cursor: "pointer" }}
                        >
                          {toggle.visible ? <Eye size={16} /> : <EyeOff size={16} />}
                        </button>
                      </div>
                    </div>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}