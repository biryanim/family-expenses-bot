package openapi_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
)

const specPath = "../../../../api/openapi.yaml"

type Suite struct {
	suite.Suite

	loader *openapi3.Loader
}

func TestOpenAPI(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupTest() {
	s.loader = openapi3.NewLoader()
}

func (s *Suite) TestSpec_Valid() {
	ctx := context.Background()

	doc, err := s.loader.LoadFromFile(specPath)
	s.Require().NoError(err)

	s.NoError(doc.Validate(ctx))
	s.Equal("3.0.3", doc.OpenAPI)
}
