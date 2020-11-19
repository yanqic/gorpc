package codec

import "io"
// Header 协议头
type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq uint64 // sequence number by client
	Error string 
}

// Codec 协议抽象
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}
// NewCodeFunc 构造函数
type NewCodeFunc func (io.ReadWriteCloser) Codec

// Type 参数类型
type Type string

const (
	// GobType Gob类型
	GobType Type = "application/gob"
	// JsonType json类型
	JsonType Type = "application/json"
)

// NewCodeFuncMap 构造函数对象
var NewCodeFuncMap map[Type]NewCodeFunc

func init() {
	NewCodeFuncMap = make(map[Type]NewCodeFunc)
	NewCodeFuncMap[GobType] = NewGobCodec
}