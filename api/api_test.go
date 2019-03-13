package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
)

func TestSwaggerRoutesExistInEcho(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	sw, err := getSwagger()
	assert.NoError(t, err)
	api, err := New(c)
	assert.NoError(t, err)
	r := api.Routes()
	for k, v := range sw.Paths {
		m := v.(map[interface{}]interface{})
		for method := range m {
			found := false
			for _, route := range r {
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
	_ = sw
	assert.NoError(t, err)
	api, err := New(c)
	assert.NoError(t, err)
	r := api.Routes()
	for _, v := range r {
		path := replaceParams(v.Path)
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
		method = strings.ToLower(v.Method)
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
