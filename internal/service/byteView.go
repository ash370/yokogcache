package service

// 封装真实的缓存值，对外提供只读接口防止底层数据被修改
// 该结构体会作为双向链表的Value字段，因此需要实现Value接口
type ByteView struct {
	b []byte
}

func cloneBytes(b []byte) []byte {
	return append([]byte{}, b...)
}

//所有方法使用值接收，不允许修改结构体本身

// 返回一份深拷贝的副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 由于不允许获取原始数据本身，这里提供一个转为string类型的接口
func (v ByteView) String() string {
	return string(v.b)
}

func (v ByteView) Len() int {
	return len(v.b)
}
