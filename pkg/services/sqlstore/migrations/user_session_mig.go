package migrations

import (
	. "github.com/grafana/grafana/pkg/services/sqlstore/migrator"
)

func addUserSessionMigrations(mg *Migrator) {
	userSessionV1 := Table{
		Name: "user_session",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "user_id", Type: DB_BigInt, Nullable: false},
			{Name: "auth_token", Type: DB_NVarchar, Length: 100, Nullable: false},
			{Name: "prev_auth_token", Type: DB_NVarchar, Length: 100, Nullable: false},
			{Name: "user_agent", Type: DB_NVarchar, Length: 255, Nullable: false},
			{Name: "client_ip", Type: DB_NVarchar, Length: 255, Nullable: false},
			{Name: "auth_token_seen", Type: DB_Bool, Nullable: false},
			{Name: "seen_at", Type: DB_Int, Nullable: true},
			{Name: "rotated_at", Type: DB_Int, Nullable: false},
			{Name: "created_at", Type: DB_Int, Nullable: false},
			{Name: "updated_at", Type: DB_Int, Nullable: false},
		},
		Indices: []*Index{
			{Cols: []string{"auth_token"}, Type: UniqueIndex},
			{Cols: []string{"prev_auth_token"}, Type: UniqueIndex},
		},
	}

	mg.AddMigration("create user session table", NewAddTableMigration(userSessionV1))
	mg.AddMigration("add unique index user_session.auth_token", NewAddIndexMigration(userSessionV1, userSessionV1.Indices[0]))
	mg.AddMigration("add unique index user_session.prev_auth_token", NewAddIndexMigration(userSessionV1, userSessionV1.Indices[1]))
}
