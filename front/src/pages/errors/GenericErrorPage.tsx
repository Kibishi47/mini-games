import Button from "../../components/common/Button";

const GenericErrorPage = ({ error }: { error?: any }) => {
    return (
        <div
            style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                minHeight: "100vh",
                textAlign: "center",
                padding: "2rem",
            }}
        >
            <h1
                className="text-gradient"
                style={{
                    fontSize: "5rem",
                    margin: 0,
                    fontWeight: 800,
                    lineHeight: 1,
                }}
            >
                Oups !
            </h1>
            <h2
                style={{
                    fontSize: "1.8rem",
                    margin: "1.5rem 0",
                    color: "var(--color-text)",
                }}
            >
                Une erreur inattendue est survenue
            </h2>
            <p
                style={{
                    maxWidth: "600px",
                    color: "rgba(255, 255, 255, 0.6)",
                    marginBottom: "2rem",
                    fontSize: "1.1rem",
                }}
            >
                Nous nous excusons pour la gêne occasionnée. N'hésitez pas à essayer de rafraîchir la page ou à revenir plus tard.
            </p>

            {error && (
                <pre style={{
                    background: "rgba(0, 0, 0, 0.3)",
                    padding: "1rem",
                    borderRadius: "8px",
                    color: "#ff8080",
                    fontSize: "0.8rem",
                    textAlign: "left",
                    maxWidth: "80%",
                    overflow: "auto",
                    marginBottom: "2rem"
                }}>
                    <code>{error.message || String(error)}</code>
                </pre>
            )}

            <Button to="/" variant="primary">
                Retour à l'accueil
            </Button>
        </div>
    );
};

export default GenericErrorPage;
