import { createContext } from "react";

export type WebSocketStatus = "connecting" | "connected" | "disconnected" | "error";

export interface Room {
    id: string;
    code: string;
    status: string;
    hostId: string;
    maxPlayers: number;
}

export interface Player {
    username: string;
    isHost: boolean;
}

export interface WebSocketContextType {
    socket: WebSocket | null;
    status: WebSocketStatus;
    roomInfo: Room | null;
    players: Player[];
    selectedGame: string;
    setRoomInfo: (room: Room | null) => void;
    sendMessage: (data: unknown) => void;
}

export const WebSocketContext = createContext<WebSocketContextType | null>(null);
