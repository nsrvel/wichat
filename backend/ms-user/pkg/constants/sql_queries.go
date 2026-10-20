package constants

// Raw SQL kept out of Bun builders (locks, FOR UPDATE, bulk deletes).

const (
	SQLLockUsers = `LOCK TABLE auth.users IN EXCLUSIVE MODE`

	SQLSelectUserByIdentity = `
		SELECT u.id, u.email, u.password_hash, u.display_name, u.created_at, u.updated_at
		FROM auth.identities i
		JOIN auth.users u ON u.id = i.user_id
		WHERE i.provider = ? AND i.subject = ?`

	SQLSelectRefreshUsedAtForUpdate = `
		SELECT used_at FROM auth.refresh_tokens WHERE token_hash = ? FOR UPDATE`

	SQLDeleteExpiredRefreshTokens = `
		DELETE FROM auth.refresh_tokens
		WHERE session_id IN (SELECT id FROM auth.sessions WHERE expires_at <= ?)`

	SQLDeleteExpiredSessions = `DELETE FROM auth.sessions WHERE expires_at <= ?`
)
