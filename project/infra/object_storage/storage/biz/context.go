package biz

import "context"

// ReadWriter 读写接口
type ReadWriter interface {
	// ReadFile 读取文件
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// WriteFile 写入文件
	WriteFile(ctx context.Context, path string, data []byte) error
}

type Context struct {
	ctx context.Context

	rw ReadWriter
}

// ContextBuilder ctx 构建器
type ContextBuilder struct {
	rw ReadWriter
}

// New 创建 ctx
func (c *ContextBuilder) New(ctx context.Context) *Context {
	return &Context{
		ctx: ctx,
		rw:  c.rw,
	}
}
