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

type Session struct {
	sock                   *sock
	scheduler              *gocron.Scheduler
	inPduCh                chan *Pdu
	outReqCh               chan *req
	outRespCh              chan *Pdu
	closed                 chan struct{}
	rpsLimit               int32
	winLimit               int32
	throttlePauseSec       int32
	reqTimeoutSec          int32
	enquireLinkIntervalSec int64
}

func NewSession(conn net.Conn) *Session {
	s := &Session{
		sock:                   newSock(conn),
		scheduler:              gocron.NewScheduler(time.UTC),
		inPduCh:                make(chan *Pdu, chanBuffSize),
		outReqCh:               make(chan *req),
		outRespCh:              make(chan *Pdu, chanBuffSize),
		closed:                 make(chan struct{}),
		rpsLimit:               1,
		winLimit:               1,
		throttlePauseSec:       1,
		reqTimeoutSec:          2,
		enquireLinkIntervalSec: 15,
	}

	s.scheduler.StartAsync()

	outWinSema := make(chan struct{}, 1)
	outEnquireLinkReqCh := make(chan *req, 1)
	var outWin int32
	var inWin int32
	reqsInFlight := make(map[uint32]*req)
	var lastThrottle time.Time
	lastReading := time.Now().Unix()
	mu := sync.Mutex{}

	if _, err := s.scheduler.Every(uint64(s.enquireLinkIntervalSec)).Seconds().Do(func() {
		silenceIntervalSec := time.Now().Unix() - atomic.LoadInt64(&lastReading)
		enquireLinkIntervalSec := atomic.LoadInt64(&s.enquireLinkIntervalSec)
		if silenceIntervalSec > 2*enquireLinkIntervalSec {
			//TODO ...
			if err := s.sock.close(); err != nil {
				//TODO ...
			}
		} else if silenceIntervalSec > enquireLinkIntervalSec {
			select {
			case outEnquireLinkReqCh <- newReq(NewPdu(EnquireLink)):
			default:
			}
		}
	}); err != nil {
		//TODO ...
		if err := s.sock.close(); err != nil {
			//TODO ...
		}
	}

	//handling outgoing responses
	go func() {
		for {
			select {
			case r := <-s.outRespCh:
				err := s.sock.write(r)

				if err != nil {
					//TODO ...
					if err := s.sock.close(); err != nil {
						//TODO ...
					}
				}

				atomic.AddInt32(&inWin, -1)
			case <-s.closed:
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
				//TODO ...
				if err := s.sock.close(); err != nil {
					//TODO ...
				}
			}

			now := time.Now()

			if s := now.Unix(); s != sec {
				sec = s
				sentReqs = 0
			}

			sentReqs++

			var pause time.Duration
			if sentReqs >= atomic.LoadInt32(&s.rpsLimit) {
				pause = time.Duration(time.Second.Nanoseconds() - int64(now.Nanosecond()))
			}

			mu.Lock()
			reqsInFlight[seq] = r
			if throttlePause := time.Second*time.Duration(atomic.LoadInt32(&s.throttlePauseSec)) - now.Sub(lastThrottle);
				throttlePause > pause {
				pause = throttlePause
			}
			mu.Unlock()

			_seq := seq
			if r.j, err = s.scheduler.Every(uint64(s.reqTimeoutSec)).Seconds().Do(func() {
				mu.Lock()
				defer mu.Unlock()

				if req, ok := reqsInFlight[_seq]; ok {
					s.scheduler.RemoveByReference(req.j)

					if atomic.AddInt32(&outWin, -1) < atomic.LoadInt32(&s.winLimit) {
						select {
						case <-outWinSema:
						default:
						}
					}

					req.respCh <- &Resp{
						Err: ErrTimeout,
					}
					delete(reqsInFlight, _seq)
				} else {
					//TODO ...
				}
			}); err != nil {
				//TODO ...
			}

			if atomic.AddInt32(&outWin, 1) < atomic.LoadInt32(&s.winLimit) {
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
					return
				}
			case <-s.closed:
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

			if err != nil {
				//TODO ...
				break
			}

			select {
			case <-s.closed:
				return
			default:
			}

			now := time.Now()
			atomic.StoreInt64(&lastReading, now.Unix())

			if pdu.IsReq() {
				if atomic.AddInt32(&inWin, 1) <= atomic.LoadInt32(&s.winLimit) {
					if s := now.Unix(); s != sec {
						sec = s
						receivedReqs = 0
					}

					receivedReqs++

					if receivedReqs > atomic.LoadInt32(&s.rpsLimit) {
						resp, err := pdu.CreateResp(EsmeRThrottled)

						if err != nil {
							//TODO ...
						}

						s.SendResp(resp)
					} else {
						if pdu.Id == EnquireLink {
							resp, err := pdu.CreateResp(EsmeROk)

							if err != nil {
								//TODO ...
							}

							s.SendResp(resp)
						} else {
							s.inPduCh <- pdu
						}
					}
				} else {
					resp, err := pdu.CreateResp(EsmeRThrottled)

					if err != nil {
						//TODO ...
					}

					s.SendResp(resp)
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
						if atomic.AddInt32(&outWin, -1) < atomic.LoadInt32(&s.winLimit) {
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
						//TODO ...
					}
				}()
			}
		}
	}()

	return s
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
		//TODO ...
	}
}

func (s *Session) InPduCh() <-chan *Pdu {
	return s.inPduCh
}

func (s *Session) Close() error {
	s.scheduler.Clear()
	s.scheduler.Stop()
	close(s.closed)
	return s.sock.close()
}
