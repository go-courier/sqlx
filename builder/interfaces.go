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
	DataType(driverName string) string
}

type Model interface {
	TableName() string
}

type WithTableDescription interface {
	TableDescription() []string
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

type WithRelations interface {
	ColRelations() map[string][]string
}

type WithColDescriptions interface {
	ColDescriptions() map[string][]string
}

type Dialect interface {
	DriverName() string
	PrimaryKeyName() string
	IsErrorUnknownDatabase(err error) bool
	IsErrorConflict(err error) bool
	CreateDatabase(dbName string) SqlExpr
	CreateSchema(schemaName string) SqlExpr
	DropDatabase(dbName string) SqlExpr
	CreateTableIsNotExists(t *Table) []SqlExpr
	DropTable(t *Table) SqlExpr
	TruncateTable(t *Table) SqlExpr
	AddColumn(col *Column) SqlExpr
	RenameColumn(col *Column, target *Column) SqlExpr
	ModifyColumn(col *Column) SqlExpr
	DropColumn(col *Column) SqlExpr
	AddIndex(key *Key) SqlExpr
	DropIndex(key *Key) SqlExpr
	DataType(columnType *ColumnType) SqlExpr
}
