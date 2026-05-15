package game

import "context"

type Service interface {
	ListGames(ctx context.Context) ([]GameDefinition, error)
}

type service struct {
	games []GameDefinition
}

func (s *service) ListGames(ctx context.Context) ([]GameDefinition, error) {
	var enabledGames []GameDefinition
	for _, g := range s.games {
		if g.Enabled {
			enabledGames = append(enabledGames, g)
		}
	}
	return enabledGames, nil
}

func NewService() Service {
	return &service{
		games: []GameDefinition{
			{
				ID:          "Wordle",
				Name:        "Wordle",
				Description: "Devinez le mot secret en 6 essais.",
				Icon:        "📝",
				Enabled:     true,
				MinPlayers:  1,
				MaxPlayers:  10,
				Options: []ConfigOption{
					{
						ID:           "wordLength",
						Label:        "Nombre de lettres",
						Type:         ConfigOptionRange,
						DefaultValue: 5,
						Min:          intPtr(4),
						Max:          intPtr(8),
						Step:         intPtr(1),
					},
					{
						ID:           "rounds",
						Label:        "Nombre de rounds",
						Type:         ConfigOptionNumber,
						DefaultValue: 3,
						Min:          intPtr(1),
						Max:          intPtr(10),
					},
					{
						ID:           "maxTime",
						Label:        "Temps max (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 60,
						Min:          intPtr(10),
						Max:          intPtr(300),
					},
					{
						ID:           "maxGuesses",
						Label:        "Essais max",
						Type:         ConfigOptionNumber,
						DefaultValue: 6,
						Min:          intPtr(1),
						Max:          intPtr(10),
					},
				},
			},
			{
				ID:          "BetweenLines",
				Name:        "Between Lines",
				Description: "Lisez entre les lignes pour gagner.",
				Icon:        "📖",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  10,
				Options: []ConfigOption{
					{
						ID:           "rounds",
						Label:        "Nombre de rounds",
						Type:         ConfigOptionNumber,
						DefaultValue: 5,
						Min:          intPtr(1),
						Max:          intPtr(10),
					},
				},
			},
			{
				ID:          "Checkers",
				Name:        "Dames",
				Description: "Le classique jeu de dames.",
				Icon:        "🏁",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  2,
				Options: []ConfigOption{
					{
						ID:           "turnTime",
						Label:        "Temps par tour (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 30,
						Min:          intPtr(5),
						Max:          intPtr(120),
					},
				},
			},
			{
				ID:          "Chess",
				Name:        "Échecs",
				Description: "Battez vos amis aux échecs.",
				Icon:        "♟️",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  2,
				Options: []ConfigOption{
					{
						ID:           "timeControl",
						Label:        "Temps total (min)",
						Type:         ConfigOptionNumber,
						DefaultValue: 10,
						Min:          intPtr(1),
						Max:          intPtr(60),
					},
					{
						ID:           "increment",
						Label:        "Incrément (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 0,
						Min:          intPtr(0),
						Max:          intPtr(30),
					},
				},
			},
			{
				ID:          "FourInARow",
				Name:        "Puissance 4",
				Description: "Alignez 4 jetons pour gagner.",
				Icon:        "🔴",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  2,
				Options: []ConfigOption{
					{
						ID:           "winCondition",
						Label:        "Jetons à aligner",
						Type:         ConfigOptionRange,
						DefaultValue: 4,
						Min:          intPtr(3),
						Max:          intPtr(5),
					},
					{
						ID:           "turnTime",
						Label:        "Temps par tour (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 15,
						Min:          intPtr(5),
						Max:          intPtr(60),
					},
				},
			},
			{
				ID:          "Trivia",
				Name:        "Trivia",
				Description: "Testez votre culture générale.",
				Icon:        "❓",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  10,
				Options: []ConfigOption{
					{
						ID:           "questionCount",
						Label:        "Nombre de questions",
						Type:         ConfigOptionNumber,
						DefaultValue: 10,
						Min:          intPtr(5),
						Max:          intPtr(50),
					},
					{
						ID:           "difficulty",
						Label:        "Difficulté",
						Type:         ConfigOptionString,
						DefaultValue: "médium",
					},
					{
						ID:           "timePerQuestion",
						Label:        "Temps par question (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 15,
						Min:          intPtr(5),
						Max:          intPtr(30),
					},
				},
			},
			{
				ID:          "Snake",
				Name:        "Snake",
				Description: "Ne vous mordez pas la queue !",
				Icon:        "🐍",
				Enabled:     false, // Exemple de maintenance
				MinPlayers:  1,
				MaxPlayers:  10,
				Options:     []ConfigOption{},
			},
			{
				ID:          "Minesweeper",
				Name:        "Démineur",
				Description: "Évitez toutes les mines.",
				Icon:        "💣",
				Enabled:     false,
				MinPlayers:  1,
				MaxPlayers:  10,
				Options: []ConfigOption{
					{
						ID:           "gridSize",
						Label:        "Taille de la grille",
						Type:         ConfigOptionRange,
						DefaultValue: 10,
						Min:          intPtr(8),
						Max:          intPtr(20),
					},
					{
						ID:           "minesCount",
						Label:        "Nombre de mines",
						Type:         ConfigOptionNumber,
						DefaultValue: 15,
						Min:          intPtr(5),
						Max:          intPtr(50),
					},
				},
			},
			{
				ID:          "Werewolf",
				Name:        "Loup-Garou",
				Description: "Éliminez les loups avant qu'ils ne vous mangent.",
				Icon:        "🐺",
				Enabled:     false,
				MinPlayers:  5,
				MaxPlayers:  20,
				Options: []ConfigOption{
					{
						ID:           "werewolfCount",
						Label:        "Nombre de loups",
						Type:         ConfigOptionNumber,
						DefaultValue: 2,
						Min:          intPtr(1),
						Max:          intPtr(5),
					},
					{
						ID:           "discussionTime",
						Label:        "Temps de discussion (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 120,
						Min:          intPtr(30),
						Max:          intPtr(300),
					},
					{
						ID:           "hasSeer",
						Label:        "Inclure la Voyante",
						Type:         ConfigOptionBoolean,
						DefaultValue: true,
					},
					{
						ID:           "hasHunter",
						Label:        "Inclure le Chasseur",
						Type:         ConfigOptionBoolean,
						DefaultValue: true,
					},
				},
			},
			{
				ID:          "Wavelength",
				Name:        "Wavelength",
				Description: "Êtes-vous sur la même longueur d'onde ?",
				Icon:        "🌊",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  12,
				Options: []ConfigOption{
					{
						ID:           "targetPoints",
						Label:        "Points pour gagner",
						Type:         ConfigOptionNumber,
						DefaultValue: 10,
						Min:          intPtr(5),
						Max:          intPtr(20),
					},
				},
			},
			{
				ID:          "Tapple",
				Name:        "Tapple",
				Description: "Trouvez un mot avant que le temps ne s'écoule.",
				Icon:        "🍎",
				Enabled:     false,
				MinPlayers:  2,
				MaxPlayers:  8,
				Options: []ConfigOption{
					{
						ID:           "timeLimit",
						Label:        "Temps par lettre (sec)",
						Type:         ConfigOptionNumber,
						DefaultValue: 10,
						Min:          intPtr(5),
						Max:          intPtr(20),
					},
				},
			},
			{
				ID:          "BombDroper",
				Name:        "Bomb Droper",
				Description: "Évitez les bombes qui tombent du ciel.",
				Icon:        "💣",
				Enabled:     false,
				MinPlayers:  1,
				MaxPlayers:  10,
				Options: []ConfigOption{
					{
						ID:           "lives",
						Label:        "Vies",
						Type:         ConfigOptionNumber,
						DefaultValue: 3,
						Min:          intPtr(1),
						Max:          intPtr(5),
					},
					{
						ID:           "bombSpeed",
						Label:        "Vitesse des bombes",
						Type:         ConfigOptionRange,
						DefaultValue: 5,
						Min:          intPtr(1),
						Max:          intPtr(10),
					},
				},
			},
		},
	}
}

func intPtr(i int) *int {
	return &i
}
