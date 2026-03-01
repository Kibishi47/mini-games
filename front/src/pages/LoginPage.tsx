import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Button from "../components/common/Button";
import { useAuth } from "../auth/AuthContext";
import { authApi } from "../auth/auth.api";
import "@/styles/pages/auth.css";

const LoginPage = () => {
    const { setAuth } = useAuth();
    const navigate = useNavigate();
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [formData, setFormData] = useState({
        username: "",
        password: "",
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setFormData((prev) => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setIsLoading(true);
        try {
            const { accessToken } = await authApi.login(formData);
            await setAuth(accessToken);
            navigate("/");
        } catch (err: any) {
            setError(err.body?.message || "Identifiants invalides");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="container" style={{ paddingTop: "80px" }}>
            <div className="auth-container">
                <div className="auth-card">
                    <h1 className="auth-title text-gradient">Connexion</h1>

                    {error && <div className="auth-error" style={{ color: "red", marginBottom: "1rem", textAlign: "center" }}>{error}</div>}

                    <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "1.5rem" }}>
                        <div className="form-group">
                            <label htmlFor="username" className="form-label">
                                Nom d'utilisateur
                            </label>
                            <input
                                type="text"
                                id="username"
                                name="username"
                                className="form-input"
                                placeholder="Entrez votre pseudo"
                                value={formData.username}
                                onChange={handleChange}
                                required
                            />
                        </div>

                        <div className="form-group">
                            <label htmlFor="password" className="form-label">
                                Mot de passe
                            </label>
                            <input
                                type="password"
                                id="password"
                                name="password"
                                className="form-input"
                                placeholder="Entrez votre mot de passe"
                                value={formData.password}
                                onChange={handleChange}
                                required
                            />
                        </div>

                        <Button variant="primary" type="submit" className="btn-full" style={{ marginTop: "1rem" }} disabled={isLoading}>
                            {isLoading ? "Connexion..." : "Se connecter"}
                        </Button>
                    </form>

                    <div className="form-footer">
                        Pas encore de compte ?
                        <Link to="/register" className="auth-link">
                            S'inscrire
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;
