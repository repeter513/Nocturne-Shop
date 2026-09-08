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

var protoJSON = protojson.MarshalOptions{EmitUnpopulated: false}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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

func writeError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, grpcCodeToHTTP(st.Code()), map[string]string{"error": st.Message()})
}

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
