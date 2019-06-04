// +build disabled

package sqlbuilder

import (
	"time"

	gc "gopkg.in/check.v1"

	"github.com/dropbox/godropbox/errors"
)

type StmtSuite struct {
}

var _ = gc.Suite(&StmtSuite{})

// NOTE: tables / columns are defined in test_utils.go

//
// SELECT Statement tests
//

func (s *StmtSuite) TestSelectEmptyProjection(c *gc.C) {
	_, err := table1.Select().String()

	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestSelectSingleColumn(c *gc.C) {
	sql, err := table1.Select(table1Col1).String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1")
}

func (s *StmtSuite) TestSelectMultiColumns(c *gc.C) {
	sql, err := table1.Select(table1Col1, table1ColFloat).String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table1.col2 FROM db.table1")
}

func (s *StmtSuite) TestSelectWhere(c *gc.C) {
	q := table1.Select(table1Col1).Where(GtL(table1Col1, 123))
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 WHERE table1.col1>123")
}

func (s *StmtSuite) TestSelectWhereDate(c *gc.C) {
	date := time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC)

	q := table1.Select(table1Col1).Where(GtL(table1ColTime, date))
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 "+
			"WHERE table1.col4>'1999-01-02 03:04:05.000000'")
}

func (s *StmtSuite) TestSelectAndWhere(c *gc.C) {
	q := table1.Select(table1Col1).AndWhere(GtL(table1Col1, 123))
	q.AndWhere(LtL(table1Col1, 321))
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 WHERE (table1.col1>123 AND table1.col1<321)")
}

func (s *StmtSuite) TestSelectCopy(c *gc.C) {
	q := table1.Select(table1Col1).Where(GtL(table1Col1, 123))
	qq := q.Copy().Where(GtL(table1Col1, 321)).OrderBy(table1Col1)

	// Initial query unchanged
	sql, err := q.String()
	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 WHERE table1.col1>123")
	// New query changed
	sql, err = qq.String()
	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 WHERE table1.col1>321 ORDER BY table1.col1")

}

func (s *StmtSuite) TestSelectLimitWithoutOffset(c *gc.C) {
	q := table1.Select(table1Col1).Limit(5)
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 LIMIT 5")
}

func (s *StmtSuite) TestSelectLimitWithOffset(c *gc.C) {
	q := table1.Select(table1Col1).Limit(5).Offset(2)
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 LIMIT 2, 5")
}

func (s *StmtSuite) TestSelectGroupBy(c *gc.C) {
	q := table1.Select(
		table1Col1,
		table1ColFloat,
		Alias("total", SqlFunc("sum", table1Col3)))
	q.GroupBy(table1Col1, table1ColFloat)
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table1.col2,"+
			"(sum(table1.col3)) AS \"total\" "+
			"FROM db.table1 GROUP BY table1.col1,table1.col2")
}

func (s *StmtSuite) TestSelectSingleOrderBy(c *gc.C) {
	q := table1.Select(table1Col1, table1ColFloat).OrderBy(table1ColFloat)
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table1.col2 "+
			"FROM db.table1 ORDER BY table1.col2")
}

func (s *StmtSuite) TestSelectOrderByAsc(c *gc.C) {
	q := table1.Select(table1Col1, table1ColFloat).OrderBy(ASC(table1ColFloat))
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table1.col2 "+
			"FROM db.table1 ORDER BY table1.col2 ASC")
}

func (s *StmtSuite) TestSelectOrderByDesc(c *gc.C) {
	q := table1.Select(table1Col1, table1ColFloat).OrderBy(DESC(table1ColFloat))
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table1.col2 "+
			"FROM db.table1 ORDER BY table1.col2 DESC")
}

func (s *StmtSuite) TestSelectMultiOrderBy(c *gc.C) {
	q := table1.Select(table1Col1, table1ColFloat)
	q.OrderBy(table1ColFloat, table1Col1)
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table1.col2 "+
			"FROM db.table1 "+
			"ORDER BY table1.col2,table1.col1")
}

func (s *StmtSuite) TestSelectOnJoin(c *gc.C) {

	join := table1.InnerJoinOn(table2, EQ(table1Col3, table2Col3))
	sql, err := join.Select(table1Col1, table2Col4).String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1,table2.col4 "+
			"FROM db.table1 JOIN db.table2 "+
			"ON table1.col3=table2.col3")
}

func (s *StmtSuite) TestSelectWithSharedLock(c *gc.C) {

	q := table1.Select(table1Col1).Where(GtL(table1Col1, 123)).WithSharedLock()
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT table1.col1 FROM db.table1 "+
			"WHERE table1.col1>123 LOCK IN SHARE MODE")
}

func (s *StmtSuite) TestSelectDistinct(c *gc.C) {
	q := table1.Select(table1Col1).Distinct()
	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"SELECT DISTINCT table1.col1 FROM db.table1")
}

//
// INSERT Statement tests
//

func (s *StmtSuite) TestInsertNoColumn(c *gc.C) {
	_, err := table1.INSERT().Add().String()

	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestInsertNoRow(c *gc.C) {
	_, err := table1.INSERT(table1Col1).String()

	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestInsertColumnLengthMismatch(c *gc.C) {
	_, err := table1.INSERT(table1Col1, table1ColFloat).Add(nil).String()

	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestInsertNilValue(c *gc.C) {
	_, err := table1.INSERT(table1Col1).Add(nil).String()

	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestInsertNilColumn(c *gc.C) {
	_, err := table1.INSERT(nil).Add(literal(1)).String()

	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestInsertSingleValue(c *gc.C) {
	sql, err := table1.INSERT(table1Col1).Add(literal(1)).String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT INTO db.table1 (table1.col1) VALUES (1)")
}

func (s *StmtSuite) TestInsertDate(c *gc.C) {
	date := time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC)

	sql, err := table1.INSERT(table1ColTime).Add(literal(date)).String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT INTO db.table1 (table1.col4) "+
			"VALUES ('1999-01-02 03:04:05.000000')")
}

func (s *StmtSuite) TestInsertIgnore(c *gc.C) {
	stmt := table1.INSERT(table1Col1).Add(literal(1)).IgnoreDuplicates(true)
	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT IGNORE INTO db.table1 (table1.col1) VALUES (1)")
}

func (s *StmtSuite) TestInsertMultipleValues(c *gc.C) {
	stmt := table1.INSERT(table1Col1, table1ColFloat, table1Col3)
	stmt.Add(literal(1), literal(2), literal(3))

	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT INTO db.table1 "+
			"(table1.col1,table1.col2,table1.col3) "+
			"VALUES (1,2,3)")
}

func (s *StmtSuite) TestInsertMultipleRows(c *gc.C) {
	stmt := table1.INSERT(table1Col1, table1ColFloat)
	stmt.Add(literal(1), literal(2))
	stmt.Add(literal(11), literal(22))
	stmt.Add(literal(111), literal(222))

	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT INTO db.table1 "+
			"(table1.col1,table1.col2) "+
			"VALUES (1,2), (11,22), (111,222)")
}

func (s *StmtSuite) TestOnDuplicateKeyUpdateNilCol(c *gc.C) {
	stmt := table1.INSERT(table1Col1, table1ColFloat)
	stmt.Add(literal(1), literal(2))
	stmt.AddOnDuplicateKeyUpdate(nil, literal(3))

	_, err := stmt.String()
	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestOnDuplicateKeyUpdateNilExpr(c *gc.C) {
	stmt := table1.INSERT(table1Col1, table1ColFloat)
	stmt.Add(literal(1), literal(2))
	stmt.AddOnDuplicateKeyUpdate(table1Col1, nil)

	_, err := stmt.String()
	c.Assert(err, gc.NotNil)
}

func (s *StmtSuite) TestOnDuplicateKeyUpdateSingle(c *gc.C) {
	stmt := table1.INSERT(table1Col1, table1ColFloat)
	stmt.Add(literal(1), literal(2))
	stmt.AddOnDuplicateKeyUpdate(table1Col3, literal(3))

	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT INTO db.table1 "+
			"(table1.col1,table1.col2) "+
			"VALUES (1,2) "+
			"ON DUPLICATE KEY UPDATE table1.col3=3")
}

func (s *StmtSuite) TestOnDuplicateKeyUpdateMulti(c *gc.C) {
	stmt := table1.INSERT(table1Col1, table1ColFloat)
	stmt.Add(literal(1), literal(2))
	stmt.AddOnDuplicateKeyUpdate(table1Col3, literal(3))
	stmt.AddOnDuplicateKeyUpdate(table1ColFloat, literal(4))

	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)

	c.Assert(
		sql,
		gc.Equals,
		"INSERT INTO db.table1 "+
			"(table1.col1,table1.col2) "+
			"VALUES (1,2) "+
			"ON DUPLICATE KEY UPDATE table1.col3=3, table1.col2=4")
}

//
// LOCK/UNLOCK Statement tests ================================================
//

func (s *StmtSuite) TestLockStatement(c *gc.C) {
	stmt := NewLockStatement().AddReadLock(table1).AddWriteLock(table2)
	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)

	c.Assert(sql, gc.Equals, "LOCK TABLES db.table1 READ, db.table2 WRITE")
}

func (s *StmtSuite) TestUnlockStatement(c *gc.C) {
	stmt := NewUnlockStatement()
	sql, err := stmt.String()
	c.Assert(err, gc.IsNil)
	c.Assert(sql, gc.Equals, "UNLOCK TABLES")

}

func (s *StmtSuite) TestUnionSelectStatement(c *gc.C) {
	select_queries := make([]selectStatement, 0, 3)

	select_queries = append(select_queries,
		table1.Select(table1Col1).Where(GtL(table1Col1, 123)),
		table1.Select(table1Col1).Where(GtL(table1Col1, 456)),
		table1.Select(table1Col1).Where(LtL(table1Col1, 23)),
	)

	q := UNION(select_queries...)

	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"(SELECT table1.col1 FROM db.table1 WHERE table1.col1>123) "+
			"UNION (SELECT table1.col1 FROM db.table1 WHERE table1.col1>456) "+
			"UNION (SELECT table1.col1 FROM db.table1 WHERE table1.col1<23)")
}

func (s *StmtSuite) TestUnionLimitWithoutOrderBy(c *gc.C) {
	select_queries := make([]selectStatement, 0, 3)

	select_queries = append(select_queries,
		table1.Select(table1Col1).Where(GtL(table1Col1, 123)).OrderBy(table1ColFloat),
		table1.Select(table1Col1).Where(GtL(table1Col1, 456)),
		table1.Select(table1Col1).Where(LtL(table1Col1, 23)),
	)

	q := UNION(select_queries...)

	_, err := q.String()

	c.Assert(err, gc.NotNil)
	c.Assert(
		errors.GetMessage(err),
		gc.Equals,
		"All inner selects in UNION Statement must have LIMIT if they have ORDER BY")
}

func (s *StmtSuite) TestUnionSelectWithMismatchedColumns(c *gc.C) {
	select_queries := make([]selectStatement, 0, 3)

	select_queries = append(select_queries,

		table1.Select(
			table1Col1,
			table1ColFloat,
			table1Col3,
			table1ColTime).AndWhere(GtL(table1Col1, 123)).AndWhere(LtL(table1Col1, 321)),
		table1.Select(table1Col1).Where(And(GtL(table1Col1, 123), LtL(table1Col1, 321))),
		table1.Select(table1Col1).Where(LtL(table1Col1, 23)).OrderBy(table1ColTime).Limit(20),
	)

	q := UNION(select_queries...)
	q = q.Where(And(LtL(table1Col1, 1000), GtL(table1Col1, 15)))
	q = q.ORDER_BY(DESC(table1ColTime), ASC(table1Col3))
	q = q.LIMIT(5)

	_, err := q.String()

	c.Assert(err, gc.NotNil)
	c.Assert(
		errors.GetMessage(err),
		gc.Equals,
		"All inner selects in UNION Statement must select the "+
			"same number of columns.  For sanity, you probably "+
			"want to select the same tableName columns in the same "+
			"orderBy.  If you are selecting on multiple tables, "+
			"use Null to pad to the right number of fields.")
}

func (s *StmtSuite) TestComplicatedUnionSelectWithWhereStatement(c *gc.C) {

	// tests on outer Statement: Group By, Order By, LIMIT
	// on inner Statement: AndWhere, WHERE (with AND), Order By, LIMIT
	select_queries := make([]selectStatement, 0, 3)

	// We're not trying to write a SQL parser, so we won't warn if you do something silly like
	// try to apply a where clause on more columns than you've selected in your union select
	select_queries = append(select_queries,
		table1.Select(
			table1Col1,
		).AndWhere(GtL(table1Col1, 123)).AndWhere(LtL(table1Col1, 321)),
		table1.Select(
			table1Col1,
		).Where(And(GtL(table1Col1, 456), LtL(table1Col1, 654))),
		table1.Select(
			table1Col1,
		).Where(LtL(table1Col1, 23)).OrderBy(table1ColTime).Limit(20),
	)

	q := UNION(select_queries...)
	q = q.Where(And(LtL(table1Col1, 1000), GtL(table1Col1, 15)))

	q = q.ORDER_BY(DESC(table1ColTime), ASC(table1Col3))
	q = q.LIMIT(5)
	q = q.GroupBy(table1ColTime)

	sql, err := q.String()

	c.Assert(err, gc.IsNil)
	c.Assert(
		sql,
		gc.Equals,
		"(SELECT table1.col1 FROM db.table1 WHERE "+
			"(table1.col1>123 AND table1.col1<321)) "+
			"UNION (SELECT table1.col1 FROM db.table1 "+
			"WHERE (table1.col1>456 AND table1.col1<654)) "+
			"UNION (SELECT table1.col1 FROM db.table1 "+
			"WHERE table1.col1<23 ORDER BY table1.col4 LIMIT 20) "+
			"WHERE (table1.col1<1000 AND table1.col1>15) "+
			"GROUP BY table1.col4 ORDER BY table1.col4 DESC,table1.col3 ASC "+
			"LIMIT 5")

}
