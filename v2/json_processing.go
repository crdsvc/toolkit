package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
)

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1024 * 1024
	if t.MaxJSONBytes != 0 {
		maxBytes = t.MaxJSONBytes
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	if !t.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(data)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains malformed JSON (at char %d)", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type at char %d", unmarshalTypeError.Offset)
		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("error unmarshaling JSON %s", err.Error())
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains malformed JSON")
		case errors.Is(err, io.EOF):
			return errors.New("this is an empty JSON")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown field %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("more than one json in a file %w", err)
	}

	return nil

}

func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	json_bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		// for k, v := range headers[0] {
		// 	w.Header()[k] = v
		// }
		maps.Copy(w.Header(), headers[0])
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(json_bytes)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusInternalServerError
	if len(status) > 0 {
		statusCode = status[0]
	}
	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return t.WriteJSON(w, statusCode, payload)
}

func (t *Tools) PushJSONTORemote(uri string, data any, client ...*http.Client) (*http.Response, int, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, 0, err
	}

	httpClient := &http.Client{}
	if len(client) > 0 {
		httpClient = client[0]
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, 0, err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	return res, res.StatusCode, nil
}
