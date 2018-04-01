package socket

// Result 结果
type Result int32

const (
	Result_OK            Result = iota
	Result_SocketError          // 网络错误
	Result_SocketTimeout        // Socket超时
	Result_PackageCrack         // 封包破损
	Result_CodecError
	Result_RequestClose // 请求关闭
	Result_NextChain
	Result_RPCTimeout
)
