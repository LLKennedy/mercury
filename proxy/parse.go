package proxy

import (
	"encoding/json"
	"fmt"

	"github.com/LLKennedy/mercury/httpapi"
	"github.com/peterbourgon/mergemap"
)

func parseQuery(query map[string]*httpapi.MultiVal) map[string]interface{} {
	js := map[string]interface{}{}
	for key, value := range query {
		// TODO: don't ignore/overwrite duplicate keys here
		merged := ""
		for _, merged = range value.GetValues() {
		}
		parsed := parseQueryString(merged)
		js[key] = parsed
	}
	return js
}

func parseQueryString(part string) interface{} {
	js := map[string]interface{}{}
	err := json.Unmarshal([]byte(part), &js)
	if err == nil {
		for key, value := range js {
			strVal, ok := value.(string)
			if ok {
				js[key] = parseQueryString(strVal)
			}
		}
		return js
	}
	return part
}

func parseRequest(req *httpapi.Request) (finalJSON []byte, err error) {
	// First we convert query parameters to a map
	queryMap := parseQuery(req.GetParams())
	switch req.GetMethod() {
	case httpapi.Method_CONNECT, httpapi.Method_GET, httpapi.Method_HEAD, httpapi.Method_OPTIONS, httpapi.Method_TRACE:
		// No request body, only query params are possible
		if len(queryMap) > 0 {
			finalJSON, err = json.Marshal(queryMap)
			if err != nil {
				err = fmt.Errorf("failed to marshal query parameters to JSON: %v", err)
			}
		}
	case httpapi.Method_DELETE, httpapi.Method_PATCH, httpapi.Method_POST, httpapi.Method_PUT:
		// Merge request body with query params
		bodyJSON := req.GetPayload()
		if bodyJSON != nil && len(queryMap) > 0 {
			var bodyMap map[string]interface{}
			err = json.Unmarshal(bodyJSON, &bodyMap)
			if err != nil {
				err = fmt.Errorf("failed to unmarshall request body JSON: %v", err)
				break
			}
			// Merge both maps, using request body's values on conflict
			mergedMaps := mergemap.Merge(queryMap, bodyMap)
			finalJSON, err = json.Marshal(mergedMaps)
		} else if bodyJSON != nil {
			finalJSON = bodyJSON
		} else if len(queryMap) > 0 {
			finalJSON, err = json.Marshal(queryMap)
		}
	default:
		// Invalid http method
		// It shouldn't be possible to hit this normally, we do validation before we reach this point
		err = fmt.Errorf("invalid http method")
	}
	return
}

func methodToString(in httpapi.Method) (out string, err error) {
	switch in {
	case httpapi.Method_GET:
		out = "GET"
	case httpapi.Method_HEAD:
		out = "HEAD"
	case httpapi.Method_POST:
		out = "POST"
	case httpapi.Method_PUT:
		out = "PUT"
	case httpapi.Method_DELETE:
		out = "DELETE"
	case httpapi.Method_CONNECT:
		out = "CONNECT"
	case httpapi.Method_OPTIONS:
		out = "OPTIONS"
	case httpapi.Method_TRACE:
		out = "TRACE"
	case httpapi.Method_PATCH:
		out = "PATCH"
	}
	if out == "" {
		err = fmt.Errorf("unknown HTTP method")
	}
	return
}
