package builder

import (
	"context"
)

type CommentAddition struct {
}

func (CommentAddition) AdditionType() AdditionType {
	return AdditionComment
}

func Comment(c string) *comment {
	return &comment{text: []byte(c)}
}

var _ Addition = (*comment)(nil)

type comment struct {
	CommentAddition

	text []byte
}

func (c *comment) IsNil() bool {
	return c == nil || len(c.text) == 0
}

func (c *comment) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.WhiteComments(c.text)
	return e.Ex(ctx)
}
