package client

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FormatCommandError turns an error returned from a command into a clean,
// user-facing message. For gRPC errors it unwraps the underlying status so
// the user sees the actual server message (e.g. "failed to load node") and a
// short, actionable hint based on the status code instead of the raw
// "rpc error: code = Unknown desc = ..." wrapper.
func FormatCommandError(err error) string {
	if err == nil {
		return ""
	}

	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC status error, return as-is.
		return err.Error()
	}

	msg := st.Message()
	if hint := hintFor(st.Code(), msg); hint != "" {
		return fmt.Sprintf("%s\n%s", msg, hint)
	}
	return msg
}

// hintFor returns a short, actionable hint for an error. It inspects the
// message first (to catch cases the status code alone can not distinguish,
// e.g. a missing-role error returned as Unauthenticated) and falls back to a
// per-code hint. An empty string means no extra hint should be shown.
func hintFor(code codes.Code, msg string) string {
	lower := strings.ToLower(msg)

	// Missing roles/permissions can be reported by the backend as either
	// PermissionDenied or Unauthenticated. Detect it by message so we do not
	// wrongly tell the user their session expired.
	if strings.Contains(lower, "required role") || strings.Contains(lower, "permission") {
		return "Hint: your account is missing the roles required for this action. Ask an administrator to grant the listed roles."
	}

	switch code {
	case codes.NotFound:
		return "Hint: the resource was not found. Verify the IDs you passed and that the selected project (\"project use\" or --project-id) is the one that owns the resource."
	case codes.PermissionDenied:
		return "Hint: you do not have permission to access this resource."
	case codes.Unauthenticated:
		return "Hint: authentication failed. Your session may have expired, try logging in again."
	case codes.InvalidArgument:
		return "Hint: one or more arguments are invalid. Check the values passed to the command."
	default:
		return ""
	}
}
