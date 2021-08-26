package scanner

import (
	"context"
	"testing"

	"github.com/go-courier/sqlx/v2/scanner/nullable"

	"github.com/DATA-DOG/go-sqlmock"
)

func BenchmarkScanStruct(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows([]string{"f_i", "f_s"})

	b.Run("Direct", func(b *testing.B) {
		target := &T2{}
		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		rows, _ := db.Query("SELECT f_i,f_s from t")

		for i := 0; i < b.N; i++ {
			mockRows.AddRow(i, "a")

			cs, _ := rows.Columns()

			if rows.Next() {
				dest := make([]interface{}, len(cs))
				crs := target.ColumnReceivers()
				for i := range cs {
					dest[i] = nullable.NewNullIgnoreScanner(crs[cs[i]])
				}
				_ = rows.Scan(dest...)
			}
		}

		b.Log(target)
	})

	b.Run("Scan by reflect", func(b *testing.B) {
		target := &T{}
		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		rows, _ := db.Query("SELECT f_i,f_s from t")

		for i := 0; i < b.N; i++ {
			mockRows.AddRow(i, "b")

			if rows.Next() {
				_ = scanTo(context.Background(), rows, target)
			}
		}

		b.Log(target)
	})

	b.Run("Scan by column receivers", func(b *testing.B) {
		target := &T2{}
		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		rows, _ := db.Query("SELECT f_i,f_s from t")

		for i := 0; i < b.N; i++ {
			mockRows.AddRow(i, "c")

			if rows.Next() {
				_ = scanTo(context.Background(), rows, target)
			}
		}

		b.Log(target)
	})
}
