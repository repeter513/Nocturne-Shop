// Package http implements REST handlers and middleware for the BFF.
// Пакет http реализует REST-обработчики и middleware для BFF.
package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// protoJSON marshals protobuf messages without unpopulated fields.
// protoJSON сериализует protobuf-сообщения без незаполненных полей.
var protoJSON = protojson.MarshalOptions{EmitUnpopulated: false}

// writeJSON writes a JSON or protobuf JSON response.
// writeJSON записывает JSON- или protobuf JSON-ответ.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Proto responses use protojson (int64 as strings, enum names, etc.).
	// Proto-ответы сериализуются через protojson (int64 как строки, имена enum и т.д.).
	if msg, ok := v.(proto.Message); ok {
		b, err := protoJSON.Marshal(msg)
		if err != nil {
			_, _ = w.Write([]byte(`{"error":"encode failed"}`))
			return
		}
		_, _ = w.Write(b)
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps gRPC errors to HTTP JSON error responses.
// writeError преобразует gRPC-ошибки в HTTP JSON-ответы об ошибке.
func writeError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, grpcCodeToHTTP(st.Code()), map[string]string{"error": st.Message()})
}

// grpcCodeToHTTP maps gRPC status codes to HTTP status codes.
// grpcCodeToHTTP сопоставляет коды статуса gRPC с HTTP-кодами.
func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.FailedPrecondition:
		return http.StatusConflict
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// queryInt reads an integer query parameter with a default fallback.
// queryInt читает целочисленный query-параметр с запасным значением.
func queryInt(r *http.Request, key string, def int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

// queryInt64 reads an int64 query parameter, returning 0 on failure.
// queryInt64 читает int64 query-параметр, возвращая 0 при ошибке.
func queryInt64(r *http.Request, key string) int64 {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// pathInt64 parses an int64 path parameter from the request.
// pathInt64 разбирает int64 path-параметр из запроса.
func pathInt64(r *http.Request, key string) (int64, error) {
	raw := r.PathValue(key)
	if raw == "" {
		return 0, status.Error(codes.InvalidArgument, key+" required")
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "invalid "+key)
	}
	return v, nil
}

// decodeJSON decodes the request body into dst.
// decodeJSON декодирует тело запроса в dst.
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return status.Error(codes.InvalidArgument, "empty body")
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return status.Error(codes.InvalidArgument, "invalid json")
	}
	return nil
}
