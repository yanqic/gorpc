package gorpc

import (
  "encoding/json"
  "fmt"
  "gorpc/codec"
  "io"
  "log"
  "net"
  "reflect"
  "sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
  MagicNumber int
  CodecType codec.Type
}

var DefaultOption = &Option{
  MagicNumber: MagicNumber,
  CodecType: codec.GobType,
}

type Server struct {}

func NewServer() *Server {
  return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener)  {
  for {
    conn, err := lis.Accept()
    if err != nil {
      log.Println("rpc server: accept error:", err)
      return
    }
    go server.ServeConn(conn) // TODO:
  }
}

func Accept(lis net.Listener)  {
  DefaultServer.Accept(lis)
}

func (server *Server) ServeConn(conn io.ReadWriteCloser)  {
  defer func() { _ = conn.Close() }()
  var opt Option
  if err := json.NewDecoder(conn).Decode(&opt); err != nil {
    log.Println("rpc server: options error: ", err)
    return
  }
  if opt.MagicNumber != MagicNumber {
    log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
    return
  }
  f := codec.NewCodeFuncMap[opt.CodecType]
  if f == nil {
    log.Printf("rpc server: invalid codec type %s ", opt.CodecType)
  }
  server.serveCodec(f(conn))
}

var invalidRequest = struct {}{}

func (server *Server) serveCodec(cc codec.Codec)  {
  sending := new(sync.Mutex)
  wg := new(sync.WaitGroup)
  for {
    req, err := server.readRequest(cc)
    if err != nil {
      if req == nil {
        break
      }
      req.h.Error = err.Error()
      server.sendResponse(cc, req.h, invalidRequest, sending)
      continue
    }
    wg.Add(1)
    go server.sendResponse(cc, req.h, invalidRequest, sending)
  }
  wg.Wait()
  _ = cc.Close()
}

type request struct {
  h *codec.Header
  argv, replyVal reflect.Value
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
  var h codec.Header
  if err := cc.ReadHeader(&h); err != nil {
    if err != io.EOF && err != io.ErrUnexpectedEOF {
      log.Println("rpc server: read header error:", err)
    }
    return nil, err
  }
  return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
  h, err := server.readRequestHeader(cc)
  if err != nil {
    return nil, err
  }
  req := &request{h: h}
  // TODO: just string
  req.argv = reflect.New(reflect.TypeOf("")) // TODO:
  if err = cc.ReadBody(req.argv.Interface()); err != nil { // ??? what the mess?
    log.Println("rpc server: read argv err:", err)
  }
  return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
  sending.Lock()
  defer sending.Unlock()
  if err := cc.Write(h, body); err != nil {
    log.Println("rpc server: write response error:", err)
  }
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup)  {
  defer wg.Done()
  log.Println(req.h, req.argv.Elem())
  req.replyVal = reflect.ValueOf(fmt.Sprintf("gorpc resp %d", req.h.Seq))
  server.sendResponse(cc, req.h, req.replyVal.Interface(), sending)
}