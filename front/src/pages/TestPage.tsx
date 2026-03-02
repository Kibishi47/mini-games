import React, { useState, useEffect, useCallback, useRef } from "react";
import { useWebSocket } from "@/websocket/useWebSocket";

interface MessageEntry {
  id: string;
  data: any;
  timestamp: Date;
  type: "sent" | "received";
}

const TestPage: React.FC = () => {
  const { status, sendMessage, socket } = useWebSocket();
  const [inputMessage, setInputMessage] = useState("");
  const [entries, setEntries] = useState<MessageEntry[]>([]);
  const scrollRef = useRef<HTMLDivElement>(null);

  const addEntry = useCallback((data: any, type: "sent" | "received") => {
    setEntries((prev) => [
      ...prev,
      {
        id: Math.random().toString(36).substring(7),
        data,
        timestamp: new Date(),
        type,
      },
    ]);
  }, []);

  useEffect(() => {
    if (!socket) return;

    const handleMessage = (event: MessageEvent) => {
      let data = event.data;
      try {
        data = JSON.parse(event.data);
      } catch {
        // Not JSON, keep as string
      }
      addEntry(data, "received");
    };

    socket.addEventListener("message", handleMessage);
    return () => {
      socket.removeEventListener("message", handleMessage);
    };
  }, [socket, addEntry]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [entries]);

  const handleSend = useCallback(() => {
    if (inputMessage.trim()) {
      const payload = { type: "test", content: inputMessage, timestamp: new Date().toISOString() };
      sendMessage(payload);
      addEntry(payload, "sent");
      setInputMessage("");
    }
  }, [inputMessage, sendMessage, addEntry]);

  return (
    <div style={{
      padding: "2rem",
      maxWidth: "800px",
      margin: "0 auto",
      height: "100vh",
      display: "flex",
      flexDirection: "column",
      fontFamily: "'Inter', system-ui, sans-serif"
    }}>
      <header style={{ marginBottom: "1.5rem" }}>
        <h1 style={{ margin: 0, fontSize: "1.5rem" }}>WebSocket Debugger</h1>
        <div style={{ display: "flex", alignItems: "center", gap: "0.5rem", marginTop: "0.5rem" }}>
          <div style={{
            width: "10px",
            height: "10px",
            borderRadius: "50%",
            background: status === "connected" ? "#10b981" : status === "error" ? "#ef4444" : "#f59e0b"
          }} />
          <span style={{ fontSize: "0.875rem", color: "#666", fontWeight: 500 }}>
            {status.toUpperCase()}
          </span>
        </div>
      </header>

      <div
        ref={scrollRef}
        style={{
          flex: 1,
          border: "1px solid #e5e7eb",
          borderRadius: "8px",
          padding: "1rem",
          overflowY: "auto",
          background: "#f9fafb",
          marginBottom: "1rem",
          display: "flex",
          flexDirection: "column",
          gap: "0.75rem"
        }}
      >
        {entries.length === 0 ? (
          <div style={{ color: "#9ca3af", textAlign: "center", marginTop: "2rem" }}>
            No messages yet. Send one to start testing.
          </div>
        ) : (
          entries.map((entry) => (
            <div
              key={entry.id}
              style={{
                alignSelf: entry.type === "sent" ? "flex-end" : "flex-start",
                maxWidth: "80%",
                background: entry.type === "sent" ? "#3b82f6" : "#ffffff",
                color: entry.type === "sent" ? "#ffffff" : "#111827",
                padding: "0.75rem 1rem",
                borderRadius: "12px",
                boxShadow: "0 1px 2px rgba(0,0,0,0.05)",
                border: entry.type === "received" ? "1px solid #e5e7eb" : "none"
              }}
            >
              <div style={{ fontSize: "0.75rem", opacity: 0.7, marginBottom: "0.25rem", display: "flex", justifyContent: "space-between", gap: "1rem" }}>
                <span>{entry.type === "sent" ? "You" : "Server"}</span>
                <span>{entry.timestamp.toLocaleTimeString()}</span>
              </div>
              <pre style={{
                margin: 0,
                whiteSpace: "pre-wrap",
                wordBreak: "break-all",
                fontSize: "0.875rem",
                fontFamily: "monospace"
              }}>
                {typeof entry.data === "object" ? JSON.stringify(entry.data, null, 2) : entry.data}
              </pre>
            </div>
          ))
        )}
      </div>

      <div style={{ display: "flex", gap: "0.5rem", background: "#fff", padding: "1rem", border: "1px solid #e5e7eb", borderRadius: "8px" }}>
        <input
          type="text"
          value={inputMessage}
          onChange={(e) => setInputMessage(e.target.value)}
          placeholder="Type a message..."
          style={{
            flex: 1,
            padding: "0.625rem",
            border: "1px solid #d1d5db",
            borderRadius: "6px",
            outline: "none"
          }}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
        />
        <button
          onClick={handleSend}
          disabled={status !== "connected"}
          style={{
            padding: "0.625rem 1.25rem",
            background: status === "connected" ? "#3b82f6" : "#9ca3af",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: status === "connected" ? "pointer" : "not-allowed",
            fontWeight: 600
          }}
        >
          Send
        </button>
      </div>
    </div>
  );
};

export default TestPage;
