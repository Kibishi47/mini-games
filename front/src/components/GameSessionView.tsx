import React, { useState, useRef, useEffect } from "react";
import { useWebSocket } from "../websocket/useWebSocket";
import { useAuth } from "../auth/AuthContext";
import Button from "./common/Button";
import "@/styles/pages/play.css";

const GameSessionView: React.FC = () => {
    const ws = useWebSocket();
    const { user } = useAuth();
    const { roomInfo, session, chatMessages, sendMessage } = ws || {};
    const [chatInput, setChatInput] = useState("");
    const chatEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [chatMessages]);

    const handleSendChat = (e: React.FormEvent) => {
        e.preventDefault();
        if (!chatInput.trim() || !sendMessage) return;
        
        sendMessage({
            type: "CHAT_MESSAGE",
            payload: {
                text: chatInput,
                time: new Date().toISOString()
            }
        });
        setChatInput("");
    };

    if (!session || !roomInfo) return <div>Chargement de la session...</div>;

    return (
        <div className="container" style={{ paddingTop: "80px", maxWidth: "1200px" }}>
            <div className="lobby-header" style={{ marginBottom: "1rem" }}>
                <div className="lobby-code-badge">
                    <span className="label">JEU EN COURS</span>
                    <span className="code">{session.game}</span>
                </div>
                <div className="connection-status">
                    <span className="status-dot connected"></span>
                    <span>En direct</span>
                </div>
                {roomInfo.hostId === user?.id && (
                    <Button 
                        variant="danger" 
                        size="sm" 
                        style={{ marginLeft: "1rem" }}
                        onClick={() => sendMessage?.({ type: "STOP_GAME", payload: {} })}
                    >
                        Stopper la partie
                    </Button>
                )}
            </div>

            <div className="lobby-grid" style={{ gridTemplateColumns: "1fr 350px" }}>
                {/* Espace principal du jeu */}
                <div className="lobby-section" style={{ minHeight: "500px", display: "flex", alignItems: "center", justifyContent: "center" }}>
                    <div style={{ textAlign: "center", opacity: 0.7 }}>
                        <h2 style={{ fontSize: "2rem", marginBottom: "1rem" }}>{session.game}</h2>
                        <p>Le jeu va bientôt démarrer ici...</p>
                        <p style={{ fontSize: "0.85rem", marginTop: "2rem" }}>Config: {JSON.stringify(session.config)}</p>
                    </div>
                </div>

                {/* Chat en temps réel */}
                <div className="lobby-section players-section" style={{ display: "flex", flexDirection: "column", padding: "1rem" }}>
                    <h2 className="section-title" style={{ fontSize: "1.1rem", marginBottom: "1rem", borderBottom: "1px solid rgba(255,255,255,0.1)", paddingBottom: "0.5rem" }}>
                        Chat du salon
                    </h2>
                    
                    <div style={{ flex: 1, overflowY: "auto", marginBottom: "1rem", display: "flex", flexDirection: "column", gap: "0.5rem" }}>
                        {chatMessages?.map((msg, idx) => {
                            const isMe = msg.userId === user?.id;
                            if (msg.system) {
                                return (
                                    <div key={idx} style={{ textAlign: "center", fontSize: "0.85rem", color: "var(--color-primary)", padding: "0.25rem 0" }}>
                                        {msg.text}
                                    </div>
                                );
                            }
                            return (
                                <div key={idx} style={{ 
                                    background: isMe ? "rgba(79, 172, 254, 0.2)" : "rgba(255,255,255,0.05)",
                                    padding: "0.5rem 0.75rem", 
                                    borderRadius: "8px",
                                    alignSelf: isMe ? "flex-end" : "flex-start",
                                    maxWidth: "85%",
                                    border: isMe ? "1px solid rgba(79, 172, 254, 0.3)" : "1px solid transparent"
                                }}>
                                    <div style={{ fontSize: "0.75rem", opacity: 0.6, marginBottom: "0.2rem" }}>{msg.username}</div>
                                    <div style={{ fontSize: "0.95rem" }}>{msg.text}</div>
                                </div>
                            );
                        })}
                        <div ref={chatEndRef} />
                    </div>

                    <form onSubmit={handleSendChat} style={{ display: "flex", gap: "0.5rem" }}>
                        <input 
                            type="text" 
                            value={chatInput}
                            onChange={(e) => setChatInput(e.target.value)}
                            placeholder="Écrivez un message..."
                            style={{ 
                                flex: 1, 
                                background: "rgba(0,0,0,0.3)", 
                                border: "1px solid rgba(255,255,255,0.1)", 
                                color: "white", 
                                padding: "0.75rem", 
                                borderRadius: "8px",
                                outline: "none"
                            }}
                        />
                        <Button type="submit" variant="primary" style={{ padding: "0 1.25rem" }}>Envoyer</Button>
                    </form>
                </div>
            </div>
        </div>
    );
};

export default GameSessionView;
