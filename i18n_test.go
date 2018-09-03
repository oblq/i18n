package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

const (
	en = `
GEM:
  one: "Something went wrong, please try again later %s"
  other: "Some things went wrong, please try again later %s"
`

	it = `
GEM:
  one: "Qualcosa è andato storto, riprova più tardi %s"
  other: "Alcune cose sono andate storte, riprova più tardi %s"
`
)

func TestNew(t *testing.T) {
	locConfigFile := New("./i18n.yaml", nil)
	assert.Equal(
		t,
		locConfigFile.T("en", false, "GEM", "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")

	locConfigPath := New(
		"",
		&Config{
			Locales: []string{
				language.English.String(),
				language.Italian.String(),
			},
			LocalizationsPath: "./example/i18n",
		},
	)
	assert.Equal(
		t,
		locConfigPath.T("en", false, "GEM", "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")

	hardcodedLocs := make(map[string]string)
	hardcodedLocs[language.English.String()] = en
	hardcodedLocs[language.Italian.String()] = it
	// Will throw `log.Fatal()` if an error occour.
	locConfigMap := New(
		"",
		&Config{
			Locales: []string{
				language.English.String(),
				language.Italian.String(),
			},
			LocalizationsMap: hardcodedLocs,
		},
	)
	assert.Equal(
		t,
		locConfigMap.T("en", false, "GEM", "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")

	assert.Equal(
		t,
		locConfigMap.T("en", true, "GEM", "Marco"),
		"Some things went wrong, please try again later Marco",
		"wrong localization")
}
