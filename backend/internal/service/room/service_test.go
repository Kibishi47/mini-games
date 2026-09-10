package room

import (
	"strings"
	"testing"
)

func TestGenerateRoomCode(t *testing.T) {
	service := NewRoomService(nil)

	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := service.GenerateRoomCode()
		parts := strings.Split(code, "-")
		if len(parts) != 2 {
			t.Fatalf("Code mal formaté: %s", code)
		}
		if len(parts[0]) != 4 || len(parts[1]) != 2 {
			t.Fatalf("Longueur des parties incorrecte: %s", code)
		}
		if codes[code] {
			t.Fatalf("Collision de code détectée: %s", code)
		}
		codes[code] = true
	}
}

func TestGenerateSessionToken(t *testing.T) {
	service := NewRoomService(nil)

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := service.GenerateSessionToken()
		if len(token) != 32 {
			t.Fatalf("Longueur de token incorrecte: %d", len(token))
		}
		if tokens[token] {
			t.Fatalf("Collision de token détectée: %s", token)
		}
		tokens[token] = true
	}
}
