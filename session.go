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
	Pdu     *Pdu
	j       *scheduler.Job
	retries int32
	Ctx     interface{}
	Sent    time.Time
}

type Resp struct {
	Err      error
	Pdu      *Pdu
	Req      *Req
	Received time.Time
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

type InWinChangedEvt struct {
	value int32
}

func (e *InWinChangedEvt) String() string {
	return fmt.Sprintf("in window changed to [%v]", e.value)
}

func (e *InWinChangedEvt) Value() int32 {
	return e.value
}

type OutWinChangedEvt struct {
	value int32
}

func (e *OutWinChangedEvt) String() string {
	return fmt.Sprintf("out window changed to [%v]", e.value)
}

func (e *OutWinChangedEvt) Value() int32 {
	return e.value
}

type PduReceivedEvt struct {
	id     Id
	status Status
}

func (e *PduReceivedEvt) String() string {
	return fmt.Sprintf("received pdu:[%v][%v]", e.id, e.status)
}

func (e *PduReceivedEvt) Id() Id {
	return e.id
}

func (e *PduReceivedEvt) Status() Status {
	return e.status
}

type PduSentEvt struct {
	id     Id
	status Status
}

func (e *PduSentEvt) String() string {
	return fmt.Sprintf("sent pdu:[%v][%v]", e.id, e.status)
}

func (e *PduSentEvt) Id() Id {
	return e.id
}

func (e *PduSentEvt) Status() Status {
	return e.status
}

type SpeedController interface {
	Out(ctx context.Context) error
	In() error
	SetRpsLimit(in, out int32)
	Run(ctx context.Context)
}

var errThrottling = errors.New("throttling error")

type DefaultSpeedController struct {
	inRpsLimit               int32
	outRpsLimit              int32
	outIntervalNSec          int64
	inSec                    int64
	inReqs                   int32
	outReqsCh                chan struct{}
	stop                     chan struct{}
	runCount                 int32
	outSpeedControlAlgorithm OutSpeedControlAlgorithm
}

type OutSpeedControlAlgorithm int

const (
	Robust OutSpeedControlAlgorithm = iota
	Risky
)

func NewDefaultSpeedController(outSpeedControlAlgorithm OutSpeedControlAlgorithm) *DefaultSpeedController {
	return &DefaultSpeedController{
		inRpsLimit:               1,
		outRpsLimit:              1,
		outIntervalNSec:          int64(time.Second / 1),
		inSec:                    0,
		inReqs:                   0,
		outReqsCh:                make(chan struct{}),
		runCount:                 0,
		outSpeedControlAlgorithm: outSpeedControlAlgorithm,
	}
}

func (c *DefaultSpeedController) Out(ctx context.Context) error {
	select {
	case c.outReqsCh <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
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

	if inReqs > atomic.LoadInt32(&c.inRpsLimit) {
		return errThrottling
	} else {
		return nil
	}
}

func (c *DefaultSpeedController) SetRpsLimit(in, out int32) {
	atomic.StoreInt32(&c.inRpsLimit, in)
	atomic.StoreInt32(&c.outRpsLimit, out)
	atomic.StoreInt64(&c.outIntervalNSec, int64(time.Second/time.Duration(out)))
}

func (c *DefaultSpeedController) Run(ctx context.Context) {
	if atomic.AddInt32(&c.runCount, 1) == 1 {
		c.stop = make(chan struct{})

		if c.outSpeedControlAlgorithm == Risky {
			go func() {
				sec := time.Now().Unix()
				var reqs int32 = 0
				for {
					select {
					case <-c.outReqsCh:
						now := time.Now()
						if s := now.Unix(); s != sec {
							sec = s
							reqs = 1
						} else {
							reqs++
						}

						if reqs >= atomic.LoadInt32(&c.outRpsLimit) {
							select {
							case <-time.After(time.Duration(time.Second.Nanoseconds() - int64(now.Nanosecond()))):
							case <-c.stop:
								return
							}
						}
					case <-c.stop:
						return
					}
				}
			}()
		} else {
			go func() {
				timer := time.NewTimer(0)
				<-timer.C

				var lag int64 = 0
				var sent int64 = 0

				for {
					select {
					case <-c.outReqsCh:
						idealInterval := atomic.LoadInt64(&c.outIntervalNSec)
						now := time.Now().UnixNano()

						if time.Duration(now-sent) > time.Second {
							lag = 0
						}

						sent = now
						interval := idealInterval - lag
						if interval > 0 {
							timer.Reset(time.Duration(interval))
							select {
							case <-timer.C:
								lag += time.Now().UnixNano() - sent - idealInterval
							case <-c.stop:
								return
							}
						} else {
							lag -= idealInterval
						}
					case <-c.stop:
						return
					}
				}
			}()
		}
	}

	select {
	case <-ctx.Done():
		if atomic.AddInt32(&c.runCount, -1) == 0 {
			close(c.stop)
		}
	}
}

type SessionConfig struct {
	InRpsLimit              int32
	OutRpsLimit             int32
	WinLimit                int32
	ThrottlePauseSec        int32
	ThrottleRetriesMaxCount int32
	ReqTimeoutSec           int32
	EnquireLinkEnabled      bool
	EnquireLinkIntervalSec  int64
	SilenceTimeoutSec       int64
	LogSeverity             Severity
}

func NewDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		InRpsLimit:              1,
		OutRpsLimit:             1,
		WinLimit:                1,
		ThrottlePauseSec:        1,
		ThrottleRetriesMaxCount: 3,
		ReqTimeoutSec:           2,
		EnquireLinkEnabled:      false,
		EnquireLinkIntervalSec:  15,
		SilenceTimeoutSec:       60,
		LogSeverity:             Info,
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
	retriesCh       chan *Req
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
	speedController.SetRpsLimit(cfg.InRpsLimit, cfg.OutRpsLimit)
	return &Session{
		sock:            newSock(conn),
		scheduler:       scheduler.New(),
		inReqCh:         make(chan *Pdu, chanBuffSize),
		evtCh:           make(chan Evt, chanBuffSize),
		outReqCh:        make(chan *Req),
		outRespCh:       make(chan *Pdu, chanBuffSize),
		inRespCh:        make(chan *Resp, chanBuffSize),
		retriesCh:       make(chan *Req, chanBuffSize),
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
				s.logEvt(Warning, func() string {
					return fmt.Sprintf("silence timeout [%v] exceeded. Socket closing...",
						atomic.LoadInt64(&s.cfg.SilenceTimeoutSec))
				})
				if err := s.sock.close(); err != nil {
					s.logEvt(Error, func() string {
						return fmt.Sprintf("can't close socket: [%v]", err)
					})
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

		s.logEvt(Debug, func() string {
			return "goroutine handling scheduler completed"
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.speedController.Run(ctx)
		s.logEvt(Debug, func() string {
			return "goroutine handling speed controller completed"
		})
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
			s.logEvt(Error, func() string {
				return fmt.Sprintf("can't close socket: [%v]", err)
			})
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

	s.logEvt(Debug, func() string {
		return "session completed"
	})

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
	defer s.logEvt(Debug, func() string {
		return fmt.Sprintf("goroutine handling outgoing responses completed")
	})

	for {
		select {
		case pdu, ok := <-s.outRespCh:
			if !ok {
				return
			}

			s.inWinChangedEvt(atomic.AddInt32(&s.inWin, -1))

			err := s.sock.write(pdu)

			if err != nil {
				s.logEvt(Error, func() string {
					return fmt.Sprintf("can't write pdu [%v] to socket: [%v]", pdu, err)
				})
				s.errEvt(err)
			} else {
				atomic.StoreInt64(&s.lastWriting, time.Now().Unix())
				s.logEvt(Debug, func() string {
					return fmt.Sprintf("sent pdu: [%v]", pdu)
				})
				s.pduSentEvt(pdu)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Session) handleIncomingPdus(ctx context.Context) {
	defer s.logEvt(Debug, func() string {
		return fmt.Sprintf("goroutine handling incoming pdus completed")
	})

	for {
		pdu, err := s.sock.read()

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			s.logEvt(Error, func() string {
				return fmt.Sprintf("can't read pdu from socket: [%v]", err)
			})
			s.errEvt(err)

			if errors.Is(err, io.EOF) {
				break
			} else {
				continue
			}
		} else {
			s.logEvt(Debug, func() string {
				return fmt.Sprintf("received pdu: [%v]", pdu)
			})
			s.pduReceivedEvt(pdu)
		}

		now := time.Now()
		atomic.StoreInt64(&s.lastReading, now.Unix())

		if pdu.IsReq() {
			inWin := atomic.AddInt32(&s.inWin, 1)
			s.inWinChangedEvt(inWin)
			if inWin <= atomic.LoadInt32(&s.cfg.WinLimit) {
				if err := s.speedController.In(); err == errThrottling {
					resp, err := pdu.CreateResp(EsmeRThrottled)

					if err != nil {
						s.logEvt(Error, func() string {
							return fmt.Sprintf("can't create resp for pdu: [%v]", pdu)
						})
						s.errEvt(err)
					} else {
						select {
						case s.outRespCh <- resp:
						case <-ctx.Done():
							return
						}
					}
				} else if err != nil {
					s.logEvt(Error, func() string {
						return fmt.Sprintf("speed_controller.In returned error: [%v]", err)
					})
					s.errEvt(err)

					resp, err := pdu.CreateResp(EsmeRSysErr)

					if err != nil {
						s.logEvt(Error, func() string {
							return fmt.Sprintf("can't create resp for pdu: [%v]", pdu)
						})
						s.errEvt(err)
					} else {
						select {
						case s.outRespCh <- resp:
						case <-ctx.Done():
							return
						}
					}
				} else {
					if pdu.id == EnquireLink {
						resp, err := pdu.CreateResp(EsmeROk)

						if err != nil {
							s.logEvt(Error, func() string {
								return fmt.Sprintf("can't create resp for pdu: [%v]", pdu)
							})
							s.errEvt(err)
						} else {
							select {
							case s.outRespCh <- resp:
							case <-ctx.Done():
								return
							}
						}
					} else {
						s.inReqCh <- pdu
					}
				}
			} else {
				resp, err := pdu.CreateResp(EsmeRThrottled)

				if err != nil {
					s.logEvt(Error, func() string {
						return fmt.Sprintf("can't create resp for pdu: [%v]", pdu)
					})
					s.errEvt(err)
				} else {
					select {
					case s.outRespCh <- resp:
					case <-ctx.Done():
						return
					}
				}
			}
		} else {
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()

				if pdu.status == EsmeRThrottled {
					s.lastThrottle = now
				}

				if req, ok := s.reqsInFlight[pdu.seq]; ok {
					req.j.Cancel()
					outWin := atomic.AddInt32(&s.outWin, -1)
					s.outWinChangedEvt(outWin)
					if outWin < atomic.LoadInt32(&s.cfg.WinLimit) {
						select {
						case <-s.outWinSema:
						default:
						}
					}

					if pdu.status == EsmeRThrottled && req.retries < atomic.LoadInt32(&s.cfg.ThrottleRetriesMaxCount) {
						select {
						case s.retriesCh <- req:
						default:
							s.errEvt(fmt.Errorf("queue of retries full: %v", len(s.retriesCh)))
							s.inRespCh <- &Resp{
								Pdu:      pdu,
								Req:      req,
								Received: now,
							}
						}
					} else {
						s.inRespCh <- &Resp{
							Pdu:      pdu,
							Req:      req,
							Received: now,
						}
					}

					delete(s.reqsInFlight, pdu.seq)
				} else {
					s.logEvt(Warning, func() string {
						return fmt.Sprintf("received unexpected pdu: [%v]", pdu)
					})
				}
			}()
		}
	}
}

func (s *Session) handleOutgoingReqs(ctx context.Context) {
	defer s.logEvt(Debug, func() string {
		return fmt.Sprintf("goroutine handling outgoing requests completed")
	})

	var seq uint32

	for {
		select {
		case s.outWinSema <- struct{}{}:
			select {
			case r := <-s.retriesCh:
				s.handleOutgoingReq(r, &seq, ctx)
			case r := <-s.outReqCh:
				s.handleOutgoingReq(r, &seq, ctx)
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Session) handleOutgoingReq(r *Req, seq *uint32, ctx context.Context) {
	if err := s.speedController.Out(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		s.logEvt(Error, func() string {
			return fmt.Sprintf("speed_controller.Out returned err: [%v]", err)
		})
		s.errEvt(err)
		s.inRespCh <- &Resp{
			Err: err,
			Req: r,
		}
	} else {
		*seq++
		_seq := *seq
		r.Pdu.SetSeq(_seq)

		now := time.Now()

		s.mu.Lock()
		r.retries++
		r.j = s.scheduler.Once(time.Second*time.Duration(atomic.LoadInt32(&s.cfg.ReqTimeoutSec)), func() {
			s.mu.Lock()
			defer s.mu.Unlock()

			if req, ok := s.reqsInFlight[_seq]; ok {
				outWin := atomic.AddInt32(&s.outWin, -1)
				s.outWinChangedEvt(outWin)
				if outWin < atomic.LoadInt32(&s.cfg.WinLimit) {
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
				s.logEvt(Warning, func() string {
					return fmt.Sprintf("req timeout exceeded for pdu [%v]", req.Pdu)
				})
			} else {
				s.logEvt(Warning, func() string {
					return fmt.Sprintf("req timeout exceeded for seq [%v], but req not found", _seq)
				})
			}
		})
		s.reqsInFlight[_seq] = r
		throttlePause := time.Second*time.Duration(atomic.LoadInt32(&s.cfg.ThrottlePauseSec)) - now.Sub(s.lastThrottle)
		r.Sent = now
		s.mu.Unlock()

		err := s.sock.write(r.Pdu)

		if err != nil {
			s.mu.Lock()
			r.j.Cancel()
			delete(s.reqsInFlight, _seq)
			s.mu.Unlock()

			s.logEvt(Error, func() string {
				return fmt.Sprintf("can't write pdu [%v] to socket: [%v]", r.Pdu, err)
			})
			s.errEvt(err)
			s.inRespCh <- &Resp{
				Err: err,
				Req: r,
			}
		} else {
			s.logEvt(Debug, func() string {
				return fmt.Sprintf("sent pdu: [%v]", r.Pdu)
			})
			s.pduSentEvt(r.Pdu)
			atomic.StoreInt64(&s.lastWriting, now.Unix())

			outWin := atomic.AddInt32(&s.outWin, 1)
			s.outWinChangedEvt(outWin)
			if outWin < atomic.LoadInt32(&s.cfg.WinLimit) {
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
}

func (s *Session) logEvt(severity Severity, msgCreator func() string) {
	if severity < s.cfg.LogSeverity {
		return
	}

	s.evtCh <- &LogEvt{severity: severity, msg: fmt.Sprintf("[%v]: %v", s.RemoteAddr(), msgCreator())}
}

func (s *Session) errEvt(err error) {
	s.evtCh <- &ErrEvt{err: err}
}

func (s *Session) inWinChangedEvt(value int32) {
	s.evtCh <- &InWinChangedEvt{value: value}
}

func (s *Session) outWinChangedEvt(value int32) {
	s.evtCh <- &OutWinChangedEvt{value: value}
}

func (s *Session) pduReceivedEvt(pdu *Pdu) {
	s.evtCh <- &PduReceivedEvt{id: pdu.id, status: pdu.status}
}

func (s *Session) pduSentEvt(pdu *Pdu) {
	s.evtCh <- &PduSentEvt{id: pdu.id, status: pdu.status}
}

func (s *Session) RemoteAddr() net.Addr {
	return s.sock.c.RemoteAddr()
}

func (s *Session) SetConfig(cfg *SessionConfig) {
	atomic.StoreInt32(&s.cfg.InRpsLimit, cfg.InRpsLimit)
	atomic.StoreInt32(&s.cfg.OutRpsLimit, cfg.OutRpsLimit)
	atomic.StoreInt32(&s.cfg.WinLimit, cfg.WinLimit)
	atomic.StoreInt32(&s.cfg.ThrottlePauseSec, cfg.ThrottlePauseSec)
	atomic.StoreInt32(&s.cfg.ThrottleRetriesMaxCount, cfg.ThrottleRetriesMaxCount)
	atomic.StoreInt32(&s.cfg.ReqTimeoutSec, cfg.ReqTimeoutSec)
	s.cfg.EnquireLinkEnabled = cfg.EnquireLinkEnabled
	atomic.StoreInt64(&s.cfg.EnquireLinkIntervalSec, cfg.EnquireLinkIntervalSec)
	atomic.StoreInt64(&s.cfg.SilenceTimeoutSec, cfg.SilenceTimeoutSec)
	s.cfg.LogSeverity = cfg.LogSeverity

	s.speedController.SetRpsLimit(cfg.InRpsLimit, cfg.OutRpsLimit)
}

func (s *Session) GetConfig() *SessionConfig {
	return &SessionConfig{
		InRpsLimit:              atomic.LoadInt32(&s.cfg.InRpsLimit),
		OutRpsLimit:             atomic.LoadInt32(&s.cfg.OutRpsLimit),
		WinLimit:                atomic.LoadInt32(&s.cfg.WinLimit),
		ThrottlePauseSec:        atomic.LoadInt32(&s.cfg.ThrottlePauseSec),
		ThrottleRetriesMaxCount: atomic.LoadInt32(&s.cfg.ThrottleRetriesMaxCount),
		ReqTimeoutSec:           atomic.LoadInt32(&s.cfg.ReqTimeoutSec),
		EnquireLinkEnabled:      s.cfg.EnquireLinkEnabled,
		EnquireLinkIntervalSec:  atomic.LoadInt64(&s.cfg.EnquireLinkIntervalSec),
		SilenceTimeoutSec:       atomic.LoadInt64(&s.cfg.SilenceTimeoutSec),
		LogSeverity:             s.cfg.LogSeverity,
	}
}
