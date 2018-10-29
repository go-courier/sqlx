package builder

// replace ? as some query snippet
//
// examples:
// ? => ST_GeomFromText(?)
//
type ValuerExpr interface {
	ValueEx() string
}

type DataTypeDescriber interface {
	DataType(engine string) string
}

type Model interface {
	TableName() string
}

type Indexes map[string][]string

type WithPrimaryKey interface {
	PrimaryKey() []string
}

type WithUniqueIndexes interface {
	UniqueIndexes() Indexes
}

type WithIndexes interface {
	Indexes() Indexes
}

type WithComments interface {
	Comments() map[string]string
}

type Dialect interface {
	DriverName() string
	BindVar(i int) string
	IsErrorUnknownDatabase(err error) bool
	IsErrorConflict(err error) bool
	CreateDatabaseIfNotExists(dbName string) SqlExpr
	DropDatabase(dbName string) SqlExpr
	CreateTableIsNotExists(t *Table) []SqlExpr
	DropTable(t *Table) SqlExpr
	TruncateTable(t *Table) SqlExpr
	AddColumn(col *Column) SqlExpr
	ModifyColumn(col *Column) SqlExpr
	DropColumn(col *Column) SqlExpr
	AddIndex(key *Key) SqlExpr
	DropIndex(key *Key) SqlExpr
	DataType(columnType *ColumnType) SqlExpr
}
