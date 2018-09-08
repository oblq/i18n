package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

const (
	GEM = "GEM" // generic_error_message
)

var hardcodedLocs = map[string]map[string]Localization{
	language.English.String(): {
		GEM: {
			One:   "Something went wrong, please try again later %s",
			Other: "Some things went wrong, please try again later %s",
		},
	},
	language.Italian.String(): {
		GEM: {
			One:   "Qualcosa è andato storto, riprova più tardi %s",
			Other: "Alcune cose sono andate storte, riprova più tardi %s",
		},
	},
}

func TestNewConfigFilePath(t *testing.T) {
	locConfigFile := New("./i18n.yaml", nil, nil)
	assert.Equal(
		t,
		locConfigFile.T("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")
}

func TestNewConfig(t *testing.T) {
	locConfigPath := New(
		"",
		&Config{
			Locales: []string{
				language.English.String(),
				language.Italian.String(),
			},
			Path: "./example/i18n",
		},
		nil,
	)
	assert.Equal(
		t,
		locConfigPath.T("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")
}

func TestNewLocs(t *testing.T) {
	// Will throw `log.Fatal()` if an error occour.
	locConfigMap := New("", nil, hardcodedLocs)
	assert.Equal(
		t,
		locConfigMap.TP("en", false, GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")

	assert.Equal(
		t,
		locConfigMap.TP("en", true, GEM, "Marco"),
		"Some things went wrong, please try again later Marco",
		"wrong localization")
}
