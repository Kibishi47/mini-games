import { useAuth } from "../auth/AuthContext";
import Button from "../components/common/Button";
import { useMemo } from "react";
import "@/styles/pages/auth.css";

const ProfilePage = () => {
    const { user, logout } = useAuth();

    const avatarStyle = useMemo(() => {
        if (!user) return {};

        // Generate a random background color based on username
        let hash = 0;
        for (let i = 0; i < user.username.length; i++) {
            hash = user.username.charCodeAt(i) + ((hash << 5) - hash);
        }

        const h = Math.abs(hash) % 360;
        return {
            backgroundColor: `hsl(${h}, 70%, 50%)`,
            width: "120px",
            height: "120px",
            borderRadius: "50%",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: "3rem",
            fontWeight: "bold",
            color: "white",
            textTransform: "uppercase" as const,
            margin: "0 auto 2rem"
        };
    }, [user]);

    if (!user) return null;

    return (
        <div className="container" style={{ paddingTop: "80px" }}>
            <div className="auth-container">
                <div className="auth-card" style={{ maxWidth: "600px" }}>
                    <h1 className="auth-title text-gradient">Mon Profil</h1>

                    <div style={avatarStyle}>
                        {user.username.charAt(0)}
                    </div>

                    <form style={{ display: "flex", flexDirection: "column", gap: "1.5rem" }} onSubmit={(e) => e.preventDefault()}>
                        <div className="form-group">
                            <label className="form-label">Nom d'utilisateur (non modifiable)</label>
                            <input
                                type="text"
                                className="form-input"
                                value={user.username}
                                readOnly
                                style={{ opacity: 0.7, cursor: "not-allowed" }}
                            />
                        </div>

                        <div className="form-group">
                            <label className="form-label">Nom d'affichage</label>
                            <input
                                type="text"
                                className="form-input"
                                defaultValue={user.displayName}
                                placeholder="Votre nom d'affichage"
                            />
                        </div>

                        <div className="form-group">
                            <label className="form-label">URL de l'avatar (optionnel)</label>
                            <input
                                type="text"
                                className="form-input"
                                defaultValue={user.avatarUrl}
                                placeholder="https://..."
                            />
                        </div>

                        <div className="form-group">
                            <label className="form-label">Membre depuis le</label>
                            <input
                                type="text"
                                className="form-input"
                                value={new Date(user.createdAt).toLocaleDateString()}
                                readOnly
                                style={{ opacity: 0.7, cursor: "not-allowed" }}
                            />
                        </div>

                        <Button variant="primary" className="btn-full" disabled style={{ marginTop: "1rem" }}>
                            Sauvegarder les modifications (Bientôt)
                        </Button>
                    </form>

                    <div style={{ marginTop: "2rem", paddingTop: "1.5rem", borderTop: "1px solid rgba(255,255,255,0.05)", textAlign: "center" }}>
                        <Button
                            variant="danger"
                            onClick={logout}
                            className="btn-full"
                        >
                            Déconnexion
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ProfilePage;
