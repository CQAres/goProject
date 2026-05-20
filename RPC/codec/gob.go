package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec 基于 gob 编码的 Codec 实现
type GobCodec struct {
	conn io.ReadWriteCloser // conn 底层网络连接
	buf  *bufio.Writer      // buf 写缓冲，减少网络 IO 次数
	dec  *gob.Decoder       // dec gob 解码器，从连接读取
	enc  *gob.Encoder       // enc gob 编码器，写入缓冲
}

// ReadHeader 从连接中解码读取 RPC 消息头
func (g *GobCodec) ReadHeader(header *Header) error {
	return g.dec.Decode(header)
}

// ReadBody 从连接中解码读取 RPC 消息体
func (g *GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}

// Write 将消息头和消息体编码写入连接，写入后自动刷新缓冲区；编码失败时关闭连接
func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.enc.Encode(header); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err = g.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

// Close 关闭底层网络连接
func (g *GobCodec) Close() error {
	return g.conn.Close()
}

// 确保 GobCodec 实现了 Codec 接口
var _ Codec = (*GobCodec)(nil)

// NewGobCodec 创建基于 gob 编码的 Codec 实例
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
