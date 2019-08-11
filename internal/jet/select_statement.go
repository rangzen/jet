package jet

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-jet/jet/execution"
)

// SelectStatement is interface for SQL SELECT statements
type SelectStatement interface {
	Statement
	IExpression

	DISTINCT() SelectStatement
	FROM(table ReadableTable) SelectStatement
	WHERE(expression BoolExpression) SelectStatement
	GROUP_BY(groupByClauses ...GroupByClause) SelectStatement
	HAVING(boolExpression BoolExpression) SelectStatement
	ORDER_BY(orderByClauses ...OrderByClause) SelectStatement
	LIMIT(limit int64) SelectStatement
	OFFSET(offset int64) SelectStatement
	FOR(lock SelectLock) SelectStatement

	UNION(rhs SelectStatement) SelectStatement
	UNION_ALL(rhs SelectStatement) SelectStatement
	INTERSECT(rhs SelectStatement) SelectStatement
	INTERSECT_ALL(rhs SelectStatement) SelectStatement
	EXCEPT(rhs SelectStatement) SelectStatement
	EXCEPT_ALL(rhs SelectStatement) SelectStatement

	AsTable(alias string) SelectTable

	projections() []Projection
}

//SELECT creates new SelectStatement with list of projections
func SELECT(projection1 Projection, projections ...Projection) SelectStatement {
	return newSelectStatement(nil, append([]Projection{projection1}, projections...))
}

type selectStatementImpl struct {
	ExpressionInterfaceImpl
	parent SelectStatement

	table          ReadableTable
	distinct       bool
	projectionList []Projection
	where          BoolExpression
	groupBy        []GroupByClause
	having         BoolExpression

	orderBy       []OrderByClause
	limit, offset int64

	lockFor SelectLock
}

func newSelectStatement(table ReadableTable, projections []Projection) SelectStatement {
	newSelect := &selectStatementImpl{
		table:          table,
		projectionList: projections,
		limit:          -1,
		offset:         -1,
		distinct:       false,
	}

	newSelect.ExpressionInterfaceImpl.Parent = newSelect
	newSelect.parent = newSelect

	return newSelect
}

func (s *selectStatementImpl) FROM(table ReadableTable) SelectStatement {
	s.table = table
	return s.parent
}

func (s *selectStatementImpl) AsTable(alias string) SelectTable {
	//return newSelectTable(s.parent, alias)
	panic("UNimplemented.")
}

func (s *selectStatementImpl) WHERE(expression BoolExpression) SelectStatement {
	s.where = expression
	return s.parent
}

func (s *selectStatementImpl) GROUP_BY(groupByClauses ...GroupByClause) SelectStatement {
	s.groupBy = groupByClauses
	return s.parent
}

func (s *selectStatementImpl) HAVING(expression BoolExpression) SelectStatement {
	s.having = expression
	return s.parent
}

func (s *selectStatementImpl) ORDER_BY(clauses ...OrderByClause) SelectStatement {
	s.orderBy = clauses
	return s.parent
}

func (s *selectStatementImpl) OFFSET(offset int64) SelectStatement {
	s.offset = offset
	return s.parent
}

func (s *selectStatementImpl) LIMIT(limit int64) SelectStatement {
	s.limit = limit
	return s.parent
}

func (s *selectStatementImpl) DISTINCT() SelectStatement {
	s.distinct = true
	return s.parent
}

func (s *selectStatementImpl) FOR(lock SelectLock) SelectStatement {
	s.lockFor = lock
	return s.parent
}

func (s *selectStatementImpl) UNION(rhs SelectStatement) SelectStatement {
	return UNION(s.parent, rhs)
}

func (s *selectStatementImpl) UNION_ALL(rhs SelectStatement) SelectStatement {
	return UNION_ALL(s.parent, rhs)
}

func (s *selectStatementImpl) INTERSECT(rhs SelectStatement) SelectStatement {
	return INTERSECT(s.parent, rhs)
}

func (s *selectStatementImpl) INTERSECT_ALL(rhs SelectStatement) SelectStatement {
	return INTERSECT_ALL(s.parent, rhs)
}

func (s *selectStatementImpl) EXCEPT(rhs SelectStatement) SelectStatement {
	return EXCEPT(s.parent, rhs)
}

func (s *selectStatementImpl) EXCEPT_ALL(rhs SelectStatement) SelectStatement {
	return EXCEPT_ALL(s.parent, rhs)
}

func (s *selectStatementImpl) projections() []Projection {
	return s.projectionList
}

func (s *selectStatementImpl) serialize(statement StatementType, out *SqlBuilder, options ...SerializeOption) error {
	if s == nil {
		return errors.New("jet: Select expression is nil. ")
	}
	out.WriteString("(")

	out.increaseIdent()
	err := s.serializeImpl(out)
	out.decreaseIdent()

	if err != nil {
		return err
	}

	out.NewLine()
	out.WriteString(")")

	return nil
}

func (s *selectStatementImpl) serializeImpl(out *SqlBuilder) error {
	if s == nil {
		return errors.New("jet: Select expression is nil. ")
	}

	out.NewLine()
	out.WriteString("SELECT")

	if s.distinct {
		out.WriteString("DISTINCT")
	}

	if len(s.projectionList) == 0 {
		return errors.New("jet: no column selected for Projection")
	}

	err := out.writeProjections(SelectStatementType, s.projectionList)

	if err != nil {
		return err
	}

	if s.table != nil {
		if err := out.writeFrom(SelectStatementType, s.table); err != nil {
			return err
		}
	}

	if s.where != nil {
		err := out.writeWhere(SelectStatementType, s.where)

		if err != nil {
			return nil
		}
	}

	if s.groupBy != nil && len(s.groupBy) > 0 {
		err := out.writeGroupBy(SelectStatementType, s.groupBy)

		if err != nil {
			return err
		}
	}

	if s.having != nil {
		err := out.writeHaving(SelectStatementType, s.having)

		if err != nil {
			return err
		}
	}

	if s.orderBy != nil {
		err := out.writeOrderBy(SelectStatementType, s.orderBy)

		if err != nil {
			return err
		}
	}

	if s.limit >= 0 {
		out.NewLine()
		out.WriteString("LIMIT")
		out.insertParametrizedArgument(s.limit)
	}

	if s.offset >= 0 {
		out.NewLine()
		out.WriteString("OFFSET")
		out.insertParametrizedArgument(s.offset)
	}

	if s.lockFor != nil {
		out.NewLine()
		out.WriteString("FOR")
		err := s.lockFor.serialize(SelectStatementType, out)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *selectStatementImpl) accept(visitor visitor) {
	visitor.visit(s)

	if s.table != nil {
		s.table.accept(visitor)
	}

	if s.where != nil {
		s.where.accept(visitor)
	}

	if s.having != nil {
		s.having.accept(visitor)
	}
}

func (s *selectStatementImpl) Sql(dialect ...Dialect) (query string, args []interface{}, err error) {

	queryData := &SqlBuilder{
		Dialect: detectDialect(s, dialect...),
	}

	err = s.serializeImpl(queryData)

	if err != nil {
		return "", nil, err
	}

	query, args = queryData.finalize()

	return
}

func (s *selectStatementImpl) DebugSql(dialect ...Dialect) (query string, err error) {
	return debugSql(s.parent, dialect...)
}

func (s *selectStatementImpl) Query(db execution.DB, destination interface{}) error {
	return query(s.parent, db, destination)
}

func (s *selectStatementImpl) QueryContext(context context.Context, db execution.DB, destination interface{}) error {
	return queryContext(context, s.parent, db, destination)
}

func (s *selectStatementImpl) Exec(db execution.DB) (res sql.Result, err error) {
	return exec(s.parent, db)
}

func (s *selectStatementImpl) ExecContext(context context.Context, db execution.DB) (res sql.Result, err error) {
	return execContext(context, s.parent, db)
}

// SelectLock is interface for SELECT statement locks
type SelectLock interface {
	Serializer

	NOWAIT() SelectLock
	SKIP_LOCKED() SelectLock
}

type selectLockImpl struct {
	lockStrength       string
	noWait, skipLocked bool
}

func NewSelectLock(name string) func() SelectLock {
	return func() SelectLock {
		return newSelectLock(name)
	}
}

func newSelectLock(lockStrength string) SelectLock {
	return &selectLockImpl{lockStrength: lockStrength}
}

func (s *selectLockImpl) NOWAIT() SelectLock {
	s.noWait = true
	return s
}

func (s *selectLockImpl) SKIP_LOCKED() SelectLock {
	s.skipLocked = true
	return s
}

func (s *selectLockImpl) serialize(statement StatementType, out *SqlBuilder, options ...SerializeOption) error {
	out.WriteString(s.lockStrength)

	if s.noWait {
		out.WriteString("NOWAIT")
	}

	if s.skipLocked {
		out.WriteString("SKIP LOCKED")
	}

	return nil
}
