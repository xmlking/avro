package rpc

import (
	"bufio"
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hamba/avro"
)

var (
	ErrServerClosed = errors.New("rpc: server closed")
)

type ResponseWriter interface {
	Metadata(key string, value []byte)

	Write(value interface{})

	Error(value interface{})
}

type Handler interface {
	Serve(ResponseWriter, *Request)
}

type response struct {
	conn      *conn
	cancelCtx context.CancelFunc
	req       *Request
}

type ConnState int

const (
	StateNew ConnState = iota
	StateActive
	StateIdle
	StateClosed
)

type conn struct {
	server *Server

	rwc        net.Conn
	remoteAddr string
	cancelCtx  context.CancelFunc

	bufr *bufio.Reader
	bufw *bufio.Writer

	werr error

	state uint32
}

func (c *conn) setState(nc net.Conn, state ConnState) {
	atomic.StoreUint32(&c.state, uint32(state))

	srv := c.server
	switch state {
	case StateNew:
		srv.trackConn(c, true)
	case StateClosed:
		srv.trackConn(c, false)
	}
}

func (c *conn) getState() ConnState {
	return ConnState(atomic.LoadUint32(&c.state))
}

func (c *conn) readRequest(ctx context.Context) (w *response, err error) {
	if d := c.server.ReadTimeout; d != 0 {
		c.rwc.SetReadDeadline(time.Now().Add(d))
	}
	if d := c.server.WriteTimeout; d != 0 {
		defer func() {
			c.rwc.SetWriteDeadline(time.Now().Add(d))
		}()
	}

	//TODO: Handshake

	req, err := ReadRequest(c.bufr)
	if err != nil {
		return nil, err
	}
	ctx, cancelCtx := context.WithCancel(ctx)
	req.ctx = ctx
	req.RemoteAddr = c.remoteAddr

	return &response{
		conn:      c,
		cancelCtx: cancelCtx,
		req:       req,
	}, nil
}

func (c *conn) serve(ctx context.Context) {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	defer func() {
		if err := recover(); err != nil {
			//TODO: do something with this error
		}

		c.close()
		c.setState(c.rwc, StateClosed)
	}()

	ctx, cancelCtx := context.WithCancel(ctx)
	c.cancelCtx = cancelCtx
	defer cancelCtx()

	//TODO: pool these
	c.bufr = bufio.NewReader(c.rwc)
	c.bufw = bufio.NewWriter(c.rwc)

	for {
		w, err := c.readRequest(ctx)
		if err != nil {
			//TODO: write a system error back
		}
		c.setState(c.rwc, StateActive)

		req := w.req
		c.server.Handler.Serve(w, req)
		w.cancelCtx()
		c.flush()

		if c.werr != nil {
			return
		}
		c.setState(c.rwc, StateIdle)

		if d := c.server.idleTimeout(); d != 0 {
			c.rwc.SetReadDeadline(time.Now().Add(d))
			if _, err := c.bufr.Peek(1); err != nil {
				return
			}
		}
		c.rwc.SetReadDeadline(time.Time{})
	}
}

func (c *conn) flush() {
	c.bufw.Flush()
}

func (c *conn) close() {
	c.flush()
	c.rwc.Close()
}

type atomicBool int32

func (b *atomicBool) isSet() bool {
	return atomic.LoadInt32((*int32)(b)) != 0
}

func (b *atomicBool) setTrue() {
	atomic.StoreInt32((*int32)(b), 1)
}

type Server struct {
	Addr string

	Protocol *avro.Protocol

	Handler Handler

	ReadTimeout time.Duration

	WriteTimeout time.Duration

	IdleTimeout time.Duration

	inShutdown atomicBool

	mu         sync.Mutex
	listeners  map[*net.Listener]struct{}
	activeConn map[*conn]struct{}
}

func (s *Server) idleTimeout() time.Duration {
	if s.IdleTimeout != 0 {
		return s.IdleTimeout
	}
	return s.ReadTimeout
}

func (s *Server) trackListener(ln *net.Listener, add bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listeners == nil {
		s.listeners = make(map[*net.Listener]struct{})
	}

	if add {
		if s.inShutdown.isSet() {
			return false
		}

		s.listeners[ln] = struct{}{}
		return true
	}

	delete(s.listeners, ln)
	return true
}

func (s *Server) trackConn(c *conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[*conn]struct{})
	}

	if add {
		s.activeConn[c] = struct{}{}
		return
	}

	delete(s.activeConn, c)
}

func (s *Server) ListenAndServe() error {
	if s.inShutdown.isSet() {
		return ErrServerClosed
	}

	addr := s.Addr
	if addr == "" {
		addr = ":8090"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(ln.(*net.TCPListener))
}

func (s *Server) newConn(rwc net.Conn) *conn {
	return &conn{
		server: s,
		rwc:    rwc,
	}
}

type onceCloseListener struct {
	net.Listener
	once sync.Once
	err  error
}

func (l *onceCloseListener) close() {
	l.err = l.Listener.Close()
}

func (l *onceCloseListener) Close() error {
	l.once.Do(l.close)
	return l.err
}

func (s *Server) Serve(ln net.Listener) error {
	if s.Protocol == nil {
		return errors.New("rpc: protocol is required")
	}

	if s.Handler == nil {
		return errors.New("rpc: handler is required")
	}

	ln = &onceCloseListener{Listener: ln}
	defer ln.Close()

	if !s.trackListener(&ln, true) {
		return ErrServerClosed
	}
	defer s.trackListener(&ln, false)

	ctx := context.Background()
	for {
		rw, err := ln.Accept()
		if err != nil {
			if s.inShutdown.isSet() {
				return ErrServerClosed
			}

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			return err
		}

		c := s.newConn(rw)
		c.setState(c.rwc, StateNew)
		go c.serve(ctx)
	}
}

func (s *Server) Close() error {
	s.inShutdown.setTrue()

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.closeListenersLocked()
	for c := range s.activeConn {
		c.rwc.Close()
		delete(s.activeConn, c)
	}
	return err
}

var shutdownPollInterval = 100 * time.Millisecond

func (s *Server) Shutdown(ctx context.Context) error {
	s.inShutdown.setTrue()

	s.mu.Lock()
	err := s.closeListenersLocked()
	s.mu.Unlock()

	ticker := time.NewTicker(shutdownPollInterval)
	defer ticker.Stop()
	for {
		if s.closeIdleConns() {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Server) closeListenersLocked() error {
	var err error
	for ln := range s.listeners {
		if cerr := (*ln).Close(); cerr != nil && err == nil {
			err = cerr
		}
		delete(s.listeners, ln)
	}
	return err
}

func (s *Server) closeIdleConns() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	quiescent := true
	for c := range s.activeConn {
		state := c.getState()
		if state != StateIdle {
			quiescent = false
			continue
		}

		c.rwc.Close()
		delete(s.activeConn, c)
	}

	return quiescent
}
