package database

import (
	"fmt"
	"time"

	github_com_go_courier_sqlx "github.com/go-courier/sqlx"
	github_com_go_courier_sqlx_builder "github.com/go-courier/sqlx/builder"
	github_com_go_courier_sqlx_datatypes "github.com/go-courier/sqlx/datatypes"
)

var UserTable *github_com_go_courier_sqlx_builder.Table

func init() {
	UserTable = DBTest.Register(&User{})
}

func (User) TableName() string {
	return "t_user"
}

func (User) D() *github_com_go_courier_sqlx.Database {
	return DBTest
}

func (User) T() *github_com_go_courier_sqlx_builder.Table {
	return UserTable
}

func (User) FieldKeyID() string {
	return "ID"
}

func (m *User) FieldID() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyID())
}

func (User) FieldKeyName() string {
	return "Name"
}

func (m *User) FieldName() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyName())
}

func (User) FieldKeyUsername() string {
	return "Username"
}

func (m *User) FieldUsername() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyUsername())
}

func (User) FieldKeyNickname() string {
	return "Nickname"
}

func (m *User) FieldNickname() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyNickname())
}

func (User) FieldKeyGender() string {
	return "Gender"
}

func (m *User) FieldGender() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyGender())
}

func (User) FieldKeyBirthday() string {
	return "Birthday"
}

func (m *User) FieldBirthday() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyBirthday())
}

func (User) FieldKeyBoolean() string {
	return "Boolean"
}

func (m *User) FieldBoolean() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyBoolean())
}

func (User) FieldKeyCreatedAt() string {
	return "CreatedAt"
}

func (m *User) FieldCreatedAt() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyCreatedAt())
}

func (User) FieldKeyUpdatedAt() string {
	return "UpdatedAt"
}

func (m *User) FieldUpdatedAt() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyUpdatedAt())
}

func (User) FieldKeyEnabled() string {
	return "Enabled"
}

func (m *User) FieldEnabled() *github_com_go_courier_sqlx_builder.Column {
	return m.T().F(m.FieldKeyEnabled())
}

func (m *User) IndexFieldNames() []string {
	return []string{
		"ID",
		"Name",
		"Nickname",
		"Username",
	}
}

func (m *User) ConditionByStruct() *github_com_go_courier_sqlx_builder.Condition {
	table := m.T()
	fieldValues := github_com_go_courier_sqlx_builder.FieldValuesFromStructByNonZero(m)

	conditions := make([]*github_com_go_courier_sqlx_builder.Condition, 0)

	for _, fieldName := range m.IndexFieldNames() {
		if v, exists := fieldValues[fieldName]; exists {
			conditions = append(conditions, table.F(fieldName).Eq(v))
			delete(fieldValues, fieldName)
		}
	}

	if len(conditions) == 0 {
		panic(fmt.Errorf("at least one of field for indexes has value"))
	}

	for fieldName, v := range fieldValues {
		conditions = append(conditions, table.F(fieldName).Eq(v))
	}

	condition := github_com_go_courier_sqlx_builder.And(conditions...)

	condition = github_com_go_courier_sqlx_builder.And(condition, table.F("Enabled").Eq("github_com_go_courier_sqlx_datatypes.BOOL_TRUE"))
	return condition
}

func (User) PrimaryKey() github_com_go_courier_sqlx_builder.FieldNames {
	return github_com_go_courier_sqlx_builder.FieldNames{
		"ID",
	}
}

func (User) Indexes() github_com_go_courier_sqlx_builder.Indexes {
	return github_com_go_courier_sqlx_builder.Indexes{
		"I_nickname": github_com_go_courier_sqlx_builder.FieldNames{
			"Nickname",
		},
		"I_username": github_com_go_courier_sqlx_builder.FieldNames{
			"Username",
		},
	}
}

func (User) UniqueIndexes() github_com_go_courier_sqlx_builder.Indexes {
	return github_com_go_courier_sqlx_builder.Indexes{
		"I_name": github_com_go_courier_sqlx_builder.FieldNames{
			"Name",
			"Enabled",
		},
	}
}

func (User) Comments() map[string]string {
	return map[string]string{
		"Birthday":  "",
		"Boolean":   "",
		"CreatedAt": "",
		"Name":      "姓名",
		"Username":  "",
		"Nickname":  "",
		"Enabled":   "",
		"ID":        "",
		"Gender":    "",
		"UpdatedAt": "",
	}
}

func (m *User) Create(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	if m.CreatedAt.IsZero() {
		m.CreatedAt = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	d := m.D()

	result, err := db.ExecExpr(github_com_go_courier_sqlx_builder.Insert().
		Into(m.T(), github_com_go_courier_sqlx_builder.Comment("User.Create")).
		Set(d.Assignments(m)...))

	if err == nil {
		lastInsertID, _ := result.LastInsertId()
		m.ID = uint64(lastInsertID)
	}

	return err
}

func (m *User) CreateOnDuplicateWithUpdateFields(db *github_com_go_courier_sqlx.DB, updateFields []string) error {

	if len(updateFields) == 0 {
		panic(fmt.Errorf("must have update fields"))
	}

	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	if m.CreatedAt.IsZero() {
		m.CreatedAt = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	fieldValues := github_com_go_courier_sqlx_builder.FieldValuesFromStructByNonZero(m, updateFields...)

	delete(fieldValues, "ID")

	table := m.T()

	cols, vals := table.ColumnsAndValuesByFieldValues(fieldValues)

	fields := make(map[string]bool, len(updateFields))
	for _, field := range updateFields {
		fields[field] = true
	}

	for _, fieldNames := range m.UniqueIndexes() {
		for _, field := range fieldNames {
			delete(fields, field)
		}
	}

	if len(fields) == 0 {
		panic(fmt.Errorf("no fields for updates"))
	}

	for field := range fieldValues {
		if !fields[field] {
			delete(fieldValues, field)
		}
	}

	_, err := db.ExecExpr(github_com_go_courier_sqlx_builder.Insert().
		Into(
			m.T(),
			github_com_go_courier_sqlx_builder.OnDuplicateKeyUpdate(table.AssignmentsByFieldValues(fieldValues)...),
			github_com_go_courier_sqlx_builder.Comment("User.CreateOnDuplicateWithUpdateFields"),
		).
		Values(cols, vals...),
	)

	return err
}

func (m *User) DeleteByStruct(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	if m.CreatedAt.IsZero() {
		m.CreatedAt = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	_, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Delete().
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(m.ConditionByStruct()),
				github_com_go_courier_sqlx_builder.Comment("User.DeleteByStruct"),
			),
	)

	return err
}

func (m *User) FetchByID(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	err := db.QueryExprAndScan(
		github_com_go_courier_sqlx_builder.Select(nil).
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(github_com_go_courier_sqlx_builder.And(
					table.F("ID").Eq(m.ID),
					table.F("Enabled").Eq(m.Enabled),
				)),
				github_com_go_courier_sqlx_builder.Comment("User.FetchByID"),
			),
		m,
	)

	return err
}

func (m *User) UpdateByIDWithMap(db *github_com_go_courier_sqlx.DB, fieldValues github_com_go_courier_sqlx_builder.FieldValues) error {

	if _, ok := fieldValues["UpdatedAt"]; !ok {
		fieldValues["UpdatedAt"] = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	result, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Update(m.T()).
			Where(
				github_com_go_courier_sqlx_builder.And(
					table.F("ID").Eq(m.ID),
					table.F("Enabled").Eq(m.Enabled),
				),
				github_com_go_courier_sqlx_builder.Comment("User.UpdateByIDWithMap"),
			).
			Set(table.AssignmentsByFieldValues(fieldValues)...),
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return m.FetchByID(db)
	}

	return nil

}

func (m *User) UpdateByIDWithStruct(db *github_com_go_courier_sqlx.DB, zeroFields ...string) error {

	fieldValues := github_com_go_courier_sqlx_builder.FieldValuesFromStructByNonZero(m, zeroFields...)
	return m.UpdateByIDWithMap(db, fieldValues)

}

func (m *User) FetchByIDForUpdate(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	err := db.QueryExprAndScan(
		github_com_go_courier_sqlx_builder.Select(nil).
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(github_com_go_courier_sqlx_builder.And(
					table.F("ID").Eq(m.ID),
					table.F("Enabled").Eq(m.Enabled),
				)),
				github_com_go_courier_sqlx_builder.ForUpdate(),
				github_com_go_courier_sqlx_builder.Comment("User.FetchByIDForUpdate"),
			),
		m,
	)

	return err
}

func (m *User) DeleteByID(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	_, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Delete().
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(github_com_go_courier_sqlx_builder.And(
					table.F("ID").Eq(m.ID),
					table.F("Enabled").Eq(m.Enabled),
				)),
				github_com_go_courier_sqlx_builder.Comment("User.DeleteByID"),
			),
	)

	return err
}

func (m *User) SoftDeleteByID(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	fieldValues := github_com_go_courier_sqlx_builder.FieldValues{}
	if _, ok := fieldValues["Enabled"]; !ok {
		fieldValues["Enabled"] = github_com_go_courier_sqlx_datatypes.BOOL_FALSE
	}

	if _, ok := fieldValues["UpdatedAt"]; !ok {
		fieldValues["UpdatedAt"] = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	_, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Update(m.T()).
			Where(
				github_com_go_courier_sqlx_builder.And(
					table.F("ID").Eq(m.ID),
					table.F("Enabled").Eq(m.Enabled),
				),
				github_com_go_courier_sqlx_builder.Comment("User.SoftDeleteByID"),
			).
			Set(table.AssignmentsByFieldValues(fieldValues)...),
	)

	if err != nil {
		dbErr := github_com_go_courier_sqlx.DBErr(err)
		if dbErr.IsConflict() {
			return m.DeleteByID(db)
		}
	}

	return nil

}

func (m *User) FetchByName(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	err := db.QueryExprAndScan(
		github_com_go_courier_sqlx_builder.Select(nil).
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(github_com_go_courier_sqlx_builder.And(
					table.F("Name").Eq(m.Name),
					table.F("Enabled").Eq(m.Enabled),
				)),
				github_com_go_courier_sqlx_builder.Comment("User.FetchByName"),
			),
		m,
	)

	return err
}

func (m *User) UpdateByNameWithMap(db *github_com_go_courier_sqlx.DB, fieldValues github_com_go_courier_sqlx_builder.FieldValues) error {

	if _, ok := fieldValues["UpdatedAt"]; !ok {
		fieldValues["UpdatedAt"] = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	result, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Update(m.T()).
			Where(
				github_com_go_courier_sqlx_builder.And(
					table.F("Name").Eq(m.Name),
					table.F("Enabled").Eq(m.Enabled),
				),
				github_com_go_courier_sqlx_builder.Comment("User.UpdateByNameWithMap"),
			).
			Set(table.AssignmentsByFieldValues(fieldValues)...),
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return m.FetchByName(db)
	}

	return nil

}

func (m *User) UpdateByNameWithStruct(db *github_com_go_courier_sqlx.DB, zeroFields ...string) error {

	fieldValues := github_com_go_courier_sqlx_builder.FieldValuesFromStructByNonZero(m, zeroFields...)
	return m.UpdateByNameWithMap(db, fieldValues)

}

func (m *User) FetchByNameForUpdate(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	err := db.QueryExprAndScan(
		github_com_go_courier_sqlx_builder.Select(nil).
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(github_com_go_courier_sqlx_builder.And(
					table.F("Name").Eq(m.Name),
					table.F("Enabled").Eq(m.Enabled),
				)),
				github_com_go_courier_sqlx_builder.ForUpdate(),
				github_com_go_courier_sqlx_builder.Comment("User.FetchByNameForUpdate"),
			),
		m,
	)

	return err
}

func (m *User) DeleteByName(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	_, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Delete().
			From(
				m.T(),
				github_com_go_courier_sqlx_builder.Where(github_com_go_courier_sqlx_builder.And(
					table.F("Name").Eq(m.Name),
					table.F("Enabled").Eq(m.Enabled),
				)),
				github_com_go_courier_sqlx_builder.Comment("User.DeleteByName"),
			),
	)

	return err
}

func (m *User) SoftDeleteByName(db *github_com_go_courier_sqlx.DB) error {
	m.Enabled = github_com_go_courier_sqlx_datatypes.BOOL_TRUE

	table := m.T()

	fieldValues := github_com_go_courier_sqlx_builder.FieldValues{}
	if _, ok := fieldValues["Enabled"]; !ok {
		fieldValues["Enabled"] = github_com_go_courier_sqlx_datatypes.BOOL_FALSE
	}

	if _, ok := fieldValues["UpdatedAt"]; !ok {
		fieldValues["UpdatedAt"] = github_com_go_courier_sqlx_datatypes.MySQLTimestamp(time.Now())
	}

	_, err := db.ExecExpr(
		github_com_go_courier_sqlx_builder.Update(m.T()).
			Where(
				github_com_go_courier_sqlx_builder.And(
					table.F("Name").Eq(m.Name),
					table.F("Enabled").Eq(m.Enabled),
				),
				github_com_go_courier_sqlx_builder.Comment("User.SoftDeleteByName"),
			).
			Set(table.AssignmentsByFieldValues(fieldValues)...),
	)

	if err != nil {
		dbErr := github_com_go_courier_sqlx.DBErr(err)
		if dbErr.IsConflict() {
			return m.DeleteByName(db)
		}
	}

	return nil

}

func (m *User) List(db *github_com_go_courier_sqlx.DB, condition *github_com_go_courier_sqlx_builder.Condition, additions ...github_com_go_courier_sqlx_builder.Addition) ([]User, error) {

	list := make([]User, 0)

	table := m.T()

	condition = github_com_go_courier_sqlx_builder.And(condition, table.F("Enabled").Eq(github_com_go_courier_sqlx_datatypes.BOOL_TRUE))

	finalAdditions := []github_com_go_courier_sqlx_builder.Addition{
		github_com_go_courier_sqlx_builder.Where(condition),
		github_com_go_courier_sqlx_builder.Comment("User.List"),
	}

	if len(additions) > 0 {
		finalAdditions = append(finalAdditions, additions...)
	}

	err := db.QueryExprAndScan(
		github_com_go_courier_sqlx_builder.Select(nil).
			From(m.T(), finalAdditions...),
		&list,
	)

	return list, err

}

func (m *User) Count(db *github_com_go_courier_sqlx.DB, condition *github_com_go_courier_sqlx_builder.Condition, additions ...github_com_go_courier_sqlx_builder.Addition) (int, error) {

	count := -1

	table := m.T()

	condition = github_com_go_courier_sqlx_builder.And(condition, table.F("Enabled").Eq(github_com_go_courier_sqlx_datatypes.BOOL_TRUE))

	finalAdditions := []github_com_go_courier_sqlx_builder.Addition{
		github_com_go_courier_sqlx_builder.Where(condition),
		github_com_go_courier_sqlx_builder.Comment("User.Count"),
	}

	if len(additions) > 0 {
		finalAdditions = append(finalAdditions, additions...)
	}

	err := db.QueryExprAndScan(
		github_com_go_courier_sqlx_builder.Select(
			github_com_go_courier_sqlx_builder.Count(),
		).
			From(m.T(), finalAdditions...),
		&count,
	)

	return count, err

}

func (m *User) BatchFetchByIDList(db *github_com_go_courier_sqlx.DB, values []string) ([]User, error) {

	if len(values) == 0 {
		return nil, nil
	}

	table := m.T()

	condition := table.F("ID").In(values)

	return m.List(db, condition)

}

func (m *User) BatchFetchByNameList(db *github_com_go_courier_sqlx.DB, values []string) ([]User, error) {

	if len(values) == 0 {
		return nil, nil
	}

	table := m.T()

	condition := table.F("Name").In(values)

	return m.List(db, condition)

}

func (m *User) BatchFetchByNicknameList(db *github_com_go_courier_sqlx.DB, values []string) ([]User, error) {

	if len(values) == 0 {
		return nil, nil
	}

	table := m.T()

	condition := table.F("Nickname").In(values)

	return m.List(db, condition)

}

func (m *User) BatchFetchByUsernameList(db *github_com_go_courier_sqlx.DB, values []string) ([]User, error) {

	if len(values) == 0 {
		return nil, nil
	}

	table := m.T()

	condition := table.F("Username").In(values)

	return m.List(db, condition)

}
