package diff

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type ApprHandler struct {
	Error       func(context.Context, string)
	ApprService ApprService
	ModelType   reflect.Type
	IdNames     []string
	Indexes     map[string]int
	Offset      int
	Log         func(ctx context.Context, resource string, action string, success bool, desc string) error
	Resource    string
	Action1     string
	Action2     string
}

func NewApprHandler(apprService ApprService, modelType reflect.Type, logError func(context.Context, string), option ...int) *ApprHandler {
	offset := 1
	if len(option) > 0 && option[0] >= 0 {
		offset = option[0]
	}
	return NewApprHandlerWithKeysAndLog(apprService, modelType, offset, logError, nil, nil)
}
func NewApprHandlerWithLogs(apprService ApprService, modelType reflect.Type, offset int, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *ApprHandler {
	return NewApprHandlerWithKeysAndLog(apprService, modelType, offset, logError, nil, writeLog, options...)
}
func NewApprHandlerWithKeys(apprService ApprService, modelType reflect.Type, logError func(context.Context, string), idNames []string, option ...int) *ApprHandler {
	offset := 1
	if len(option) > 0 && option[0] >= 0 {
		offset = option[0]
	}
	return NewApprHandlerWithKeysAndLog(apprService, modelType, offset, logError, idNames, nil)
}
func NewApprHandlerWithKeysAndLog(apprService ApprService, modelType reflect.Type, offset int, logError func(context.Context, string), idNames []string, writeLog func(context.Context, string, string, bool, string) error, options ...string) *ApprHandler {
	if offset < 0 {
		offset = 1
	}
	if idNames == nil || len(idNames) == 0 {
		idNames = getJsonPrimaryKeys(modelType)
	}
	indexes := getIndexes(modelType)
	var resource, action1, action2 string
	if len(options) > 0 && len(options[0]) > 0 {
		action1 = options[0]
	} else {
		action1 = "approve"
	}
	if len(options) > 1 && len(options[1]) > 0 {
		action2 = options[1]
	} else {
		action2 = "reject"
	}
	if len(options) > 2 && len(options[2]) > 0 {
		resource = options[2]
	} else {
		resource = buildResourceName(modelType.Name())
	}
	return &ApprHandler{Log: writeLog, ApprService: apprService, ModelType: modelType, IdNames: idNames, Indexes: indexes, Offset: offset, Error: logError, Resource: resource, Action1: action1, Action2: action2}
}

func (c *ApprHandler) newModel(body interface{}) (out interface{}) {
	req := reflect.New(c.ModelType).Interface()
	if body != nil {
		switch s := body.(type) {
		case io.Reader:
			err := json.NewDecoder(s).Decode(&req)
			if err != nil {
				return err
			}
			return req
		}
	}
	return req
}

func (c *ApprHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := buildId(r, c.ModelType, c.IdNames, c.Indexes, c.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		result, err := c.ApprService.Approve(r.Context(), id)
		if err != nil {
			handleError(w, r, http.StatusOK, internalServerError, c.Error, c.Resource, c.Action1, err, c.Log)
		} else {
			succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action1)
		}
	}
}

func (c *ApprHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := buildId(r, c.ModelType, c.IdNames, c.Indexes, c.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		result, err := c.ApprService.Reject(r.Context(), id)
		if err != nil {
			handleError(w, r, http.StatusOK, internalServerError, c.Error, c.Resource, c.Action2, err, c.Log)
		} else {
			succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action2)
		}
	}
}
