package main

import (
	"errors"
	"flag"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
)

type string2Slice struct{ arr []string }

func (s *string2Slice) String() string {
	return strings.Join(s.arr, " ")
}

func (s *string2Slice) Set(str string) error {

	log.Println(str)

	if str == "" {
		return errors.New("empty input")
	}
	s.arr = strings.Split(str, " ")
	return nil
}

func main() {
	h := flag.String("h", "", "dns host")
	q := string2Slice{}
	flag.Var(&q, "q", "domain name")

	flag.Parse()

	if *h == "" {
		return
	}

	hosts := q.arr

	log.Println(q.arr)

	wg := sync.WaitGroup{}
	sTime := time.Now()

	client := dns.Client{
		Net:     "udp",
		UDPSize: 1500,
		Timeout: 3 * time.Second,
		//SingleInflight: true,
	}

	errorCount := int32(0)

	for _, host := range hosts {
		if host == "" {
			continue
		}

		wg.Add(1)

		go func(host string) {
			defer wg.Done()

			q := dns.Question{
				Name:   host + ".",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			}

			msg := &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:                 0,
					Response:           false,
					Opcode:             0,
					Authoritative:      false,
					Truncated:          false,
					RecursionDesired:   true,
					RecursionAvailable: false,
					Zero:               false,
					AuthenticatedData:  false,
					CheckingDisabled:   false,
					Rcode:              0,
				},
				Compress: false,
				Question: []dns.Question{q},
				Answer:   make([]dns.RR, 0),
				Ns:       make([]dns.RR, 0),
				Extra:    make([]dns.RR, 0),
			}

			r, _, e := client.Exchange(msg, *h)
			if e != nil {
				log.Println("dns query error", e)
				atomic.AddInt32(&errorCount, 1)
				return
			}

			if r != nil {
				log.Println(host, r.Answer[0].String())
			}
		}(host)
	}

	wg.Wait()
	eTime := time.Now()
	log.Println("using seconds", eTime.Unix()-sTime.Unix(), errorCount)
}
