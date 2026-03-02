import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "../styles/index.css";
import { AuthProvider } from "@/auth/AuthProvider";
import { WebSocketProvider } from "@/websocket/WebSocketProvider";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AuthProvider>
      <WebSocketProvider>
        <App />
      </WebSocketProvider>
    </AuthProvider>
  </React.StrictMode>
);