package api

import (
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
)

func TestSwaggerRoutesExistInEcho(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	sw, err := getSwagger()
	assert.NoError(t, err)
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	db, err := storage.New(c)
	assert.NoError(t, err)
	comedian := comedianbot.New(bundle, db)
	api, err := New(c, db, comedian)
	assert.NoError(t, err)
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

func TestEchoRoutesExistInSwagger(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	sw, err := getSwagger()
	assert.NoError(t, err)
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	db, err := storage.New(c)
	assert.NoError(t, err)
	comedian := comedianbot.New(bundle, db)
	api, err := New(c, db, comedian)
	assert.NoError(t, err)

	routes := api.echo.Routes()
	for _, route := range routes {
		path := replaceParams(route.Path)
		found := false
		if path == "/swagger.yaml" {
			found = true
		}
		path = strings.Replace(path, sw.BasePath, "", -1)
		d, ok := sw.Paths[path]
		if !ok && !found {
			t.Errorf("could not find any documentation for %s path", path)
		}
		var method interface{}
		method = strings.ToLower(route.Method)
		route, ok := d.(map[interface{}]interface{})
		if !ok && !found {
			t.Errorf("could not find documentation for %s path and %s method", path, method)
		}
		_, ok = route[method]
		if !ok && !found {
			t.Errorf("could not find documentation for %s path and %s method", path, method)
		}
	}
}
