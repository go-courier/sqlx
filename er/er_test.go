package er_test

import (
	"encoding/json"

	"github.com/go-courier/sqlx/v2/er"
	"github.com/go-courier/sqlx/v2/generator/__examples__/database"
	"github.com/go-courier/sqlx/v2/postgresqlconnector"
)

func ExampleDatabaseERFromDB() {
	ers := er.DatabaseERFromDB(database.DBTest, &postgresqlconnector.PostgreSQLConnector{})
	_, _ = json.MarshalIndent(ers, "", "  ")
	// Output:
}
