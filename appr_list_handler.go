package diff

import (
	"context"
	"net/http"
	"reflect"
)

type ApprListHandler struct {
	Error           func(context.Context, string)
	ApprListService ApprListService
	ModelType       reflect.Type
	IdNames         []string
	Log             func(ctx context.Context, resource string, action string, success bool, desc string) error
	Resource        string
	Action1         string
	Action2         string
}

func NewApprListHandler(apprListService ApprListService, modelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *ApprListHandler {
	return NewApprListHandlerWithKeys(apprListService, modelType, logError, nil, writeLog, options...)
}

func NewApprListHandlerWithKeys(apprListService ApprListService, modelType reflect.Type, logError func(context.Context, string), idNames []string, writeLog func(context.Context, string, string, bool, string) error, options ...string) *ApprListHandler {
	if len(idNames) == 0 {
		idNames = getJsonPrimaryKeys(modelType)
	}
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
	return &ApprListHandler{ApprListService: apprListService, ModelType: modelType, IdNames: idNames, Resource: resource, Error: logError, Log: writeLog, Action1: action1, Action2: action2}
}

func (c *ApprListHandler) Approve(w http.ResponseWriter, r *http.Request) {
	ids, err := buildIds(r, c.ModelType, c.IdNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		result, err := c.ApprListService.Approve(r.Context(), ids)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, c.Action1, err, c.Log)
		} else {
			succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action1)
		}
	}
}

func (c *ApprListHandler) Reject(w http.ResponseWriter, r *http.Request) {
	ids, err := buildIds(r, c.ModelType, c.IdNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		result, err := c.ApprListService.Reject(r.Context(), ids)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, c.Action2, err, c.Log)
		} else {
			succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action2)
		}
	}
}
