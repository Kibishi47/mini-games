import { useState } from "react";
import Button from "../components/common/Button";
import { http } from "../lib/api/http";
import { useWebSocket } from "../websocket/useWebSocket";
import { useAuth } from "../auth/AuthContext";
import type { Room } from "../websocket/WebSocketContext";
import "@/styles/pages/play.css";

const GAMES = [
    { id: "Wordle", name: "Wordle", icon: "📝", description: "Devinez le mot secret en 6 essais." },
    { id: "BetweenLines", name: "Between Lines", icon: "📖", description: "Lisez entre les lignes pour gagner." },
    { id: "Checkers", name: "Dames", icon: "🏁", description: "Le classique jeu de dames." },
    { id: "Chess", name: "Échecs", icon: "♟️", description: "Battez vos amis aux échecs." },
    { id: "FourInARow", name: "Puissance 4", icon: "🔴", description: "Alignez 4 jetons pour gagner." },
    { id: "Trivia", name: "Trivia", icon: "❓", description: "Testez votre culture générale." },
    { id: "Snake", name: "Snake", icon: "🐍", description: "Ne vous mordez pas la queue !" },
    { id: "Minesweeper", name: "Démineur", icon: "💣", description: "Évitez toutes les mines." },
];

const PlayPage = () => {
    const [activeTab, setActiveTab] = useState<"host" | "join">("host");
    const [joinCode, setJoinCode] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [showCopySuccess, setShowCopySuccess] = useState(false);
    
    const { user } = useAuth();
    const ws = useWebSocket();
    const { roomInfo, setRoomInfo, players, status, selectedGame, sendMessage } = ws || {};

    const isHost = roomInfo && user && roomInfo.hostId === user.id;

    const handleCopyCode = () => {
        if (!roomInfo) return;
        navigator.clipboard.writeText(roomInfo.code);
        setShowCopySuccess(true);
        setTimeout(() => setShowCopySuccess(false), 2000);
    };

    const handleSelectGame = (gameId: string) => {
        if (!isHost || !sendMessage) return;
        sendMessage({ type: "SELECT_GAME", payload: { gameId } });
    };

    const handleJoinCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 6);
        setJoinCode(value);
    };

    const handleCreateRoom = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const room = await http<Room>("/rooms", { method: "POST" });
            if (setRoomInfo) setRoomInfo(room);
        } catch (err: any) {
            setError(err.body?.error || "Erreur lors de la création du salon");
        } finally {
            setIsLoading(false);
        }
    };

    const handleJoinRoom = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const room = await http<Room>("/rooms/join", { 
                method: "POST",
                json: { code: joinCode }
            });
            if (setRoomInfo) setRoomInfo(room);
        } catch (err: any) {
            setError(err.body?.error || "Erreur lors de la connexion au salon");
        } finally {
            setIsLoading(false);
        }
    };

    const isJoinDisabled = joinCode.length !== 6 || isLoading;

    if (roomInfo) {
        return (
            <div className="container" style={{ paddingTop: "80px", maxWidth: "1200px" }}>
                <div className="lobby-header">
                    <div className="lobby-code-badge clickable" onClick={handleCopyCode}>
                        <span className="label">{showCopySuccess ? "COPIÉ !" : "CODE DU SALON"}</span>
                        <span className="code">{roomInfo.code}</span>
                    </div>
                    <div className="connection-status">
                        <div className={`status-dot ${status}`}></div>
                        <span>{status === "connected" ? "Connecté" : "Reconnexion..."}</span>
                    </div>
                </div>

                <div className="lobby-grid">
                    {/* Colonne Gauche: Configuration / Jeu */}
                    <div className="lobby-section game-config">
                        <h2 className="section-title">
                            {isHost ? "Configuration de la partie" : "Jeu sélectionné"}
                        </h2>
                        
                        {isHost ? (
                            <div className="games-list">
                                {GAMES.map(game => (
                                    <div 
                                        key={game.id} 
                                        className={`game-item ${selectedGame === game.id ? "selected" : ""}`}
                                        onClick={() => handleSelectGame(game.id)}
                                    >
                                        <span className="game-icon">{game.icon}</span>
                                        <div className="game-info">
                                            <span className="game-name">{game.name}</span>
                                            <span className="game-desc">{game.description}</span>
                                        </div>
                                        {selectedGame === game.id && <div className="selected-check">✓</div>}
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="selected-game-display">
                                <div className="game-hero">
                                    <span className="hero-icon">{GAMES.find(g => g.id === selectedGame)?.icon}</span>
                                    <h3>{GAMES.find(g => g.id === selectedGame)?.name}</h3>
                                    <p>{GAMES.find(g => g.id === selectedGame)?.description}</p>
                                </div>
                                <div className="waiting-host">
                                    <div className="loader-mini"></div>
                                    <span>En attente de l'hôte...</span>
                                </div>
                            </div>
                        )}
                    </div>

                    {/* Colonne Droite: Joueurs */}
                    <div className="lobby-section players-section">
                        <h2 className="section-title">Joueurs ({players?.length || 0})</h2>
                        <div className="players-list">
                            {players?.map((p, index) => (
                                <div key={p.username + index} className="player-card">
                                    <div className="player-avatar">
                                        {p.username.charAt(0).toUpperCase()}
                                    </div>
                                    <span className="player-name">
                                        {p.username}
                                        {p.isHost && <span className="host-crown" title="Hôte du salon"> 👑</span>}
                                    </span>
                                    {p.username === user?.username && <span className="you-badge">VOUS</span>}
                                </div>
                            ))}
                        </div>

                        <div className="lobby-actions">
                            {isHost ? (
                                <Button variant="primary" className="btn-full btn-lg pulse">
                                    Lancer la partie
                                </Button>
                            ) : (
                                <div className="ready-status">
                                    Préparez-vous, la partie va bientôt commencer !
                                </div>
                            )}
                            <Button 
                                variant="outline" 
                                className="btn-full" 
                                style={{ marginTop: "1rem" }}
                                onClick={() => setRoomInfo?.(null)}
                            >
                                Quitter le salon
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="container" style={{ paddingTop: "80px" }}>
            <div className="play-container">
                <h1 className="text-gradient" style={{ marginBottom: "2rem" }}>
                    Jouer
                </h1>

                <div className="play-card">
                    <div className="tabs">
                        <button
                            className={`tab ${activeTab === "host" ? "active" : ""}`}
                            onClick={() => setActiveTab("host")}
                        >
                            Créer une partie
                        </button>
                        <button
                            className={`tab ${activeTab === "join" ? "active" : ""}`}
                            onClick={() => setActiveTab("join")}
                        >
                            Rejoindre
                        </button>
                    </div>

                    <div className="tab-content">
                        {error && (
                            <div className="error-message">
                                {error}
                            </div>
                        )}

                        {activeTab === "host" ? (
                            <>
                                <p className="tab-description">
                                    Créez un salon et invitez vos amis à vous rejoindre avec un code.
                                </p>
                                <Button variant="primary" className="btn-full" onClick={handleCreateRoom} disabled={isLoading}>
                                    {isLoading ? "Création..." : "Créer un salon"}
                                </Button>
                            </>
                        ) : (
                            <>
                                <p className="tab-description">
                                    Entrez le code à 6 caractères pour rejoindre un salon existant.
                                </p>
                                <input
                                    type="text"
                                    placeholder="CODE"
                                    className="join-input"
                                    value={joinCode}
                                    onChange={handleJoinCodeChange}
                                />
                                <Button
                                    variant="primary"
                                    className="btn-full"
                                    disabled={isJoinDisabled}
                                    onClick={handleJoinRoom}
                                >
                                    {isLoading ? "Connexion..." : "Rejoindre le salon"}
                                </Button>
                            </>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default PlayPage;
