package biz

import "errors"

var (
	ErrUnimplemented  = errors.New("方法未实现")
	ErrFileNotFound   = errors.New("文件不存在")
	ErrFileReadError  = errors.New("文件读取失败")
	ErrFileWriteError = errors.New("文件写入失败")
	ErrReadLockError  = errors.New("获取读锁失败")
	ErrWriteLockError = errors.New("获取写锁失败")
)
