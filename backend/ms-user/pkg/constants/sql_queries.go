package constants

// Raw SQL kept out of Bun builders (locks, FOR UPDATE, bulk deletes).

const (
	SQLLockUsers = `LOCK TABLE users IN EXCLUSIVE MODE`

	SQLSelectUserByIdentity = `
		SELECT u.id, u.email, u.username, u.display_name, u.birth_date,
		       u.is_guest, u.is_instance_owner, u.created_at, u.updated_at,
		       COALESCE(c.password_hash, '') AS password_hash
		FROM user_connections i
		JOIN users u ON u.id = i.user_id
		LEFT JOIN user_credentials c ON c.user_id = u.id
		WHERE i.kind = 'login' AND i.provider = ? AND i.subject = ?`

	ConnectionKindLogin = "login"

	SQLListGroupsForUser = `
		SELECT g.id AS g_id, g.public_id AS g_public_id, g.created_at AS g_created_at, g.updated_at AS g_updated_at,
		       s.group_id AS s_group_id, s.name AS s_name, s.logo_url AS s_logo_url,
		       s.description AS s_description, s.is_private AS s_is_private,
		       s.banner_color AS s_banner_color, s.banner_url AS s_banner_url,
		       s.retention_days AS s_retention_days, s.trash_retention_days AS s_trash_retention_days,
		       s.created_at AS s_created_at, s.updated_at AS s_updated_at,
		       ch.public_id AS ch_public_id,
		       m.role AS member_role
		FROM groups AS g
		JOIN group_settings AS s ON s.group_id = g.id
		JOIN group_members AS m ON m.group_id = g.id
		LEFT JOIN channels AS ch ON ch.group_id = g.id AND lower(ch.name) = 'general'
		WHERE m.user_id = ?
		ORDER BY m.rail_position ASC NULLS LAST, s.name ASC`

	SQLListJoinableGroups = `
		SELECT g.id AS g_id, g.public_id AS g_public_id, g.created_at AS g_created_at, g.updated_at AS g_updated_at,
		       s.group_id AS s_group_id, s.name AS s_name, s.logo_url AS s_logo_url,
		       s.description AS s_description, s.is_private AS s_is_private,
		       s.banner_color AS s_banner_color, s.banner_url AS s_banner_url,
		       s.retention_days AS s_retention_days, s.trash_retention_days AS s_trash_retention_days,
		       s.created_at AS s_created_at, s.updated_at AS s_updated_at,
		       ch.public_id AS ch_public_id
		FROM groups AS g
		JOIN group_settings AS s ON s.group_id = g.id
		LEFT JOIN channels AS ch ON ch.group_id = g.id AND lower(ch.name) = 'general'
		WHERE g.id NOT IN (SELECT group_id FROM group_members WHERE user_id = ?)
		ORDER BY s.name ASC`
)
