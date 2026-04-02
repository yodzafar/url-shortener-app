const (
	SessionCookie = "session_id",
	UserContextKey = "user"
)

type AuthMiddleware struct {
	store *session.store
}