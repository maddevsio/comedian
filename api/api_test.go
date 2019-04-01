package api

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/storage"
	yaml "gopkg.in/yaml.v2"
)

func TestSwaggerRoutesExistInEcho(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	sw, err := getSwagger()
	assert.NoError(t, err)
	db, err := storage.New(c)
	assert.NoError(t, err)
	api := New(c, db, nil)
	routes := api.echo.Routes()

	for k, v := range sw.Paths {
		m := v.(map[interface{}]interface{})
		for method := range m {
			found := false
			for _, route := range routes {
				if route.Path == "/swagger.yaml" {
					continue
				}
				path := replaceParams(route.Path)
				m := method.(string)

				s := strings.ToLower(route.Method)
				if strings.Contains(path, k) && m == s {
					found = true
				}
			}
			if !found {
				t.Errorf("could not find %v in routes for method %v", k, method)
			}
		}
	}
}

// func TestEchoRoutesExistInSwagger(t *testing.T) {
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	sw, err := getSwagger()
// 	assert.NoError(t, err)
// 	db, err := storage.New(c)
// 	assert.NoError(t, err)
// 	api := New(c, db, nil)
// 	routes := api.echo.Routes()

// 	for _, route := range routes {
// 		path := replaceParams(route.Path)
// 		found := false
// 		if path == "/swagger.yaml" {
// 			found = true
// 		}
// 		path = strings.Replace(path, sw.BasePath, "", -1)
// 		d, ok := sw.Paths[path]
// 		if !ok && !found {
// 			t.Errorf("could not find any documentation for %s path", path)
// 		}
// 		var method interface{}
// 		method = strings.ToLower(route.Method)
// 		route, ok := d.(map[interface{}]interface{})
// 		if !ok && !found {
// 			t.Errorf("[route] could not find documentation for %s path and %s method", path, method)
// 		}
// 		_, ok = route[method]
// 		if !ok && !found {
// 			t.Errorf("[method] could not find documentation for %s path and %s method", path, method)
// 		}
// 	}
// }

func TestGetSwagger(t *testing.T) {
	_, err := getSwagger()
	assert.NoError(t, err)
}

func getSwagger() (swagger, error) {
	var sw swagger
	data, err := ioutil.ReadFile("swagger.yaml")
	if err != nil {
		return sw, err
	}
	err = yaml.Unmarshal(data, &sw)
	return sw, err
}

func replaceParams(route string) string {
	if !echoRouteRegex.MatchString(route) {
		return route
	}
	matches := echoRouteRegex.FindAllStringSubmatch(route, -1)
	return fmt.Sprintf("%s{%s}%s", matches[0][1], matches[0][2], matches[0][3])
}

// func TestHandleEvent(t *testing.T) {
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	db, err := storage.New(c)
// 	assert.NoError(t, err)
// 	api := New(c, db, nil)

// 	type EventsAPICallbackEvent struct {
// 		Type        string           `json:"type"`
// 		Token       string           `json:"token"`
// 		TeamID      string           `json:"team_id"`
// 		APIAppID    string           `json:"api_app_id"`
// 		InnerEvent  *json.RawMessage `json:"event"`
// 		AuthedUsers []string         `json:"authed_users"`
// 		EventID     string           `json:"event_id"`
// 		EventTime   int              `json:"event_time"`
// 	}

// 	var testCase = []struct {
// 		Token      string
// 		challenge  string
// 		Type       string
// 		StatusCode int
// 		Err        error
// 	}{
// 		{"1569876", "run", "url_verification", 200, nil},
// 		{"156957876", "event", "event", 200, nil},
// 		{"1955486", "app_rate_limited", "app_rate_limited", 200, nil},
// 		{"156957876", "event_callback", "event_callback", 400, nil},
// 	}

// 	for _, tt := range testCase {
// 		t.Run("TestHandleEvent", func(t *testing.T) {
// 			event := map[string]string{"token": tt.Token, "challenge": tt.challenge, "type": tt.Type}

// 			jsonEvent, err := json.Marshal(event)
// 			assert.NoError(t, err)

// 			req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewBuffer(jsonEvent))
// 			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

// 			e := echo.New()
// 			rec := httptest.NewRecorder()
// 			C := e.NewContext(req, rec)

// 			err = api.handleEvent(C)
// 			assert.Equal(t, tt.StatusCode, rec.Code)
// 			assert.Equal(t, tt.Err, err)
// 		})
// 	}

// }
