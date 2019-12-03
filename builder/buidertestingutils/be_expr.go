package buidertestingutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func BeExpr(query string, args ...interface{}) types.GomegaMatcher {
	return &SqlExprMatcher{
		QueryMatcher: gomega.Equal(strings.TrimSpace(query)),
		ArgsMatcher:  gomega.Equal(args),
	}
}

type SqlExprMatcher struct {
	QueryMatcher types.GomegaMatcher
	ArgsMatcher  types.GomegaMatcher
}

func (matcher *SqlExprMatcher) Match(actual interface{}) (success bool, err error) {
	sqlExpr, ok := actual.(builder.SqlExpr)
	if !ok {
		return false, fmt.Errorf("actual shoud be SqlExpr,  but got %#v", actual)
	}

	if builder.IsNilExpr(sqlExpr) {
		return matcher.QueryMatcher.Match("")
	}

	expr := sqlExpr.Ex(context.Background())

	queryMatched, err := matcher.QueryMatcher.Match(expr.Query())
	if err != nil {
		return false, err
	}

	argsMatched, err := matcher.ArgsMatcher.Match(expr.Args())
	if err != nil {
		return false, err
	}

	return queryMatched && argsMatched, nil
}

func (matcher *SqlExprMatcher) FailureMessage(actual interface{}) (message string) {
	expr := actual.(builder.SqlExpr).Ex(context.Background())

	return matcher.QueryMatcher.FailureMessage(expr.Query()) + "\n" + matcher.ArgsMatcher.FailureMessage(expr.Args())
}

func (matcher SqlExprMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	expr := actual.(builder.SqlExpr).Ex(context.Background())

	return matcher.QueryMatcher.NegatedFailureMessage(expr.Query()) + "\n" + matcher.ArgsMatcher.NegatedFailureMessage(expr.Args())
}
