package zipkin

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/elodina/go-zipkin/gen-go/zipkincore"
)

func SerializeSpan(s *zipkincore.Span) ([]byte, error) {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	err := s.Write(p)
	if err != nil {
		return nil, err
	}
	return t.Buffer.Bytes(), nil
}
