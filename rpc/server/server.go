package server

import (
	"fmt"
	"gore/rpc/common"
	"net"
	"reflect"
)

type Server struct {
	// 地址
	addr string
	// map 用于维护服务
	funcs map[string]reflect.Value
}

// 构造方法
func NewServer(addr string) *Server {
	return &Server{addr: addr, funcs: make(map[string]reflect.Value)}
}

// 注册服务，传入函数名和函数
func (s *Server) Register(srcName string, f interface{}) {
	// 维护一个map,如果已存在直接退出
	if _, ok := s.funcs[srcName]; ok {
		return
	}
	// 若没有，将映射加入map
	fVal := reflect.ValueOf(f)
	s.funcs[srcName] = fVal
}

// 服务端等待调用

func (s *Server) Run() {
	// 监听
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		fmt.Printf("监听 %s err :%v", s.addr, err)
		return
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("accept err,", err)
			return
		}
		serSession := common.NewSession(conn)
		// 使用RPC
		data, err := serSession.Read()
		if err != nil {
			fmt.Println("read data error,", err)
			return
		}
		// 数据解码
		rpcData, err := common.GobDecode(data)
		if err != nil {
			fmt.Println("gob Decode error,", err)
			return
		}

		// 读取name
		f, ok := s.funcs[rpcData.Name]
		if !ok {
			fmt.Println("function name %s is not exist", rpcData.Name)
			return
		}
		// 遍历解析客户端传入参数
		args := make([]reflect.Value, 0, len(rpcData.Args))
		for _, arg := range rpcData.Args {
			args = append(args, reflect.ValueOf(arg))
		}
		// 反射调用方法
		out := f.Call(args)
		// 遍历out返回
		outArgs := make([]interface{}, 0, len(out))
		for _, outArg := range out {
			outArgs = append(outArgs, outArg.Interface())
		}
		// 数据编码
		bytes, err := common.GobEncode(common.RPCData{rpcData.Name, outArgs})
		if err != nil {
			fmt.Println("GobEncode failed!", err)
			return
		}
		// 返回到客户端
		err = serSession.Write(bytes)
		if err != nil {
			fmt.Println("write to client failed!", err)
			return
		}
	}

}
