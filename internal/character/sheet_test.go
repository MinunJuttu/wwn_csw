package character

import "testing"

func TestDecodeOldSheet(t *testing.T) {
	oldJSON := `{
		"attributes": {
			"strength": "12",
			"dexterity": "14",
			"constitution": "10",
			"intelligence": "9",
			"wisdom": "11",
			"charisma": "13"
		},
		"hp": {
			"current": "5",
			"max": "8"
		}
	}`

	sheet, err := DecodeSheet(oldJSON)
	if err != nil {
		t.Fatalf(
			"DecodeSheet returned error: %v",
			err,
		)
	}

	if sheet.Attributes.Strength != "12" {
		t.Errorf(
			"expected strength 12, got %q",
			sheet.Attributes.Strength,
		)
	}

	if sheet.Attributes.Dexterity != "14" {
		t.Errorf(
			"expected dexterity 14, got %q",
			sheet.Attributes.Dexterity,
		)
	}

	if sheet.HP.Current != "5" {
		t.Errorf(
			"expected current HP 5, got %q",
			sheet.HP.Current,
		)
	}

	if sheet.HP.Max != "8" {
		t.Errorf(
			"expected max HP 8, got %q",
			sheet.HP.Max,
		)
	}
}

func TestDecodeSheetCreatesDefaultRows(t *testing.T) {
	sheet, err := DecodeSheet("{}")
	if err != nil {
		t.Fatalf(
			"DecodeSheet returned error: %v",
			err,
		)
	}

	if len(sheet.Foci) != DefaultFociRows {
		t.Errorf(
			"expected %d focus rows, got %d",
			DefaultFociRows,
			len(sheet.Foci),
		)
	}

	if len(sheet.Weapons) != DefaultWeaponRows {
		t.Errorf(
			"expected %d weapon rows, got %d",
			DefaultWeaponRows,
			len(sheet.Weapons),
		)
	}

	if len(sheet.ReadiedItems) != DefaultReadiedItemRows {
		t.Errorf(
			"expected %d readied item rows, got %d",
			DefaultReadiedItemRows,
			len(sheet.ReadiedItems),
		)
	}
}

func TestEncodeAndDecodeSheet(t *testing.T) {
	original := Sheet{
		Player: "MinunJuttu",

		Attributes: Attributes{
			Strength:    "15",
			StrengthMod: "+1",
		},

		Saves: Saves{
			Physical: "13",
		},
	}

	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf(
			"Encode returned error: %v",
			err,
		)
	}

	decoded, err := DecodeSheet(encoded)
	if err != nil {
		t.Fatalf(
			"DecodeSheet returned error: %v",
			err,
		)
	}

	if decoded.Player != original.Player {
		t.Errorf(
			"expected player %q, got %q",
			original.Player,
			decoded.Player,
		)
	}

	if decoded.Attributes.Strength != "15" {
		t.Errorf(
			"expected strength 15, got %q",
			decoded.Attributes.Strength,
		)
	}

	if decoded.Saves.Physical != "13" {
		t.Errorf(
			"expected physical save 13, got %q",
			decoded.Saves.Physical,
		)
	}
}
