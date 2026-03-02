import { createContext } from "react";

export type WebSocketStatus = "connecting" | "connected" | "disconnected" | "error";

export interface WebSocketContextType {
    socket: WebSocket | null;
    status: WebSocketStatus;
    sendMessage: (data: unknown) => void;
}

export const WebSocketContext = createContext<WebSocketContextType | null>(null);
