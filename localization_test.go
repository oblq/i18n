package i18n

import (
	"testing"

	"golang.org/x/text/language"
	"gotest.tools/assert"
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

func TestI18n_UnmarshalLocalizationBytes(t *testing.T) {
	hardcodedLocs := make(map[string][]byte)
	hardcodedLocs[language.English.String()] = []byte(en)
	hardcodedLocs[language.Italian.String()] = []byte(it)

	localizer := NewI18n(
		"",
		&Config{
			Locales: []string{
				language.English.String(),
				language.Italian.String(),
			},
			LocalizationsBytes: hardcodedLocs,
		},
	)

	assert.NilError(t, localizer.UnmarshalLocalizationBytes(hardcodedLocs))
}
