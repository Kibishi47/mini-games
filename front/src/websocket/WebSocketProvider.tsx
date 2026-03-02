import React, { useEffect, useState, useRef, useCallback } from "react";
import type { ReactNode } from "react";
import { WebSocketContext } from "./WebSocketContext";
import type { WebSocketStatus } from "./WebSocketContext";
import { useAuth } from "@/auth/AuthContext";

interface WebSocketProviderProps {
    children: ReactNode;
}

const RECONNECT_INTERVAL = 3000;

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
    const [status, setStatus] = useState<WebSocketStatus>("disconnected");
    const { accessToken, isLoading } = useAuth();
    const socketRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<number | null>(null);

    const connect = useCallback(() => {
        if (isLoading) return;

        if (!accessToken) {
            console.log("WebSocket: User not authenticated, skipping connection.");
            setStatus("disconnected");
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
            finalUrl = url.toString();
        } catch {
            // Fallback for non-absolute URLs if any
            finalUrl += (wsUrl.includes("?") ? "&" : "?") + `token=${accessToken}`;
        }

        console.log(`Connecting to WebSocket: ${finalUrl.replace(accessToken, "***")}`);
        setStatus("connecting");

        const socket = new WebSocket(finalUrl);
        socketRef.current = socket;

        socket.onopen = () => {
            console.log("WebSocket connected");
            setStatus("connected");
        };

        socket.onclose = () => {
            console.log("WebSocket disconnected");
            setStatus("disconnected");

            // Only reconnect if this was the active socket and we still have a token
            if (socketRef.current === socket) {
                socketRef.current = null;
                if (accessToken) {
                    reconnectTimeoutRef.current = window.setTimeout(connect, RECONNECT_INTERVAL);
                }
            }
        };

        socket.onerror = (error) => {
            console.error("WebSocket error:", error);
            setStatus("error");
        };

        socket.onmessage = (_event) => {
            // Global message handling can go here if needed
        };
    }, [accessToken, isLoading]);

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
        <WebSocketContext.Provider value={{ socket: socketRef.current, status, sendMessage }}>
            {children}
        </WebSocketContext.Provider>
    );
};
