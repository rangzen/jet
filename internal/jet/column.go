// Modeling of columns

package jet

// Column is common column interface for all types of columns.
type Column interface {
	Name() string
	TableName() string

	setTableName(table string)
	setSubQuery(subQuery SelectTable)
	defaultAlias() string
}

// ColumnExpression interface
type ColumnExpression interface {
	Column
	Expression
}

// The base type for real materialized columns.
type columnImpl struct {
	expressionInterfaceImpl

	name      string
	tableName string

	subQuery SelectTable
}

func newColumn(name string, tableName string, parent ColumnExpression) columnImpl {
	bc := columnImpl{
		name:      name,
		tableName: tableName,
	}

	bc.expressionInterfaceImpl.Parent = parent

	return bc
}

func (c *columnImpl) Name() string {
	return c.name
}

func (c *columnImpl) TableName() string {
	return c.tableName
}

func (c *columnImpl) setTableName(table string) {
	c.tableName = table
}

func (c *columnImpl) setSubQuery(subQuery SelectTable) {
	c.subQuery = subQuery
}

func (c *columnImpl) defaultAlias() string {
	if c.tableName != "" {
		return c.tableName + "." + c.name
	}

	return c.name
}

func (c *columnImpl) serializeForOrderBy(statement StatementType, out *SQLBuilder) {
	if statement == SetStatementType {
		// set Statement (UNION, EXCEPT ...) can reference only select projections in order by clause
		out.WriteAlias(c.defaultAlias()) //always quote

		return
	}

	c.serialize(statement, out)
}

func (c columnImpl) serializeForProjection(statement StatementType, out *SQLBuilder) {
	c.serialize(statement, out)

	out.WriteString("AS")
	out.WriteAlias(c.defaultAlias())
}

func (c columnImpl) serialize(statement StatementType, out *SQLBuilder, options ...SerializeOption) {

	if c.subQuery != nil {
		out.WriteIdentifier(c.subQuery.Alias())
		out.WriteByte('.')
		out.WriteIdentifier(c.defaultAlias(), true)
	} else {
		if c.tableName != "" {
			out.WriteIdentifier(c.tableName)
			out.WriteByte('.')
		}

		out.WriteIdentifier(c.name)
	}
}

//------------------------------------------------------//

// IColumnList is used to store list of columns for later reuse as single projection or
// column list for UPDATE and INSERT statement.
type IColumnList interface {
	Projection
	Column

	columns() []ColumnExpression
}

// ColumnList function returns list of columns that be used as projection or column list for UPDATE and INSERT statement.
func ColumnList(columns ...ColumnExpression) IColumnList {
	return columnListImpl(columns)
}

// ColumnList is redefined type to support list of columns as single Projection
type columnListImpl []ColumnExpression

func (cl columnListImpl) columns() []ColumnExpression {
	return cl
}

func (cl columnListImpl) fromImpl(subQuery SelectTable) Projection {
	newProjectionList := ProjectionList{}

	for _, column := range cl {
		newProjectionList = append(newProjectionList, column.fromImpl(subQuery))
	}

	return newProjectionList
}

func (cl columnListImpl) serializeForProjection(statement StatementType, out *SQLBuilder) {
	projections := ColumnListToProjectionList(cl)

	SerializeProjectionList(statement, projections, out)
}

// dummy column interface implementation

// Name is placeholder for ColumnList to implement Column interface
func (cl columnListImpl) Name() string { return "" }

// TableName is placeholder for ColumnList to implement Column interface
func (cl columnListImpl) TableName() string                { return "" }
func (cl columnListImpl) setTableName(name string)         {}
func (cl columnListImpl) setSubQuery(subQuery SelectTable) {}
func (cl columnListImpl) defaultAlias() string             { return "" }