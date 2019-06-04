package sqlbuilder

import "fmt"

// Representation of an escaped literal
type literalExpression struct {
	expressionInterfaceImpl
	value    interface{}
	constant bool
}

func literal(value interface{}) *literalExpression {
	exp := literalExpression{value: value}
	exp.expressionInterfaceImpl.parent = &exp

	return &exp
}

func constLiteral(value interface{}) *literalExpression {
	exp := literal(value)
	exp.constant = true

	return exp
}

func (l literalExpression) serialize(statement statementType, out *queryData, options ...serializeOption) error {
	if l.constant {
		out.insertConstantArgument(l.value)
	} else {
		out.insertPreparedArgument(l.value)
	}

	return nil
}

type integerLiteralExpression struct {
	literalExpression
	integerInterfaceImpl
}

func Int(value int64) IntegerExpression {
	numLiteral := &integerLiteralExpression{}

	numLiteral.literalExpression = *literal(value)

	numLiteral.literalExpression.parent = numLiteral
	numLiteral.integerInterfaceImpl.parent = numLiteral

	return numLiteral
}

//---------------------------------------------------//
type boolLiteralExpression struct {
	boolInterfaceImpl
	literalExpression
}

func Bool(value bool) BoolExpression {
	boolLiteralExpression := boolLiteralExpression{}

	boolLiteralExpression.literalExpression = *literal(value)
	boolLiteralExpression.boolInterfaceImpl.parent = &boolLiteralExpression

	return &boolLiteralExpression
}

//---------------------------------------------------//
type floatLiteral struct {
	floatInterfaceImpl
	literalExpression
}

func Float(value float64) FloatExpression {
	floatLiteral := floatLiteral{}
	floatLiteral.literalExpression = *literal(value)

	floatLiteral.floatInterfaceImpl.parent = &floatLiteral

	return &floatLiteral
}

//---------------------------------------------------//
type stringLiteral struct {
	stringInterfaceImpl
	literalExpression
}

func String(value string) StringExpression {
	stringLiteral := stringLiteral{}
	stringLiteral.literalExpression = *literal(value)

	stringLiteral.stringInterfaceImpl.parent = &stringLiteral

	return &stringLiteral
}

//---------------------------------------------------//
type timeLiteral struct {
	timeInterfaceImpl
	literalExpression
}

func Time(hour, minute, second, milliseconds int) TimeExpression {
	timeLiteral := timeLiteral{}
	timeStr := fmt.Sprintf("%02d:%02d:%02d.%03d", hour, minute, second, milliseconds)
	timeLiteral.literalExpression = *literal(timeStr)

	timeLiteral.timeInterfaceImpl.parent = &timeLiteral

	return timeLiteral.CAST_TO_TIME()
}

//---------------------------------------------------//
type timezLiteral struct {
	timezInterfaceImpl
	literalExpression
}

func Timez(hour, minute, second, milliseconds, timezone int) TimezExpression {
	timezLiteral := timezLiteral{}
	timeStr := fmt.Sprintf("%02d:%02d:%02d.%03d %+03d", hour, minute, second, milliseconds, timezone)
	timezLiteral.literalExpression = *literal(timeStr)

	timezLiteral.timezInterfaceImpl.parent = &timezLiteral

	return timezLiteral.CAST_TO_TIMEZ()
}

//---------------------------------------------------//
type timestampLiteral struct {
	timestampInterfaceImpl
	literalExpression
}

func Timestamp(year, month, day, hour, minute, second, milliseconds int) TimestampExpression {
	timestampLiteral := timestampLiteral{}
	timeStr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d", year, month, day, hour, minute, second, milliseconds)
	timestampLiteral.literalExpression = *literal(timeStr)

	timestampLiteral.timestampInterfaceImpl.parent = &timestampLiteral

	return timestampLiteral.CAST_TO_TIMESTAMP()
}

//---------------------------------------------------//
type timestampzLiteral struct {
	timestampzInterfaceImpl
	literalExpression
}

func Timestampz(year, month, day, hour, minute, second, milliseconds, timezone int) TimestampzExpression {
	timestampzLiteral := timestampzLiteral{}
	timeStr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d %+04d",
		year, month, day, hour, minute, second, milliseconds, timezone)

	timestampzLiteral.literalExpression = *literal(timeStr)

	timestampzLiteral.timestampzInterfaceImpl.parent = &timestampzLiteral

	return timestampzLiteral.CAST_TO_TIMESTAMPZ()
}

//---------------------------------------------------//
type dateLiteral struct {
	dateInterfaceImpl
	literalExpression
}

func Date(year, month, day int) DateExpression {
	dateLiteral := dateLiteral{}

	timeStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	dateLiteral.literalExpression = *literal(timeStr)
	dateLiteral.dateInterfaceImpl.parent = &dateLiteral

	return dateLiteral.CAST_TO_DATE()
}

//--------------------------------------------------//
type nullLiteral struct {
	expressionInterfaceImpl
}

func newNullLiteral() expression {
	nullExpression := &nullLiteral{}

	nullExpression.expressionInterfaceImpl.parent = nullExpression

	return nullExpression
}

func (n *nullLiteral) serialize(statement statementType, out *queryData, options ...serializeOption) error {
	out.writeString("NULL")
	return nil
}

//--------------------------------------------------//
type starLiteral struct {
	expressionInterfaceImpl
}

func newStarLiteral() expression {
	starExpression := &starLiteral{}

	starExpression.expressionInterfaceImpl.parent = starExpression

	return starExpression
}

func (n *starLiteral) serialize(statement statementType, out *queryData, options ...serializeOption) error {
	out.writeString("*")
	return nil
}

//---------------------------------------------------//

type wrap struct {
	expressionInterfaceImpl
	expressions []expression
}

func (n *wrap) serialize(statement statementType, out *queryData, options ...serializeOption) error {
	out.writeString("(")
	err := serializeExpressionList(statement, n.expressions, ", ", out)
	out.writeString(")")
	return err
}

func WRAP(expression ...expression) expression {
	wrap := &wrap{expressions: expression}
	wrap.expressionInterfaceImpl.parent = wrap

	return wrap
}
