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
    gameConfig: Record<string, any>;
    setRoomInfo: (room: Room | null) => void;
    setGameConfig: (config: Record<string, any>) => void;
    sendMessage: (data: any) => void;
}

export const WebSocketContext = createContext<WebSocketContextType | null>(null);
