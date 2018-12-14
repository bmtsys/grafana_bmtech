package migrations

import (
	. "github.com/grafana/grafana/pkg/services/sqlstore/migrator"
)

func addUserSessionMigrations(mg *Migrator) {
	userSessionV1 := Table{
		Name: "user_session",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "session_id", Type: DB_NVarchar, Length: 100, Nullable: false},
			{Name: "user_id", Type: DB_BigInt, Nullable: false},
			{Name: "user_agent", Type: DB_NVarchar, Length: 255, Nullable: false},
			{Name: "client_ip", Type: DB_NVarchar, Length: 255, Nullable: false},
			{Name: "refreshed_at", Type: DB_Int, Nullable: false},
			{Name: "created_at", Type: DB_Int, Nullable: false},
			{Name: "updated_at", Type: DB_Int, Nullable: false},
		},
		Indices: []*Index{
			{Cols: []string{"session_id"}, Type: UniqueIndex},
			{Cols: []string{"session_id", "user_id"}},
		},
	}

	mg.AddMigration("create user session table", NewAddTableMigration(userSessionV1))
	mg.AddMigration("add unique index user_session.session_id", NewAddIndexMigration(userSessionV1, userSessionV1.Indices[0]))
	mg.AddMigration("add index user_session.session_id_user_id", NewAddIndexMigration(userSessionV1, userSessionV1.Indices[1]))
}
