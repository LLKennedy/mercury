package proxy

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGettersAndSetters(t *testing.T) {
	t.Run("default server", func(t *testing.T) {
		var s *Server
		t.Run("check defaults", func(t *testing.T) {
			assert.Nil(t, s.getGrpcServer())
			assert.Nil(t, s.getAPI())
			assert.Nil(t, s.getInnerServer())
		})
		t.Run("set defaults", func(t *testing.T) {
			s.setGrpcServer(&grpc.Server{})
			exampleMap := map[string]map[string]apiMethod{
				"test": {
					"other": apiMethod{pattern: apiMethodPatternStructStruct, reflection: reflect.Method{Name: "method"}},
				},
			}
			s.setAPI(exampleMap)
			s.setInnerServer(12)
			assert.Equal(t, &grpc.Server{}, s.getGrpcServer())
			assert.Equal(t, exampleMap, s.getAPI())
			assert.Equal(t, 12, s.getInnerServer())
		})
	})
	t.Run("real server", func(t *testing.T) {
		s := &Server{}
		t.Run("check defaults", func(t *testing.T) {
			assert.Nil(t, s.getGrpcServer())
			assert.Nil(t, s.getAPI())
			assert.Nil(t, s.getInnerServer())
		})
		t.Run("set values", func(t *testing.T) {
			s.setGrpcServer(&grpc.Server{})
			exampleMap := map[string]map[string]apiMethod{
				"test": {
					"other": apiMethod{pattern: apiMethodPatternStructStruct, reflection: reflect.Method{Name: "method"}},
				},
			}
			s.setAPI(exampleMap)
			s.setInnerServer(12)
			assert.Equal(t, &grpc.Server{}, s.getGrpcServer())
			assert.Equal(t, exampleMap, s.getAPI())
			assert.Equal(t, 12, s.getInnerServer())
		})
	})
}
