package character

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	DefaultFociRows = 1
	MaxFociRows     = 10

	DefaultWeaponRows      = 5
	DefaultReadiedItemRows = 10
	DefaultStowedItemRows  = 20

	MaxSpellRows          = 50
	MaxMagicArtRows       = 25
	MaxMagicTraditionRows = 5
)

// Sheet содержит все данные листа,
// которые хранятся внутри JSON.
//
// Name, Level и Class сюда не входят:
// они уже лежат отдельными колонками characters.
type Sheet struct {
	// Основная информация
	Player      string `json:"player"`
	Homeland    string `json:"homeland"`
	Occupation  string `json:"occupation"`
	RaceSpecies string `json:"race_species"`
	Goal        string `json:"goal"`
	Description string `json:"description"`

	// Background и класс
	Background        string `json:"background"`
	BackgroundDetails string `json:"background_details"`
	Benefits          string `json:"benefits"`

	// Развитие
	XP string `json:"xp"`

	// Основные характеристики
	Attributes Attributes `json:"attributes"`

	// Ресурсы
	HP           HitPoints    `json:"hp"`
	SystemStrain SystemStrain `json:"system_strain"`

	// Спасброски
	Saves Saves `json:"saves"`

	// Боевые значения
	Combat Combat `json:"combat"`
	Armor  Armor  `json:"armor"`

	// Навыки
	Skills       Skills `json:"skills"`
	SkillPoints  string `json:"skill_points"`
	ExpertPoints string `json:"expert_points"`

	// Переменные списки
	Foci         []Focus       `json:"foci"`
	Weapons      []Weapon      `json:"weapons"`
	ReadiedItems []ReadiedItem `json:"readied_items"`
	StowedItems  []StowedItem  `json:"stowed_items"`

	// Инвентарь
	ReadiedMaxLoad string `json:"readied_max_load"`
	StowedMaxLoad  string `json:"stowed_max_load"`
	Ammunition     string `json:"ammunition"`

	// Пока это текстовое поле.
	// Позже при желании сюда можно добавить изображение.
	Property      Property `json:"property"`
	SketchOrSigil string   `json:"sketch_or_sigil"`

	Spells          []Spell          `json:"spells"`
	MagicTraditions []MagicTradition `json:"magic_traditions"`
	MagicArts       []MagicArt       `json:"magic_arts"`
}

// Attributes.
//
// Старые JSON-имена strength, dexterity и т. д.
// намеренно сохраняем, чтобы уже созданные персонажи
// продолжали нормально загружаться.
type Attributes struct {
	Strength    string `json:"strength"`
	StrengthMod string `json:"strength_mod"`

	Dexterity    string `json:"dexterity"`
	DexterityMod string `json:"dexterity_mod"`

	Constitution    string `json:"constitution"`
	ConstitutionMod string `json:"constitution_mod"`

	Intelligence    string `json:"intelligence"`
	IntelligenceMod string `json:"intelligence_mod"`

	Wisdom    string `json:"wisdom"`
	WisdomMod string `json:"wisdom_mod"`

	Charisma    string `json:"charisma"`
	CharismaMod string `json:"charisma_mod"`
}

type HitPoints struct {
	Current string `json:"current"`
	Max     string `json:"max"`
}

type SystemStrain struct {
	Current string `json:"current"`
	Max     string `json:"max"`
}

type Saves struct {
	Physical string `json:"physical"`
	Evasion  string `json:"evasion"`
	Mental   string `json:"mental"`
	Luck     string `json:"luck"`
}

type Combat struct {
	BaseAttackBonus string `json:"base_attack_bonus"`
	MeleeAttack     string `json:"melee_attack"`
	RangedAttack    string `json:"ranged_attack"`
	Initiative      string `json:"initiative"`
}

type Armor struct {
	DexMod    string `json:"dex_mod"`
	WornArmor string `json:"worn_armor"`
	AC        string `json:"ac"`
	Special   string `json:"special"`
}

// Skills содержит стандартный список навыков WWN.
type Skills struct {
	Administer string `json:"administer"`
	Connect    string `json:"connect"`
	Convince   string `json:"convince"`
	Craft      string `json:"craft"`
	Exert      string `json:"exert"`
	Heal       string `json:"heal"`
	Know       string `json:"know"`
	Lead       string `json:"lead"`
	Magic      string `json:"magic"`
	Notice     string `json:"notice"`
	Perform    string `json:"perform"`
	Pray       string `json:"pray"`
	Punch      string `json:"punch"`
	Ride       string `json:"ride"`
	Sail       string `json:"sail"`
	Shoot      string `json:"shoot"`
	Sneak      string `json:"sneak"`
	Stab       string `json:"stab"`
	Survive    string `json:"survive"`
	Trade      string `json:"trade"`

	// У Work может быть специализация,
	// поэтому отдельно храним её название.
	WorkName string `json:"work_name"`
	Work     string `json:"work"`
}

type Focus struct {
	Name        string `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

type Weapon struct {
	Name         string `json:"name"`
	HitBonus     string `json:"hit_bonus"`
	Damage       string `json:"damage"`
	Range        string `json:"range"`
	SpecialShock string `json:"special_shock"`
}

type ReadiedItem struct {
	Name        string `json:"name"`
	Encumbrance string `json:"encumbrance"`
	Disabled    bool   `json:"disabled"`
}

type StowedItem struct {
	Name        string `json:"name"`
	Encumbrance string `json:"encumbrance"`
	Disabled    bool   `json:"disabled"`
}

type Property struct {
	Silver            string `json:"silver"`
	Gold              string `json:"gold"`
	StoredPossessions string `json:"stored_possessions"`
}

type Spell struct {
	Name        string `json:"name"`
	Tradition   string `json:"tradition"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

type MagicTradition struct {
	Name          string `json:"name"`
	EffortMax     string `json:"effort_max"`
	EffortCurrent string `json:"effort_current"`
}

type MagicArt struct {
	Name          string `json:"name"`
	Tradition     string `json:"tradition"`
	EffortSpentOn string `json:"effort_spent_on"`
	Description   string `json:"description"`
}

// EnsureRows гарантирует, что новый или старый персонаж
// получает несколько пустых строк для редактирования.
//
// Позже JavaScript позволит добавлять и удалять строки
// без фиксированного ограничения.
func (s *Sheet) EnsureRows() {
	s.Foci = compactFoci(s.Foci)

	if len(s.Foci) == 0 {
		s.Foci = append(s.Foci, Focus{})
	}

	for len(s.Weapons) < DefaultWeaponRows {
		s.Weapons = append(s.Weapons, Weapon{})
	}

	for len(s.ReadiedItems) < DefaultReadiedItemRows {
		s.ReadiedItems = append(
			s.ReadiedItems,
			ReadiedItem{},
		)
	}

	for len(s.StowedItems) < DefaultStowedItemRows {
		s.StowedItems = append(
			s.StowedItems,
			StowedItem{},
		)
	}

	if len(s.Spells) == 0 {
		s.Spells = append(s.Spells, Spell{})
	}

	if len(s.MagicTraditions) == 0 {
		s.MagicTraditions = append(
			s.MagicTraditions,
			MagicTradition{},
		)
	}

	if len(s.MagicArts) == 0 {
		s.MagicArts = append(
			s.MagicArts,
			MagicArt{},
		)
	}
}

func compactFoci(foci []Focus) []Focus {
	result := make([]Focus, 0, len(foci))

	for _, focus := range foci {
		if strings.TrimSpace(focus.Name) == "" &&
			strings.TrimSpace(focus.Level) == "" &&
			strings.TrimSpace(focus.Description) == "" {
			continue
		}

		result = append(result, focus)
	}

	return result
}

// DecodeSheet превращает JSON из SQLite в Sheet.
func DecodeSheet(data string) (Sheet, error) {
	var sheet Sheet

	data = strings.TrimSpace(data)

	if data == "" || data == "{}" {
		sheet.EnsureRows()
		return sheet, nil
	}

	err := json.Unmarshal(
		[]byte(data),
		&sheet,
	)

	if err != nil {
		return Sheet{}, fmt.Errorf(
			"decode character sheet: %w",
			err,
		)
	}

	sheet.EnsureRows()

	return sheet, nil
}

// Encode превращает Sheet обратно в JSON.
func (s Sheet) Encode() (string, error) {
	data, err := json.Marshal(s)

	if err != nil {
		return "", fmt.Errorf(
			"encode character sheet: %w",
			err,
		)
	}

	return string(data), nil
}
