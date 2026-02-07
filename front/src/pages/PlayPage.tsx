import { useState } from "react";
import Button from "../components/common/Button";
import "@/styles/pages/play.css";

const PlayPage = () => {
    const [activeTab, setActiveTab] = useState<"host" | "join">("host");
    const [joinCode, setJoinCode] = useState("");

    const handleJoinCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        // Keep only numbers, limit to 4 characters
        const value = e.target.value.replace(/\D/g, '').slice(0, 4);
        setJoinCode(value);
    };

    const isJoinDisabled = joinCode.length !== 4;

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
                        {activeTab === "host" ? (
                            <>
                                <p style={{ color: "rgba(255,255,255,0.7)" }}>
                                    Créez un salon et invitez vos amis à vous rejoindre avec un code.
                                </p>
                                <Button variant="primary" className="btn-full" onClick={() => alert("Création de salon...")}>
                                    Créer un salon
                                </Button>
                            </>
                        ) : (
                            <>
                                <p style={{ color: "rgba(255,255,255,0.7)" }}>
                                    Entrez le code à 4 chiffres pour rejoindre un salon existant.
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
                                    onClick={() => alert(`Rejoindre le salon: ${joinCode}`)}
                                >
                                    Rejoindre le salon
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
