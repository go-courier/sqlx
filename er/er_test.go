package er_test

import (
	"encoding/json"
	"fmt"
	"github.com/go-courier/sqlx/v2/er"
	"github.com/go-courier/sqlx/v2/generator/__examples__/database"
	"github.com/go-courier/sqlx/v2/postgresqlconnector"
	"testing"
)

func TestDatabaseERFromDB(t *testing.T) {
	er := er.DatabaseERFromDB(database.DBTest, &postgresqlconnector.PostgreSQLConnector{})
	data, _ := json.MarshalIndent(er, "", "  ")

	fmt.Println(string(data))
}
