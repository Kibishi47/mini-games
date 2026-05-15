import React, { useEffect, useState, useRef, useCallback } from "react";
import type { ReactNode } from "react";
import { WebSocketContext } from "./WebSocketContext";
import type { WebSocketStatus, Room, Player } from "./WebSocketContext";
import { useAuth } from "@/auth/AuthContext";
import { http } from "@/lib/api/http";

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
    const [gameConfig, setGameConfig] = useState<Record<string, any>>({});
    const [session, setSession] = useState<GameSession | null>(null);
    const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
    const { accessToken, isLoading } = useAuth();
    const socketRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<number | null>(null);
    const isKickedRef = useRef<boolean>(false);

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
        if (isKickedRef.current) return;

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
        let wasConnected = false;

        socket.onopen = () => {
            console.log("WebSocket connected to room:", roomInfo.code);
            setStatus("connected");
            wasConnected = true;
        };

        socket.onclose = async (event) => {
            console.log("WebSocket disconnected", event.code, event.reason);
            setStatus("disconnected");
            setPlayers([]);

            if (socketRef.current === socket) {
                socketRef.current = null;
                
                // CIRCUIT BREAKER: If it never opened, it's likely a 403 or auth error
                if (!wasConnected) {
                    console.warn("WebSocket handshake failed. Stopping reconnection.");
                    isKickedRef.current = true;
                    setRoomInfo(null);
                    alert("Votre session a expiré ou l'accès au salon est refusé.");
                    window.location.href = "/";
                    return;
                }

                // If we still have a roomInfo, check if it's still valid before retrying
                if (roomInfo) {
                    try {
                        await http(`/rooms/${roomInfo.code}`);
                        // If we reach here, room exists, we can retry
                        if (accessToken && roomInfo) {
                            reconnectTimeoutRef.current = window.setTimeout(connect, RECONNECT_INTERVAL);
                        }
                        } catch (err: any) {
                        if (err.status === 404 || err.status === 403 || err.status === 401) {
                            console.warn("Room access lost or room no longer exists, clearing state.", err.status);
                            isKickedRef.current = true;
                            setRoomInfo(null);
                            if (err.status === 403) {
                                alert("Votre session a expiré ou vous avez été retiré du salon.");
                                window.location.href = "/";
                            }
                        } else {
                            // Other error, maybe server is down, still retry
                            reconnectTimeoutRef.current = window.setTimeout(connect, RECONNECT_INTERVAL);
                        }
                    }
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
                } else if (data.type === "CONFIG_UPDATED") {
                    setGameConfig(data.payload);
                } else if (data.type === "ROOM_UPDATED") {
                    setRoomInfo(data.payload.room);
                } else if (data.type === "ROOM_CLOSED") {
                    console.warn("Room was closed by server:", data.payload.reason);
                    setRoomInfo(null);
                } else if (data.type === "GAME_STARTED") {
                    setSession(data.payload.session);
                    setRoomInfo(data.payload.room);
                } else if (data.type === "CHAT_MESSAGE") {
                    setChatMessages(prev => [...prev, data.payload]);
                } else if (data.type === "CHAT_HISTORY") {
                    setChatMessages(data.payload);
                } else if (data.type === "GAME_STOPPED") {
                    setSession(null);
                    setRoomInfo(data.payload.room);
                } else if (data.type === "KICKED") {
                    console.warn("You were kicked from the room:", data.payload.reason);
                    isKickedRef.current = true;
                    setRoomInfo(null);
                    alert(data.payload.reason);
                    window.location.href = "/";
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
        <WebSocketContext.Provider value={{ socket: socketRef.current, status, roomInfo, players, selectedGame, gameConfig, session, chatMessages, setRoomInfo, setGameConfig, sendMessage }}>
            {children}
        </WebSocketContext.Provider>
    );
};
