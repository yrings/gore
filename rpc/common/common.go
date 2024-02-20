package common

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

type Session struct {
	Conn net.Conn
}

// 定义RPC交互的数据结构
type RPCData struct {
	// 访问的函数
	Name string
	// 访问时的参数
	Args []interface{}
}

func NewSession(conn net.Conn) *Session {
	return &Session{Conn: conn}
}

func (s *Session) Write(data []byte) error {
	// 定义写数据格式
	// 4字节头部 + 可变的长度
	buf := make([]byte, 4+len(data))
	// 写入头部
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	// 将数据内容放到头部后面
	copy(buf[4:], data)
	_, err := s.Conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Read() ([]byte, error) {
	// 读取头部信息
	header := make([]byte, 4)
	// 读取长度信息
	_, err := io.ReadFull(s.Conn, header)
	if err != nil {
		return nil, err
	}
	// 读取内容
	dataLen := binary.BigEndian.Uint32(header)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(s.Conn, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GobEncode(data RPCData) ([]byte, error) {
	// 得到字节数组的编码器
	var buf bytes.Buffer
	bufEnc := gob.NewEncoder(&buf)
	// 编码器对数据编码
	if err := bufEnc.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GobDecode(data []byte) (RPCData, error) {
	buf := bytes.NewBuffer(data)
	// 得到字节数组编码器
	bufDec := gob.NewDecoder(buf)
	// 解码器对数据解码
	var resData RPCData
	if err := bufDec.Decode(&resData); err != nil {
		return resData, err
	}
	return resData, nil
}
