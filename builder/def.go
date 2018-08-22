package builder

type Model interface {
	TableName() string
}

type FieldNames []string

func (fieldNames FieldNames) Map() map[string]bool {
	m := make(map[string]bool, len(fieldNames))
	for _, fieldName := range fieldNames {
		m[fieldName] = true
	}
	return m
}

type Indexes map[string]FieldNames

type IndexDeclarer func() (indexName string, fieldNames FieldNames)

type WithPrimaryKey interface {
	PrimaryKey() FieldNames
}

type WithIndexes interface {
	Indexes() Indexes
}

type WithUniqueIndexes interface {
	UniqueIndexes() Indexes
}

type WithComments interface {
	Comments() map[string]string
}
