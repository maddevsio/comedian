package api

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/stretchr/testify/assert"
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
