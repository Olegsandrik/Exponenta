package dto

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	internalErrors "github.com/Olegsandrik/Exponenta/internal/internalerrors"
)

func GetIntQueryParam(r *http.Request, name string) (int, error) {
	paramStr := r.URL.Query().Get(name)
	if paramStr == "" {
		return 0, internalErrors.ErrParamNotFound
	}

	param, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, internalErrors.ErrParamNotInteger
	}
	return param, nil
}

func GetIntURLParam(r *http.Request, name string) (int, error) {
	paramStr, ok := mux.Vars(r)[name]
	if !ok {
		return 0, internalErrors.ErrParamNotFound
	}

	param, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, internalErrors.ErrParamNotInteger
	}

	return param, nil
}

func GetStringQueryParam(r *http.Request, name string) (string, error) {
	dishTypeParam := r.URL.Query().Get(name)

	if dishTypeParam == "" {
		return "", internalErrors.ErrParamNotFound
	}
	return dishTypeParam, nil
}
