package i18n

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oblq/swap"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

const (
	GEM       = "GEM"        // generic_error_message
	GEMPlural = "GEM.plural" // generic_error_message
)

var TestHTTPLookUpStrategy = []HTTPLocalePosition{
	{HTTPLocalePositionIDHeader, "Accept-Language"},
	{HTTPLocalePositionIDCookie, "lang"},
	{HTTPLocalePositionIDQuery, "lang"},
}

var hardcodedLocs = map[string]map[string]string{
	language.English.String(): {
		GEM:       "Something went wrong, please try again later %s",
		GEMPlural: "Some things went wrong, please try again later %s",
	},
	language.Italian.String(): {
		GEM:       "Qualcosa è andato storto, riprova più tardi %s",
		GEMPlural: "Alcune cose sono andate storte, riprova più tardi %s",
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
		locConfigMap.TP("en", GEM, "Marco"),
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
		locConfigMap.TP("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco")

	assert.Equal(t,
		locConfigMap.TP("en", GEMPlural, "Marco"),
		"Some things went wrong, please try again later Marco")
}

func TestI18n_FileServer(t *testing.T) {
	localizer, err := NewWithConfig(&Config{
		HTTPLookUpStrategy: TestHTTPLookUpStrategy,
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

	// it from header
	testingLocale := "it"
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Add(TestHTTPLookUpStrategy[0].Key, testingLocale)
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), testingLocale); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}

	// en from header
	testingLocale = "en"
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Add(TestHTTPLookUpStrategy[0].Key, testingLocale)
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), testingLocale); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}

	// it from cookies
	testingLocale = "it"
	request, _ = http.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{
		Name:    TestHTTPLookUpStrategy[2].Key,
		Value:   testingLocale,
		Expires: time.Now().AddDate(0, 0, 7),
		Path:    "/",
	})
	response = httptest.NewRecorder()
	localizer.ServeHTTP(response, request)
	if contains := strings.Contains(response.Body.String(), testingLocale); !contains {
		t.Errorf("Expected response body to contain \"%s\" but found \"%s\"", testingLocale, response.Body.String())
	}

	// unhandled locale
	testingLocale = "es"
	request, _ = http.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{
		Name:    TestHTTPLookUpStrategy[2].Key,
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
		Localizer I18n `swap:"i18n"`
	}

	var toolBox Box
	if err := swap.NewBuilder("./").Build(&toolBox); err != nil {
		t.Error("swap.Parse failed: ", err.Error())
	}

	assert.Equal(t,
		toolBox.Localizer.TP("en", GEM, "Marco"),
		"Something went wrong, please try again later Marco")
}

func TestTranslate(t *testing.T) {
	locConfigMap, err := NewWithConfig(&Config{
		Locales: []string{language.Italian.String()},
		Locs:    hardcodedLocs,
	})

	assert.Equal(t, nil, err)

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Add(TestHTTPLookUpStrategy[0].Key, "it")

	assert.Equal(t,
		"Qualcosa è andato storto, riprova più tardi Marco",
		locConfigMap.AutoT(r, GEM, "Marco"))

	assert.Equal(t,
		"Qualcosa è andato storto, riprova più tardi Marco",
		locConfigMap.AutoTP(r, GEM, "Marco"))

	assert.Equal(t,
		"Alcune cose sono andate storte, riprova più tardi Marco",
		locConfigMap.AutoTP(r, GEMPlural, "Marco"))
}

func TestMiddleware(t *testing.T) {
	tt := []struct {
		name     string
		position HTTPLocalePosition
		locale   string
		want     string
	}{
		{
			name: "with a pizza ID and quantity",
			position: HTTPLocalePosition{
				ID:  HTTPLocalePositionIDHeader,
				Key: "Accept-Language",
			},
			locale: "en",
			want:   "en",
		},
		{
			name: "with a pizza ID and quantity",
			position: HTTPLocalePosition{
				ID:  HTTPLocalePositionIDCookie,
				Key: "lang",
			},
			locale: "en",
			want:   "en",
		},
		{
			name: "with a pizza ID and quantity",
			position: HTTPLocalePosition{
				ID:  HTTPLocalePositionIDQuery,
				Key: "lang",
			},
			locale: "en",
			want:   "en",
		},
	}

	localizer, err := NewWithConfig(&Config{Locales: []string{language.Italian.String(), language.English.String()}})
	if err != nil {
		panic(err)
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/test", nil)
			responseRecorder := httptest.NewRecorder()

			switch tc.position.ID {
			case HTTPLocalePositionIDHeader:
				request.Header.Add(tc.position.Key, tc.locale)
			case HTTPLocalePositionIDCookie:
				request.AddCookie(&http.Cookie{
					Name:  tc.position.Key,
					Value: tc.locale,
				})
			case HTTPLocalePositionIDQuery:
				request.URL.RawQuery = fmt.Sprintf("%s=%s", tc.position.Key, tc.locale)
			}

			localizer.Config.HTTPLookUpStrategy = []HTTPLocalePosition{tc.position}
			handler := localizer.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				locale := r.Context().Value(MiddlewareContextLocaleKey)
				_, _ = w.Write([]byte(locale.(string)))
			}))

			handler.ServeHTTP(responseRecorder, request)
			if responseRecorder.Code != http.StatusOK {
				t.Errorf("bad status code '%d'", responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}
