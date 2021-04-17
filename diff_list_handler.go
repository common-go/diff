package diff

import (
	"context"
	"net/http"
	"reflect"
	"strings"
)

type DiffListHandler struct {
	Error           func(context.Context, string)
	DiffListService DiffListService
	ModelType       reflect.Type
	modelTypeId     reflect.Type
	IdNames         []string
	Log             func(ctx context.Context, resource string, action string, success bool, desc string) error
	Resource        string
	Action          string
	Config          *DiffModelConfig
}

func NewDiffListHandler(diffListService DiffListService, modelType reflect.Type, logError func(context.Context, string), config *DiffModelConfig, writeLog func(context.Context, string, string, bool, string) error) *DiffListHandler {
	return NewDiffListHandlerWithKeys(diffListService, modelType, logError, nil, config, writeLog)
}
func NewDiffListHandlerWithKeys(diffListService DiffListService, modelType reflect.Type, logError func(context.Context, string), idNames []string, config *DiffModelConfig, writeLog func(context.Context, string, string, bool, string) error) *DiffListHandler {
	if idNames == nil || len(idNames) == 0 {
		idNames = getJsonPrimaryKeys(modelType)
	}
	modelTypeId := newModelTypeID(modelType, idNames)
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
	return &DiffListHandler{Log: writeLog, DiffListService: diffListService, ModelType: modelType, modelTypeId: modelTypeId, IdNames: idNames, Resource: resource, Action: action, Config: config, Error: logError}
}

func (c *DiffListHandler) DiffList(w http.ResponseWriter, r *http.Request) {
	ids, err := buildIds(r, c.modelTypeId, c.IdNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		list, err := c.DiffListService.Diff(r.Context(), ids)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, c.Action, err, c.Log)
		} else {
			if c.Config == nil || list == nil || len(*list) == 0 {
				succeed(w, r, http.StatusOK, list, c.Log, c.Resource, c.Action)
			} else {
				l := make([]map[string]interface{}, 0)
				for _, result := range *list {
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
					l = append(l, m)
				}
				succeed(w, r, http.StatusOK, l, c.Log, c.Resource, c.Action)
			}
		}
	}
}

func newModelTypeID(modelType reflect.Type, idJsonNames []string) reflect.Type {
	model := reflect.New(modelType).Interface()
	value := reflect.Indirect(reflect.ValueOf(model))
	sf := make([]reflect.StructField, 0)
	for i := 0; i < modelType.NumField(); i++ {
		sf = append(sf, modelType.Field(i))
		field := modelType.Field(i)
		json := field.Tag.Get("json")
		s := strings.Split(json, ",")[0]
		if find(idJsonNames, s) == false {
			sf[i].Tag = `json:"-"`
		}
	}
	newType := reflect.StructOf(sf)
	newValue := value.Convert(newType)
	return reflect.TypeOf(newValue.Interface())
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
