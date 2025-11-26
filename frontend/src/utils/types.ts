

export type MAVLinkData = {
  sysid: number;
  timestamp: number;
  lat: number;
  lon: number;
};

export type State = {
  current: Record<number, MAVLinkData>; // sysid -> latest position
  history: Record<number, MAVLinkData[]>; // sysid -> array of movements
  hidden: Set<number>
};

export type Action =
  | { type: "ADD_POSITION"; payload: MAVLinkData }
  | { type: "RESET" } | { type: "HIDE_DEVICE", payload: number } | { type: "SHOW_DEVICE", payload: number };
