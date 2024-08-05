package biz

// SaveFile 保存文件
func SaveFile(ctx *Context, file *File) error {
	if !file.isModify {
		// 文件未更新，跳过
		return nil
	}

	if !file.rwLock.TryLock() {
		return ErrWriteLockError
	}
	defer file.rwLock.Unlock()

	if err := ctx.rw.WriteFile(ctx.ctx, file.Path, file.data); err != nil {
		return err
	}
	file.isModify = false
	return nil
}
