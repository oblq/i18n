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

func TestNewWConfigFilePath(t *testing.T) {
	locConfigFile := New("./i18n.yaml", nil)
	assert.Equal(
		t,
		locConfigFile.T("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")
}

func TestNewWPath(t *testing.T) {
	locConfigPath := New(
		"",
		&Config{
			Locales: []string{
				language.English.String(),
				language.Italian.String(),
			},
			Path: "./example/i18n",
		},
	)
	assert.Equal(
		t,
		locConfigPath.T("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")
}

func TestNewWLocs(t *testing.T) {
	// Will throw `log.Fatal()` if an error occour.
	locConfigMap := New("", &Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
		Locs: hardcodedLocs,
	})
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

func TestI18n_FileServer(t *testing.T) {
	localizer := New("", &Config{
		Locales: []string{
			language.English.String(),
			language.Italian.String(),
		},
		Locs: hardcodedLocs,
	})
	localizer.SetFileServer(
		map[string]http.Handler{
			language.English.String(): http.FileServer(http.Dir("./example/web/en")),
			language.Italian.String(): http.FileServer(http.Dir("./example/web/it")),
		},
	)

	// default language
	assert.HTTPBodyContains(t, localizer.ServeHTTP, "GET", "/", nil, "en", "invalid page server by default")

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
	testingLocale = "cz"
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
		Localizer I18n `sprbox:"i18n"`
	}

	var toolBox Box
	if err := sprbox.LoadToolBox(&toolBox, "./"); err != nil {
		t.Error("SpareConfig failed: ", err.Error())
	}

	assert.Equal(
		t,
		toolBox.Localizer.TP("en", false, GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")
}

func TestNoLocalesPassed(t *testing.T) {
	locConfigMap := New("", &Config{
		Locales: []string{},
		Locs:    hardcodedLocs,
	})
	assert.Equal(
		t,
		locConfigMap.TP("en", false, GEM, "Marco"),
		"Something went wrong, please try again later Marco",
		"wrong localization")
}

func TestTranslate(t *testing.T) {
	locConfigMap := New("", &Config{
		Locales: []string{language.Italian.String()},
		Locs:    hardcodedLocs,
	})

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Add(headerKey, "it")

	assert.Equal(t,
		"Qualcosa è andato storto, riprova più tardi Marco",
		locConfigMap.AutoT(r, GEM, "Marco"),
		"wrong localization")

	assert.Equal(t,
		"Qualcosa è andato storto, riprova più tardi Marco",
		locConfigMap.AutoTP(r, false, GEM, "Marco"),
		"wrong localization")

	assert.Equal(t,
		"Alcune cose sono andate storte, riprova più tardi Marco",
		locConfigMap.AutoTP(r, true, GEM, "Marco"),
		"wrong localization")
}
