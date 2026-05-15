package game

type ConfigOptionType string

const (
	ConfigOptionNumber  ConfigOptionType = "number"
	ConfigOptionString  ConfigOptionType = "string"
	ConfigOptionBoolean ConfigOptionType = "boolean"
	ConfigOptionRange   ConfigOptionType = "range"
)

type ConfigOption struct {
	ID           string           `json:"id"`
	Label        string           `json:"label"`
	Type         ConfigOptionType `json:"type"`
	DefaultValue any              `json:"defaultValue"`
	Min          *int             `json:"min,omitempty"`
	Max          *int             `json:"max,omitempty"`
	Step         *int             `json:"step,omitempty"`
}

type GameDefinition struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	Enabled     bool           `json:"enabled"`
	MinPlayers  int            `json:"minPlayers"`
	MaxPlayers  int            `json:"maxPlayers"`
	Options     []ConfigOption `json:"options"`
}
