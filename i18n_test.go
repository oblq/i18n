package i18n

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oblq/sprbox"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

const (
	GEM = "GEM" // generic_error_message
)

var hardcodedLocs = map[string]map[string]*Localization{
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

func TestNewWithConfig(t *testing.T) {
	locConfigPath, err := NewWithConfig(&Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
		Path: "./example/i18n",
	})
	assert.Equal(t, nil, err)
	assert.Equal(t,
		"Something went wrong, please try again later Marco",
		locConfigPath.T("en", GEM, "Marco"))

	// test wrong config without tags
	_, err = NewWithConfig(&Config{
		Locales: []string{},
		Path:    "./example/i18n",
	})
	assert.NotEqual(t, nil, err)

	// test wrong config without Path nor Locs
	_, err = NewWithConfig(&Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
	})
	assert.NotEqual(t, nil, err)

	// test Locs
	locConfigMap, err := NewWithConfig(&Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
		Locs: hardcodedLocs,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t,
		locConfigMap.TP("en", false, GEM, "Marco"),
		"Something went wrong, please try again later Marco")
}

func TestNewWithConfigFile(t *testing.T) {
	locConfigFile, err := NewWithConfigFile("./i18n.yaml")
	assert.Equal(t, nil, err)
	assert.Equal(t,
		locConfigFile.T("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco")
}

func TestNewWithConfigWithLocs(t *testing.T) {
	// Will throw `log.Fatal()` if an error occour.
	locConfigMap, err := NewWithConfig(&Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
		Locs: hardcodedLocs,
	})

	assert.Equal(t, nil, err)

	assert.Equal(t,
		locConfigMap.TP("en", false, GEM, "Marco"),
		"Something went wrong, please try again later Marco")

	assert.Equal(t,
		locConfigMap.TP("en", true, GEM, "Marco"),
		"Some things went wrong, please try again later Marco")
}

func TestI18n_FileServer(t *testing.T) {
	localizer, err := NewWithConfig(&Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
		Locs: hardcodedLocs,
	})

	assert.Equal(t, nil, err)

	localizer.SetFileServer(
		map[string]http.Handler{
			language.English.String(): http.FileServer(http.Dir("./example/web/en")),
			language.Italian.String(): http.FileServer(http.Dir("./example/web/it")),
		},
	)

	// default language
	assert.HTTPBodyContains(t, localizer.ServeHTTP, "GET", "/", nil, "en")

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), "en"); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", "en", response.Body.String())
	}

	// it by header
	testingLocale := "it"
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Add(headerKey, testingLocale)
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), testingLocale); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}

	// en by header
	testingLocale = "en"
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Add(headerKey, testingLocale)
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), testingLocale); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}

	// it by cookies
	testingLocale = "it"
	request, _ = http.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{
		Name:    cookieKey2,
		Value:   testingLocale,
		Expires: time.Now().AddDate(0, 0, 7),
		Path:    "/",
	})
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), testingLocale); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}

	// unandled locale
	testingLocale = "es"
	request, _ = http.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{
		Name:    cookieKey2,
		Value:   testingLocale,
		Expires: time.Now().AddDate(0, 0, 7),
		Path:    "/",
	})
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), "en"); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}
}

func TestSpareConfig(t *testing.T) {
	type Box struct {
		Localizer I18n `sprs:"i18n"`
	}

	var toolBox Box
	if err := sprbox.NestedConfigsParser.Load(&toolBox, "./"); err != nil {
		t.Error("SpareConfig failed: ", err.Error())
	}

	assert.Equal(t,
		toolBox.Localizer.TP("en", false, GEM, "Marco"),
		"Something went wrong, please try again later Marco")
}

func TestTranslate(t *testing.T) {
	locConfigMap, err := NewWithConfig(&Config{
		Locales: []string{language.Italian.String()},
		Locs:    hardcodedLocs,
	})

	assert.Equal(t, nil, err)

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Add(headerKey, "it")

	assert.Equal(t,
		"Qualcosa è andato storto, riprova più tardi Marco",
		locConfigMap.AutoT(r, GEM, "Marco"))

	assert.Equal(t,
		"Qualcosa è andato storto, riprova più tardi Marco",
		locConfigMap.AutoTP(r, false, GEM, "Marco"))

	assert.Equal(t,
		"Alcune cose sono andate storte, riprova più tardi Marco",
		locConfigMap.AutoTP(r, true, GEM, "Marco"))
}
