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

func TestEchoRoutesExistInSwagger(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	sw, err := getSwagger()
	assert.NoError(t, err)
	db, err := storage.New(c)
	assert.NoError(t, err)
	api := New(c, db, nil)
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
			t.Errorf("[route] could not find documentation for %s path and %s method", path, method)
		}
		_, ok = route[method]
		if !ok && !found {
			t.Errorf("[method] could not find documentation for %s path and %s method", path, method)
		}
	}
}

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
