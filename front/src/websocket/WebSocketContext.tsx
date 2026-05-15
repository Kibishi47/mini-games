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

export interface GameSession {
    id: string;
    roomId: string;
    game: string;
    config: Record<string, any>;
    startedAt: string;
    endedAt?: string;
}

export interface ChatMessage {
    userId: string;
    username: string;
    text: string;
    system?: boolean;
    time: string;
}

export interface WebSocketContextType {
    socket: WebSocket | null;
    status: WebSocketStatus;
    roomInfo: Room | null;
    players: Player[];
    selectedGame: string;
    gameConfig: Record<string, any>;
    session: GameSession | null;
    chatMessages: ChatMessage[];
    setRoomInfo: (room: Room | null) => void;
    setGameConfig: (config: Record<string, any>) => void;
    sendMessage: (data: any) => void;
}

export const WebSocketContext = createContext<WebSocketContextType | null>(null);
