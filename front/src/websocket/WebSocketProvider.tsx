import React, { useEffect, useState, useRef, useCallback } from "react";
import type { ReactNode } from "react";
import { WebSocketContext } from "./WebSocketContext";
import type { WebSocketStatus, Room, Player } from "./WebSocketContext";
import { useAuth } from "@/auth/AuthContext";

interface WebSocketProviderProps {
    children: ReactNode;
}

const RECONNECT_INTERVAL = 3000;
const ROOM_STORAGE_KEY = "minigames_room";

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
    const [status, setStatus] = useState<WebSocketStatus>("disconnected");
    const [roomInfo, setRoomInfoState] = useState<Room | null>(() => {
        const saved = localStorage.getItem(ROOM_STORAGE_KEY);
        return saved ? JSON.parse(saved) : null;
    });
    const [players, setPlayers] = useState<Player[]>([]);
    const [selectedGame, setSelectedGame] = useState<string>("Wordle");
    const { accessToken, isLoading } = useAuth();
    const socketRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<number | null>(null);

    const setRoomInfo = useCallback((room: Room | null) => {
        setRoomInfoState(room);
        if (room) {
            localStorage.setItem(ROOM_STORAGE_KEY, JSON.stringify(room));
        } else {
            localStorage.removeItem(ROOM_STORAGE_KEY);
        }
    }, []);

    const connect = useCallback(() => {
        if (isLoading) return;

        // Reset if no room info
        if (!roomInfo) {
            if (socketRef.current) {
                socketRef.current.close();
                socketRef.current = null;
            }
            setStatus("disconnected");
            setPlayers([]);
            return;
        }

        if (!accessToken) {
            console.log("WebSocket: User not authenticated, skipping connection.");
            setStatus("disconnected");
            setPlayers([]);
            if (socketRef.current) {
                socketRef.current.close();
                socketRef.current = null;
            }
            return;
        }

        const wsUrl = import.meta.env.VITE_WS_URL;
        if (!wsUrl) {
            console.warn("VITE_WS_URL is not defined in .env");
            setStatus("error");
            return;
        }

        let finalUrl = wsUrl;
        try {
            const url = new URL(wsUrl);
            url.searchParams.set("token", accessToken);
            url.searchParams.set("room", roomInfo.code);
            finalUrl = url.toString();
        } catch {
            finalUrl += (wsUrl.includes("?") ? "&" : "?") + `token=${accessToken}&room=${roomInfo.code}`;
        }

        console.log(`Connecting to WebSocket: ${finalUrl.replace(accessToken, "***")}`);
        setStatus("connecting");

        const socket = new WebSocket(finalUrl);
        socketRef.current = socket;

        socket.onopen = () => {
            console.log("WebSocket connected to room:", roomInfo.code);
            setStatus("connected");
        };

        socket.onclose = () => {
            console.log("WebSocket disconnected");
            setStatus("disconnected");
            setPlayers([]);

            if (socketRef.current === socket) {
                socketRef.current = null;
                // Only reconnect if we still have a room and a token
                if (accessToken && roomInfo) {
                    reconnectTimeoutRef.current = window.setTimeout(connect, RECONNECT_INTERVAL);
                }
            }
        };

        socket.onerror = (error) => {
            console.error("WebSocket error:", error);
            setStatus("error");
        };

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === "PLAYER_LIST") {
                    setPlayers(data.payload.players);
                } else if (data.type === "GAME_SELECTED") {
                    setSelectedGame(data.payload.gameId);
                }
            } catch (err) {
                console.error("Failed to parse WS message", err);
            }
        };
    }, [accessToken, isLoading, roomInfo]);

    useEffect(() => {
        connect();

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (socketRef.current) {
                socketRef.current.close();
            }
        };
    }, [connect]);

    const sendMessage = useCallback((data: unknown) => {
        if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify(data));
        } else {
            console.warn("WebSocket is not connected. Message not sent:", data);
        }
    }, []);

    return (
        <WebSocketContext.Provider value={{ socket: socketRef.current, status, roomInfo, players, selectedGame, setRoomInfo, sendMessage }}>
            {children}
        </WebSocketContext.Provider>
    );
};
