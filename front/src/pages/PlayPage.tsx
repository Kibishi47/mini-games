import { useState, useEffect } from "react";
import Button from "../components/common/Button";
import { http } from "../lib/api/http";
import { useWebSocket } from "../websocket/useWebSocket";
import { useAuth } from "../auth/AuthContext";
import type { Room } from "../websocket/WebSocketContext";
import "@/styles/pages/play.css";

const PlayPage = () => {
    const [activeTab, setActiveTab] = useState<"host" | "join">("host");
    const [joinCode, setJoinCode] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [showCopySuccess, setShowCopySuccess] = useState(false);
    const [games, setGames] = useState<any[]>([]);
    
    const { user } = useAuth();
    const ws = useWebSocket();
    const { roomInfo, setRoomInfo, players, status, selectedGame, gameConfig, setGameConfig, sendMessage } = ws || {};

    const isHost = roomInfo && user && roomInfo.hostId === user.id;

    // Fetch games from backend
    useEffect(() => {
        const fetchGames = async () => {
            try {
                const data = await http<any[]>("/games");
                setGames(data);
            } catch (err) {
                console.error("Failed to fetch games", err);
            }
        };
        fetchGames();
    }, []);

    const handleCopyCode = () => {
        if (!roomInfo) return;
        navigator.clipboard.writeText(roomInfo.code);
        setShowCopySuccess(true);
        setTimeout(() => setShowCopySuccess(false), 2000);
    };

    const handleSelectGame = (gameId: string) => {
        if (!isHost) return;
        
        const game = games.find(g => g.id === gameId);
        let initialConfig: Record<string, any> = {};
        if (game) {
            game.options?.forEach((opt: any) => {
                initialConfig[opt.id] = opt.defaultValue;
            });
        }

        sendMessage?.({
            type: "SELECT_GAME",
            payload: { gameId }
        });

        // Also send initial config
        if (Object.keys(initialConfig).length > 0) {
            sendMessage?.({
                type: "UPDATE_CONFIG",
                payload: initialConfig
            });
        }
    };

    const handleConfigChange = (id: string, value: any) => {
        if (!isHost) return;
        
        const newConfig = { ...gameConfig, [id]: value };
        sendMessage?.({
            type: "UPDATE_CONFIG",
            payload: newConfig
        });
        // We update locally too for immediate feedback
        setGameConfig?.(newConfig);
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
                            <div className="games-config-wrapper">
                                <div className="games-list">
                                    {games.map(game => (
                                        <div 
                                            key={game.id} 
                                            className={`game-item ${selectedGame === game.id ? "selected" : ""} ${!game.enabled ? "disabled" : ""}`}
                                            onClick={() => game.enabled && handleSelectGame(game.id)}
                                        >
                                            <span className="game-icon">{game.icon}</span>
                                            <div className="game-info">
                                                <span className="game-name">
                                                    {game.name}
                                                    {!game.enabled && <span className="maintenance-badge">Maintenance</span>}
                                                </span>
                                                <span className="game-desc">{game.description}</span>
                                            </div>
                                            {selectedGame === game.id && <div className="selected-check">✓</div>}
                                        </div>
                                    ))}
                                </div>

                                {selectedGame && games.find(g => g.id === selectedGame)?.options?.length > 0 && (
                                    <div className="game-options-panel">
                                        <h3 className="options-title">Options de {games.find(g => g.id === selectedGame)?.name}</h3>
                                        <div className="options-grid">
                                            {games.find(g => g.id === selectedGame)?.options.map((opt: any) => (
                                                <div key={opt.id} className="option-field">
                                                    <label>{opt.label}</label>
                                                    {opt.type === "range" || opt.type === "number" ? (
                                                        <div className="range-input-wrapper">
                                                            <input 
                                                                type={opt.type === "range" ? "range" : "number"}
                                                                min={opt.min}
                                                                max={opt.max}
                                                                step={opt.step || 1}
                                                                value={gameConfig?.[opt.id] ?? opt.defaultValue}
                                                                disabled={!isHost}
                                                                onChange={(e) => handleConfigChange(opt.id, Number(e.target.value))}
                                                            />
                                                            <span className="range-value">{gameConfig?.[opt.id] ?? opt.defaultValue}</span>
                                                        </div>
                                                    ) : opt.type === "boolean" ? (
                                                        <input 
                                                            type="checkbox" 
                                                            checked={gameConfig?.[opt.id] ?? opt.defaultValue}
                                                            disabled={!isHost}
                                                            onChange={(e) => handleConfigChange(opt.id, e.target.checked)}
                                                        />
                                                    ) : (
                                                        <input 
                                                            type="text" 
                                                            value={gameConfig?.[opt.id] ?? opt.defaultValue}
                                                            disabled={!isHost}
                                                            onChange={(e) => handleConfigChange(opt.id, e.target.value)}
                                                        />
                                                    )}
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="selected-game-display-wrapper">
                                <div className="selected-game-display">
                                    <div className="game-hero">
                                        <span className="hero-icon">{games.find(g => g.id === selectedGame)?.icon}</span>
                                        <h3>{games.find(g => g.id === selectedGame)?.name}</h3>
                                        <p>{games.find(g => g.id === selectedGame)?.description}</p>
                                    </div>
                                    <div className="waiting-host">
                                        <div className="loader-mini"></div>
                                        <span>En attente de l'hôte...</span>
                                    </div>
                                </div>
                                
                                {selectedGame && games.find(g => g.id === selectedGame)?.options?.length > 0 && (
                                    <div className="game-options-panel readonly">
                                        <h3 className="options-title">Configuration</h3>
                                        <div className="options-grid">
                                            {games.find(g => g.id === selectedGame)?.options.map((opt: any) => (
                                                <div key={opt.id} className="option-field">
                                                    <label>{opt.label}</label>
                                                    <div className="readonly-value">
                                                        {opt.type === "boolean" ? (gameConfig?.[opt.id] ? "Oui" : "Non") : (gameConfig?.[opt.id] ?? opt.defaultValue)}
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}
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
