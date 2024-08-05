package biz

import "sync"

// File 文件对象
type File struct {
	// Path 文件路径
	Path string

	// data 文件内容字节流
	data []byte

	// md5 校验码
	md5 string

	// isModify 标记是否修改了文件内容
	isModify bool

	// rwLock 读写锁
	rwLock *sync.RWMutex
}

// GetData 获取文件内容
func (f *File) GetData(ctx *Context) ([]byte, error) {
	if f.data != nil {
		return f.data, nil
	}

	if !f.rwLock.TryRLock() {
		// 加锁失败
		return nil, ErrReadLockError
	}
	defer f.rwLock.RUnlock() 

	data, err := ctx.rw.ReadFile(ctx.ctx, f.Path)
	if err != nil {
		return nil, err
	}

	f.data = data
	return data, nil
}
