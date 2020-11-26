package zkm

import (
	"context"
	"errors"
	"fmt"
	"github.com/Boklazhenko/scheduler"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const chanBuffSize = 10000

var ErrTimeout = errors.New("timeout wait for response")
var ErrClosed = errors.New("session closed")

type Req struct {
	Pdu *Pdu
	j   *scheduler.Job
	Ctx interface{}
}

type Resp struct {
	Err error
	Pdu *Pdu
	Req *Req
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

type SpeedController interface {
	Out() error
	In() error
	SetRpsLimit(rpsLimit int32)
	Run(ctx context.Context)
}

var errThrottling = errors.New("throttling error")

type DefaultSpeedController struct {
	rpsLimit  int32
	inSec     int64
	inReqs    int32
	outReqsCh chan struct{}
	running   chan struct{}
}

func NewDefaultSpeedController() *DefaultSpeedController {
	return &DefaultSpeedController{
		rpsLimit:  1,
		inSec:     0,
		inReqs:    0,
		outReqsCh: make(chan struct{}),
		running:   make(chan struct{}, 1),
	}
}

func (c *DefaultSpeedController) Out() error {
	select {
	case _, ok := <-c.outReqsCh:
		if ok {
			return nil
		} else {
			return fmt.Errorf("controller stopped")
		}
	}
}

func (c *DefaultSpeedController) In() error {
	s := time.Now().Unix()
	var inReqs int32 = 1
	if atomic.SwapInt64(&c.inSec, s) != s {
		atomic.StoreInt32(&c.inReqs, 1)
	} else {
		inReqs = atomic.AddInt32(&c.inReqs, 1)
	}

	if inReqs > atomic.LoadInt32(&c.rpsLimit) {
		return errThrottling
	} else {
		return nil
	}
}

func (c *DefaultSpeedController) SetRpsLimit(rpsLimit int32) {
	atomic.StoreInt32(&c.rpsLimit, rpsLimit)
}

func (c *DefaultSpeedController) Run(ctx context.Context) {
	select {
	case c.running <- struct{}{}:
	case <-ctx.Done():
		return
	}

	sec := time.Now().Unix()
	var reqs int32 = 0
	for {
		select {
		case c.outReqsCh <- struct{}{}:
			now := time.Now()
			if s := now.Unix(); s != sec {
				sec = s
				reqs = 1
			} else {
				reqs++
			}

			if reqs >= atomic.LoadInt32(&c.rpsLimit) {
				select {
				case <-time.After(time.Duration(time.Second.Nanoseconds() - int64(now.Nanosecond()))):
				case <-ctx.Done():
					close(c.outReqsCh)
					return
				}
			}
		case <-ctx.Done():
			close(c.outReqsCh)
			return
		}
	}
}

type SessionConfig struct {
	RpsLimit               int32
	WinLimit               int32
	ThrottlePauseSec       int32
	ReqTimeoutSec          int32
	EnquireLinkEnabled     bool
	EnquireLinkIntervalSec int64
	SilenceTimeoutSec      int64
	LogSeverity            Severity
}

func NewDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		RpsLimit:               1,
		WinLimit:               1,
		ThrottlePauseSec:       1,
		ReqTimeoutSec:          2,
		EnquireLinkEnabled:     false,
		EnquireLinkIntervalSec: 15,
		SilenceTimeoutSec:      60,
		LogSeverity:            Info,
	}
}

type Session struct {
	sock            *sock
	scheduler       *scheduler.Scheduler
	inReqCh         chan *Pdu
	evtCh           chan Evt
	outReqCh        chan *Req
	outRespCh       chan *Pdu
	inRespCh        chan *Resp
	cfg             *SessionConfig
	speedController SpeedController
	inWin           int32
	outWin          int32
	outWinSema      chan struct{}
	lastReading     int64
	lastWriting     int64
	reqsInFlight    map[uint32]*Req
	lastThrottle    time.Time
	mu              sync.Mutex
}

func NewSession(conn net.Conn, speedController SpeedController) *Session {
	return NewSessionWithConfig(conn, NewDefaultSessionConfig(), speedController)
}

func NewSessionWithConfig(conn net.Conn, cfg *SessionConfig, speedController SpeedController) *Session {
	return &Session{
		sock:            newSock(conn),
		scheduler:       scheduler.New(),
		inReqCh:         make(chan *Pdu, chanBuffSize),
		evtCh:           make(chan Evt, chanBuffSize),
		outReqCh:        make(chan *Req),
		outRespCh:       make(chan *Pdu, chanBuffSize),
		inRespCh:        make(chan *Resp, chanBuffSize),
		cfg:             cfg,
		speedController: speedController,
		outWinSema:      make(chan struct{}, 1),
		lastReading:     time.Now().Unix(),
		lastWriting:     time.Now().Unix(),
		reqsInFlight:    make(map[uint32]*Req),
	}
}

func (s *Session) Run(ctx context.Context) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.scheduler.Every(time.Second, func() {
			now := time.Now()

			if now.Unix()-atomic.LoadInt64(&s.lastReading) >= atomic.LoadInt64(&s.cfg.SilenceTimeoutSec) {
				s.logEvt(Warning, fmt.Sprintf("silence timeout [%v] exceeded. Socket closing...",
					atomic.LoadInt64(&s.cfg.SilenceTimeoutSec)))
				if err := s.sock.close(); err != nil {
					s.logEvt(Error, fmt.Sprintf("can't close socket: [%v]", err))
					s.errEvt(err)
				}
				return
			}

			if s.cfg.EnquireLinkEnabled &&
				now.Unix()-atomic.LoadInt64(&s.lastWriting) >= atomic.LoadInt64(&s.cfg.EnquireLinkIntervalSec) {
				select {
				case s.outReqCh <- &Req{
					Pdu: NewPdu(EnquireLink),
				}:
				default:
				}
			}
		})

		s.scheduler.Run(ctx)

		s.logEvt(Debug, "goroutine handling scheduler completed")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.speedController.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleIncomingPdus(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleOutgoingReqs(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleOutgoingResponses(ctx)
	}()

	select {
	case <-ctx.Done():
		if err := s.sock.close(); err != nil {
			s.logEvt(Error, fmt.Sprintf("can't close socket: [%v]", err))
			s.errEvt(err)
		}
	}

	wg.Wait()

	for _, r := range s.reqsInFlight {
		_r := r
		s.inRespCh <- &Resp{
			Err: ErrClosed,
			Req: _r,
		}
	}

	s.logEvt(Debug, "session completed")

	close(s.evtCh)
	close(s.inRespCh)
	close(s.inReqCh)
}

func (s *Session) InRespCh() <-chan *Resp {
	return s.inRespCh
}

func (s *Session) InReqCh() <-chan *Pdu {
	return s.inReqCh
}

func (s *Session) InEvtCh() <-chan Evt {
	return s.evtCh
}

func (s *Session) OutReqCh() chan<- *Req {
	return s.outReqCh
}

func (s *Session) OutRespCh() chan<- *Pdu {
	return s.outRespCh
}

func (s *Session) handleOutgoingResponses(ctx context.Context) {
	for {
		select {
		case r := <-s.outRespCh:
			atomic.AddInt32(&s.inWin, -1)

			err := s.sock.write(r)

			if err != nil {
				s.logEvt(Error, fmt.Sprintf("can't write pdu [%v] to socket: [%v]", r, err))
				s.errEvt(err)
			} else {
				atomic.StoreInt64(&s.lastWriting, time.Now().Unix())
				s.logEvt(Debug, fmt.Sprintf("sent pdu: [%v]", r))
			}
		case <-ctx.Done():
			for range s.outRespCh {
			}
			s.logEvt(Debug, fmt.Sprintf("goroutine handling outgoing responses completed"))
			return
		}
	}
}

func (s *Session) handleIncomingPdus(ctx context.Context) {
	defer s.logEvt(Debug, fmt.Sprintf("goroutine handling incoming pdus completed"))

	for {
		pdu, err := s.sock.read()

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			s.logEvt(Error, fmt.Sprintf("can't read pdu from socket: [%v]", err))
			s.errEvt(err)

			if errors.Is(err, io.EOF) {
				break
			} else {
				continue
			}
		} else {
			s.logEvt(Debug, fmt.Sprintf("received pdu: [%v]", pdu))
		}

		now := time.Now()
		atomic.StoreInt64(&s.lastReading, now.Unix())

		if pdu.IsReq() {
			if atomic.AddInt32(&s.inWin, 1) <= atomic.LoadInt32(&s.cfg.WinLimit) {
				if err := s.speedController.In(); err == errThrottling {
					resp, err := pdu.CreateResp(EsmeRThrottled)

					if err != nil {
						s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
						s.errEvt(err)
					} else {
						s.outRespCh <- resp
					}
				} else if err != nil {
					s.logEvt(Error, fmt.Sprintf("speed_controller.In returned error: [%v]", err))
					s.errEvt(err)

					resp, err := pdu.CreateResp(EsmeRSysErr)

					if err != nil {
						s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
						s.errEvt(err)
					} else {
						s.outRespCh <- resp
					}
				} else {
					if pdu.Id == EnquireLink {
						resp, err := pdu.CreateResp(EsmeROk)

						if err != nil {
							s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
							s.errEvt(err)
						} else {
							s.outRespCh <- resp
						}
					} else {
						s.inReqCh <- pdu
					}
				}
			} else {
				resp, err := pdu.CreateResp(EsmeRThrottled)

				if err != nil {
					s.logEvt(Error, fmt.Sprintf("can't create resp for pdu: [%v]", pdu))
					s.errEvt(err)
				} else {
					s.outRespCh <- resp
				}
			}
		} else {
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()

				if pdu.Status == EsmeRThrottled {
					s.lastThrottle = now
				}

				if req, ok := s.reqsInFlight[pdu.Seq]; ok {
					req.j.Cancel()
					if atomic.AddInt32(&s.outWin, -1) < atomic.LoadInt32(&s.cfg.WinLimit) {
						select {
						case <-s.outWinSema:
						default:
						}
					}

					s.inRespCh <- &Resp{
						Pdu: pdu,
						Req: req,
					}

					delete(s.reqsInFlight, pdu.Seq)
				} else {
					s.logEvt(Warning, fmt.Sprintf("received unexpected pdu: [%v]", pdu))
				}
			}()
		}
	}
}

func (s *Session) handleOutgoingReqs(ctx context.Context) {
	defer func() {
		for r := range s.outReqCh {
			_r := r
			s.inRespCh <- &Resp{
				Err: ErrClosed,
				Req: _r,
			}
		}
		s.logEvt(Debug, fmt.Sprintf("goroutine handling outgoing requests completed"))
	}()

	var seq uint32

	for {
		select {
		case s.outWinSema <- struct{}{}:
			select {
			case r := <-s.outReqCh:
				if err := s.speedController.Out(); err != nil {
					s.logEvt(Error, fmt.Sprintf("speed_controller.Out returned err: [%v]", err))
					s.errEvt(err)
					s.inRespCh <- &Resp{
						Err: err,
						Req: r,
					}
				} else {
					seq++
					r.Pdu.Seq = seq

					err := s.sock.write(r.Pdu)

					if err != nil {
						s.logEvt(Error, fmt.Sprintf("can't write pdu [%v] to socket: [%v]", r.Pdu, err))
						s.errEvt(err)
						s.inRespCh <- &Resp{
							Err: err,
							Req: r,
						}
					} else {
						s.logEvt(Debug, fmt.Sprintf("sent pdu: [%v]", r.Pdu))

						now := time.Now()

						atomic.StoreInt64(&s.lastWriting, now.Unix())

						s.mu.Lock()
						s.reqsInFlight[seq] = r
						throttlePause := time.Second*time.Duration(atomic.LoadInt32(&s.cfg.ThrottlePauseSec)) - now.Sub(s.lastThrottle)
						s.mu.Unlock()

						_seq := seq
						r.j = s.scheduler.Once(time.Second*time.Duration(atomic.LoadInt32(&s.cfg.ReqTimeoutSec)), func() {
							s.mu.Lock()
							defer s.mu.Unlock()

							if req, ok := s.reqsInFlight[_seq]; ok {
								if atomic.AddInt32(&s.outWin, -1) < atomic.LoadInt32(&s.cfg.WinLimit) {
									select {
									case <-s.outWinSema:
									default:
									}
								}

								s.inRespCh <- &Resp{
									Err: ErrTimeout,
									Req: req,
								}
								delete(s.reqsInFlight, _seq)
								s.logEvt(Warning, fmt.Sprintf("req timeout exceeded for pdu [%v]", req.Pdu))
							} else {
								s.logEvt(Warning, fmt.Sprintf("req timeout exceeded for seq [%v], but req not found", _seq))
							}
						})

						if atomic.AddInt32(&s.outWin, 1) < atomic.LoadInt32(&s.cfg.WinLimit) {
							select {
							case <-s.outWinSema:
							default:
							}
						}

						if throttlePause > 0 {
							select {
							case <-time.After(throttlePause):
							case <-ctx.Done():
								return
							}
						}
					}
				}
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
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
	s.cfg.EnquireLinkEnabled = cfg.EnquireLinkEnabled
	atomic.StoreInt64(&s.cfg.EnquireLinkIntervalSec, cfg.EnquireLinkIntervalSec)
	s.cfg.LogSeverity = cfg.LogSeverity

	s.speedController.SetRpsLimit(cfg.RpsLimit)
}
