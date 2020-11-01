package zkm

import (
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const chanBuffSize = 10000

var ErrTimeout = errors.New("timeout wait for response")

type req struct {
	pdu    *Pdu
	respCh chan *Resp
	j      *gocron.Job
}

func newReq(pdu *Pdu) *req {
	return &req{pdu: pdu, respCh: make(chan *Resp, 1)}
}

type Resp struct {
	Err error
	Pdu *Pdu
}

type Evt interface {
	fmt.Stringer
}

type Severity int8

const (
	Debug Severity = iota
	Info
	Warning
	Error
)

type LogEvt struct {
	severity Severity
	msg      string
}

func (e *LogEvt) Severity() Severity {
	return e.severity
}

func (e *LogEvt) Msg() string {
	return e.msg
}

func (e *LogEvt) String() string {
	return e.msg
}

type ErrEvt struct {
	err error
}

func (e *ErrEvt) String() string {
	return e.err.Error()
}

func (e *ErrEvt) Err() error {
	return e.err
}

type SessionConfig struct {
	RpsLimit               int32
	WinLimit               int32
	ThrottlePauseSec       int32
	ReqTimeoutSec          int32
	EnquireLinkIntervalSec int64
	LogSeverity            Severity
}

func NewDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		RpsLimit:               1,
		WinLimit:               1,
		ThrottlePauseSec:       1,
		ReqTimeoutSec:          2,
		EnquireLinkIntervalSec: 15,
		LogSeverity:            Info,
	}
}

type Session struct {
	sock      *sock
	scheduler *gocron.Scheduler
	inPduCh   chan *Pdu
	evtCh     chan Evt
	outReqCh  chan *req
	outRespCh chan *Pdu
	closed    chan struct{}
	cfg       *SessionConfig
}

func NewSession(conn net.Conn) *Session {
	return NewSessionWithConfig(conn, NewDefaultSessionConfig())
}

func NewSessionWithConfig(conn net.Conn, cfg *SessionConfig) *Session {
	return &Session{
		sock:      newSock(conn),
		scheduler: gocron.NewScheduler(time.UTC),
		inPduCh:   make(chan *Pdu, chanBuffSize),
		evtCh:     make(chan Evt, chanBuffSize),
		outReqCh:  make(chan *req),
		outRespCh: make(chan *Pdu, chanBuffSize),
		closed:    make(chan struct{}),
		cfg:       cfg,
	}
}

func (s *Session) Run() {
	s.scheduler.StartAsync()

	outWinSema := make(chan struct{}, 1)
	outEnquireLinkReqCh := make(chan *req, 1)
	var outWin int32
	var inWin int32
	reqsInFlight := make(map[uint32]*req)
	var lastThrottle time.Time
	lastReading := time.Now().Unix()
	mu := sync.Mutex{}

	if _, err := s.scheduler.Every(uint64(s.cfg.EnquireLinkIntervalSec)).Seconds().Do(func() {
		silenceIntervalSec := time.Now().Unix() - atomic.LoadInt64(&lastReading)
		enquireLinkIntervalSec := atomic.LoadInt64(&s.cfg.EnquireLinkIntervalSec)
		if silenceIntervalSec >= 2*enquireLinkIntervalSec {
			s.logEvt(Warning, fmt.Sprintf("silence interval [%v] exceeded two enquireLinkInterval [%v]. Socket closing...",
				silenceIntervalSec, enquireLinkIntervalSec))
			if err := s.sock.close(); err != nil {
				s.logEvt(Error, fmt.Sprintf("can't close socket: [%v]", err))
				s.errEvt(err)
			}
		} else if silenceIntervalSec >= enquireLinkIntervalSec {
			select {
			case outEnquireLinkReqCh <- newReq(NewPdu(EnquireLink)):
			default:
			}
		}
	}); err != nil {
		s.logEvt(Error, fmt.Sprintf("[]can't scheduling periodically enquireLink job: [%v]", err))
		s.errEvt(err)
	}

	//handling outgoing responses
	go func() {
		for {
			select {
			case r := <-s.outRespCh:
				err := s.sock.write(r)

				if err != nil {
					s.logEvt(Error, fmt.Sprintf("can't write pdu [%v] to socket: [%v]", r, err))
					s.errEvt(err)
				} else {
					s.logEvt(Debug, fmt.Sprintf("sent pdu: [%v]", r))
				}

				atomic.AddInt32(&inWin, -1)
			case <-s.closed:
				s.logEvt(Debug, fmt.Sprintf("goroutine handling outgoing responses completed"))
				return
			}
		}
	}()

	//handling outgoing requests
	go func() {
		sec := time.Now().Unix()
		var sentReqs int32
		var seq uint32

		handleReq := func(r *req) {
			seq++
			r.pdu.Seq = seq

			err := s.sock.write(r.pdu)

			if err != nil {
				s.logEvt(Error, fmt.Sprintf("can't write pdu [%v] to socket: [%v]", r.pdu, err))
				s.errEvt(err)
			} else {
				s.logEvt(Debug, fmt.Sprintf("sent pdu: [%v]", r.pdu))
			}

			now := time.Now()

			if s := now.Unix(); s != sec {
				sec = s
				sentReqs = 0
			}

			sentReqs++

			var pause time.Duration
			if sentReqs >= atomic.LoadInt32(&s.cfg.RpsLimit) {
				pause = time.Duration(time.Second.Nanoseconds() - int64(now.Nanosecond()))
			}

			mu.Lock()
			reqsInFlight[seq] = r
			if throttlePause := time.Second*time.Duration(atomic.LoadInt32(&s.cfg.ThrottlePauseSec)) - now.Sub(lastThrottle);
				throttlePause > pause {
				pause = throttlePause
			}
			mu.Unlock()

			_seq := seq
			if r.j, err = s.scheduler.Every(uint64(s.cfg.ReqTimeoutSec)).Seconds().Do(func() {
				mu.Lock()
				defer mu.Unlock()

				if req, ok := reqsInFlight[_seq]; ok {
					s.scheduler.RemoveByReference(req.j)

					if atomic.AddInt32(&outWin, -1) < atomic.LoadInt32(&s.cfg.WinLimit) {
						select {
						case <-outWinSema:
						default:
						}
					}

					req.respCh <- &Resp{
						Err: ErrTimeout,
					}
					delete(reqsInFlight, _seq)
					s.logEvt(Warning, fmt.Sprintf("req timeout exceeded for pdu [%v]", req.pdu))
				} else {
					s.logEvt(Warning, fmt.Sprintf("req timeout exceeded for seq [%v], but req not found", _seq))
				}
			}); err != nil {
				s.logEvt(Error, fmt.Sprintf("[]can't scheduling req timeout job: [%v]", err))
				s.errEvt(err)
			}

			if atomic.AddInt32(&outWin, 1) < atomic.LoadInt32(&s.cfg.WinLimit) {
				select {
				case <-outWinSema:
				default:
				}
			}

			time.Sleep(pause)
		}

		for {
			select {
			case outWinSema <- struct{}{}:
				select {
				case r := <-outEnquireLinkReqCh:
					handleReq(r)
				case r := <-s.outReqCh:
					handleReq(r)
				case <-s.closed:
					s.logEvt(Debug, fmt.Sprintf("goroutine handling outgoing requests completed"))
					return
				}
			case <-s.closed:
				s.logEvt(Debug, fmt.Sprintf("goroutine handling outgoing requests completed"))
				return
			}
		}
	}()

	//handling incoming requests and responses
	go func() {
		sec := time.Now().Unix()
		var receivedReqs int32
		for {
			pdu, err := s.sock.read()

			select {
			case <-s.closed:
				s.logEvt(Debug, fmt.Sprintf("goroutine handling incoming pdus completed"))
				return
			default:
			}

			if err != nil {
				s.logEvt(Error, fmt.Sprintf("can't read pdu from socket: [%v]", err))
				s.errEvt(err)
				continue
			} else {
				s.logEvt(Debug, fmt.Sprintf("received pdu: [%v]", pdu))
			}

			now := time.Now()
			atomic.StoreInt64(&lastReading, now.Unix())

			if pdu.IsReq() {
				if atomic.AddInt32(&inWin, 1) <= atomic.LoadInt32(&s.cfg.WinLimit) {
					if s := now.Unix(); s != sec {
						sec = s
						receivedReqs = 0
					}

					receivedReqs++

					if receivedReqs > atomic.LoadInt32(&s.cfg.RpsLimit) {
						resp, err := pdu.CreateResp(EsmeRThrottled)

						if err != nil {
							s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
							s.errEvt(err)
						} else {
							s.SendResp(resp)
						}
					} else {
						if pdu.Id == EnquireLink {
							resp, err := pdu.CreateResp(EsmeROk)

							if err != nil {
								s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
								s.errEvt(err)
							} else {
								s.SendResp(resp)
							}
						} else {
							s.inPduCh <- pdu
						}
					}
				} else {
					resp, err := pdu.CreateResp(EsmeRThrottled)

					if err != nil {
						s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
						s.errEvt(err)
					} else {
						s.SendResp(resp)
					}
				}
			} else {
				func() {
					mu.Lock()
					defer mu.Unlock()

					if pdu.Status == EsmeRThrottled {
						lastThrottle = now
					}

					if req, ok := reqsInFlight[pdu.Seq]; ok {
						s.scheduler.RemoveByReference(req.j)
						if atomic.AddInt32(&outWin, -1) < atomic.LoadInt32(&s.cfg.WinLimit) {
							select {
							case <-outWinSema:
							default:
							}
						}

						req.respCh <- &Resp{
							Pdu: pdu,
						}
						delete(reqsInFlight, pdu.Seq)
					} else {
						s.logEvt(Warning, fmt.Sprintf("received unexpected pdu: [%v]", pdu))
					}
				}()
			}
		}
	}()
}

func (s *Session) SendReq(pdu *Pdu) <-chan *Resp {
	req := newReq(pdu)
	if req.pdu.IsReq() {
		s.outReqCh <- req
	} else {
		req.respCh <- &Resp{
			Err: fmt.Errorf("cmd id [%v] is not request", req.pdu.Id),
		}
	}
	return req.respCh
}

func (s *Session) SendResp(pdu *Pdu) {
	if !pdu.IsReq() {
		s.outRespCh <- pdu
	} else {
		s.logEvt(Error, fmt.Sprintf("trying send bad resp: [%v]", pdu))
	}
}

func (s *Session) InPduCh() <-chan *Pdu {
	return s.inPduCh
}

func (s *Session) EvtCh() <-chan Evt {
	return s.evtCh
}

func (s *Session) Close() error {
	s.scheduler.Clear()
	s.scheduler.Stop()
	close(s.closed)
	return s.sock.close()
}

func (s *Session) logEvt(severity Severity, msg string) {
	if severity < s.cfg.LogSeverity {
		return
	}

	s.evtCh <- &LogEvt{severity: severity, msg: fmt.Sprintf("[%v]: %v", s.RemoteAddr(), msg)}
}

func (s *Session) errEvt(err error) {
	s.evtCh <- &ErrEvt{err: err}
}

func (s *Session) RemoteAddr() net.Addr {
	return s.sock.c.RemoteAddr()
}

func (s *Session) SetConfig(cfg *SessionConfig) {
	atomic.StoreInt32(&s.cfg.RpsLimit, cfg.RpsLimit)
	atomic.StoreInt32(&s.cfg.WinLimit, cfg.WinLimit)
	atomic.StoreInt32(&s.cfg.ThrottlePauseSec, cfg.ThrottlePauseSec)
	atomic.StoreInt32(&s.cfg.ReqTimeoutSec, cfg.ReqTimeoutSec)
	atomic.StoreInt64(&s.cfg.EnquireLinkIntervalSec, cfg.EnquireLinkIntervalSec)
	s.cfg.LogSeverity = cfg.LogSeverity
}
