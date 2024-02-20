package client

import (
	"fmt"
	"net"
	"reflect"
	"yrings-gore-rpc/common"
)

// 客户端只有函数原型，使用reflect.MakeFunc() 可以完成原型函数到函数的调用

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

// 实现通用的RPC客户端
// 传入访问的函数名
// fPtr指向的是函数原型
// var select fun xx(User)
// cli.callRPC("selectUser",&select)
func (c *Client) CallRPC(rpcName string, fPtr interface{}) {
	// 调过反射，获取fPtr未初始化的函数原型
	fn := reflect.ValueOf(fPtr).Elem()
	// 处理函数参数

	f := func(args []reflect.Value) []reflect.Value {
		// 处理参数
		inArgs := make([]interface{}, 0, len(args))
		for _, arg := range args {
			inArgs = append(inArgs, arg.Interface())
		}
		// 连接
		cliSession := common.NewSession(c.conn)
		// 编码数据
		reqRPC := common.RPCData{Name: rpcName, Args: inArgs}
		data, err := common.GobEncode(reqRPC)
		if err != nil {
			panic(err)
		}
		// 写数据
		err = cliSession.Write(data)
		if err != nil {
			panic(err)
		}
		// 接收服务端返回值
		respBytes, err := cliSession.Read()
		if err != nil {
			panic(err)
		}
		// 解码
		respRPC, err := common.GobDecode(respBytes)
		if err != nil {
			panic(err)
		}
		fmt.Println("respData:", respRPC)
		fmt.Println("respData.len:", len(respRPC.Args))
		// 处理服务端数据
		outArgs := make([]reflect.Value, 0, len(respRPC.Args))
		for i, arg := range respRPC.Args {
			// 对nil进行转化，reflect.Zero()会返回类型的零值的value
			// .out()会返回函数输出的参数类型
			if arg == nil {
				outArgs = append(outArgs, reflect.Zero(fn.Type().Out(i)))
				continue
			}
			outArgs = append(outArgs, reflect.ValueOf(arg))
		}
		return outArgs
	}
	v := reflect.MakeFunc(fn.Type(), f)
	fn.Set(v)
}
