package diff

import (
	"context"
	"net/http"
	"reflect"
)

type DiffModelConfig struct {
	Id       string `mapstructure:"id" json:"id,omitempty" gorm:"column:id" bson:"_id,omitempty" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Origin   string `mapstructure:"origin" json:"origin,omitempty" gorm:"column:origin" bson:"origin,omitempty" dynamodbav:"origin,omitempty" firestore:"origin,omitempty"`
	Value    string `mapstructure:"value" json:"value,omitempty" gorm:"column:value" bson:"value,omitempty" dynamodbav:"value,omitempty" firestore:"value,omitempty"`
	By       string `mapstructure:"by" json:"by,omitempty" gorm:"column:by" bson:"by,omitempty" dynamodbav:"by,omitempty" firestore:"by,omitempty"`
	Resource string `mapstructure:"resource" json:"resource,omitempty" gorm:"column:resource" bson:"resource,omitempty" dynamodbav:"resource,omitempty" firestore:"resource,omitempty"`
	Action   string `mapstructure:"action" json:"action,omitempty" gorm:"column:action" bson:"action,omitempty" dynamodbav:"action,omitempty" firestore:"action,omitempty"`
}
type DiffHandler struct {
	Error       func(context.Context, string)
	DiffService DiffService
	ModelType   reflect.Type
	IdNames     []string
	Indexes     map[string]int
	Offset      int
	Log         func(ctx context.Context, resource string, action string, success bool, desc string) error
	Resource    string
	Action      string
	Config      *DiffModelConfig
}

func NewDiffHandler(diffService DiffService, modelType reflect.Type, logError func(context.Context, string), config *DiffModelConfig, writeLog func(context.Context, string, string, bool, string) error, options ...int) *DiffHandler {
	return NewDiffHandlerWithKeys(diffService, modelType, logError, nil, config, writeLog, options...)
}
func NewDiffHandlerWithKeys(diffService DiffService, modelType reflect.Type, logError func(context.Context, string), idNames []string, config *DiffModelConfig, writeLog func(context.Context, string, string, bool, string) error, options ...int) *DiffHandler {
	offset := 1
	if len(options) > 0 {
		offset = options[0]
	}
	if idNames == nil || len(idNames) == 0 {
		idNames = getListFieldsTagJson(modelType)
	}
	indexes := getIndexes(modelType)
	var resource, action string
	if config != nil {
		resource = config.Resource
		action = config.Action
	}
	if len(resource) == 0 {
		resource = buildResourceName(modelType.Name())
	}
	if len(action) == 0 {
		action = "diff"
	}
	return &DiffHandler{Log: writeLog, DiffService: diffService, ModelType: modelType, IdNames: idNames, Indexes: indexes, Resource: resource, Offset: offset, Config: config, Error: logError}
}

func (c *DiffHandler) Diff(w http.ResponseWriter, r *http.Request) {
	id, err := buildId(r, c.ModelType, c.IdNames, c.Indexes, c.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		result, err := c.DiffService.Diff(r.Context(), id)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, c.Action, err, c.Log)
		} else {
			if c.Config == nil {
				succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
			} else {
				m := make(map[string]interface{})
				if result.Id != nil {
					m[c.Config.Id] = result.Id
				}
				if result.Origin != nil {
					m[c.Config.Origin] = result.Origin
				}
				if result.Value != nil {
					m[c.Config.Value] = result.Value
				}
				if len(result.By) > 0 {
					m[c.Config.By] = result.By
				}
				succeed(w, r, http.StatusOK, m, c.Log, c.Resource, c.Action)
			}
		}
	}
}
