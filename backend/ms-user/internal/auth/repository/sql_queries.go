package repository

// Advanced SQL kept out of Bun builders (locks, FOR UPDATE, joins, bulk deletes).

const (
	sqlLockUsers = `LOCK TABLE auth.users IN EXCLUSIVE MODE`

	sqlSelectRefreshUsedAtForUpdate = `
		SELECT used_at FROM auth.refresh_tokens WHERE token_hash = ? FOR UPDATE`

	sqlDeleteExpiredRefreshTokens = `
		DELETE FROM auth.refresh_tokens
		WHERE session_id IN (SELECT id FROM auth.sessions WHERE expires_at <= ?)`

	sqlDeleteExpiredSessions = `DELETE FROM auth.sessions WHERE expires_at <= ?`

	sqlSelectUserByIdentity = `
		SELECT u.id, u.email, u.password_hash, u.display_name, u.created_at, u.updated_at
		FROM auth.identities i
		JOIN auth.users u ON u.id = i.user_id
		WHERE i.provider = ? AND i.subject = ?`
)
