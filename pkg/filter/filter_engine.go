package filter

import (
	"encoding/json"
	"errors"

	"github.com/jmespath/go-jmespath"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type FilterEngine struct {
	expr *jmespath.JMESPath
}

func NewFilterEngine(expr *jmespath.JMESPath) *FilterEngine {
	return &FilterEngine{expr: expr}
}

type filtrableResource struct {
	Attr     interface{}
	Res      resource.Resource
	Type, Id string
}

func (e *FilterEngine) Run(resources []resource.Resource) ([]resource.Resource, error) {

	if e.expr == nil {
		return nil, errors.New("expression is nil")
	}

	// We convert a list of resource in a list of DTO to run JMESPath on
	filtrableResources := make([]filtrableResource, 0, len(resources))
	for _, res := range resources {
		// We need to serialize all attributes to untyped interface from JMESPath to work
		// map[string]string and map[string]SomeThing will not work without it
		// https://github.com/jmespath/go-jmespath/issues/22
		ctyVal := res.CtyValue()
		if ctyVal == nil {
			ctyVal = &cty.EmptyObjectVal
		}
		bytes, _ := ctyjson.Marshal(*ctyVal, ctyVal.Type())
		var attrs interface{}
		_ = json.Unmarshal(bytes, &attrs)
		f := filtrableResource{
			Attr: attrs,
			Res:  res,
			Id:   res.TerraformId(),
			Type: res.TerraformType(),
		}
		filtrableResources = append(
			filtrableResources,
			f,
		)
	}

	// Do the filter
	JMESPathOutput, err := e.expr.Search(filtrableResources)
	if err != nil {
		return nil, err
	}

	// Convert back filtered results into a resource list
	filteredRawList := JMESPathOutput.([]interface{})
	results := make([]resource.Resource, 0, len(filteredRawList))
	for _, elem := range filteredRawList {
		results = append(results, elem.(filtrableResource).Res)
	}

	return results, nil
}
