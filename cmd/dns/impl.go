package dns

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/miekg/dns"
)

var DNSRecTypes = [...]uint16{
	dns.TypeSOA,
	dns.TypeNS,
	dns.TypeA,
	dns.TypeAAAA,
	dns.TypeSRV,
	dns.TypeMX,
	dns.TypeCNAME,
	dns.TypeTXT,
	dns.TypeMD,
	dns.TypeMF,
	dns.TypeMB,
	dns.TypeMG,
	dns.TypeMR,
	dns.TypeNULL,
	dns.TypePTR,
	dns.TypeHINFO,
	dns.TypeMINFO,
	dns.TypeRP,
	dns.TypeAFSDB,
	dns.TypeX25,
	dns.TypeISDN,
	dns.TypeRT,
	dns.TypeNSAPPTR,
	dns.TypeSIG,
	dns.TypeKEY,
	dns.TypePX,
	dns.TypeGPOS,
	dns.TypeLOC,
	dns.TypeNXT,
	dns.TypeEID,
	dns.TypeNIMLOC,
	dns.TypeATMA,
	dns.TypeNAPTR,
	dns.TypeKX,
	dns.TypeCERT,
	dns.TypeDNAME,
	dns.TypeOPT,
	dns.TypeAPL,
	dns.TypeDS,
	dns.TypeSSHFP,
	dns.TypeIPSECKEY,
	dns.TypeRRSIG,
	dns.TypeNSEC,
	dns.TypeDNSKEY,
	dns.TypeDHCID,
	dns.TypeNSEC3,
	dns.TypeNSEC3PARAM,
	dns.TypeTLSA,
	dns.TypeSMIMEA,
	dns.TypeHIP,
	dns.TypeNINFO,
	dns.TypeRKEY,
	dns.TypeTALINK,
	dns.TypeCDS,
	dns.TypeCDNSKEY,
	dns.TypeOPENPGPKEY,
	dns.TypeCSYNC,
	dns.TypeZONEMD,
	dns.TypeSVCB,
	dns.TypeHTTPS,
	dns.TypeSPF,
	dns.TypeUINFO,
	dns.TypeUID,
	dns.TypeGID,
	dns.TypeUNSPEC,
	dns.TypeNID,
	dns.TypeL32,
	dns.TypeL64,
	dns.TypeLP,
	dns.TypeEUI48,
	dns.TypeEUI64,
	dns.TypeNXNAME,
	dns.TypeURI,
	dns.TypeCAA,
	dns.TypeAVC,
	dns.TypeAMTRELAY,
	dns.TypeAXFR,
}

var domainTypes = [...]uint16{
	dns.TypeNS,
	dns.TypeSOA,
	dns.TypeCNAME,
}

type DNSTask struct {
	Domain     string
	Nameserver string
}

type Records struct {
	Data map[uint16][]dns.RR
	mu   sync.Mutex
}

func NewRecords() *Records {
	return &Records{
		Data: make(map[uint16][]dns.RR),
	}
}

func (r *Records) Print() {
	for _, recordType := range DNSRecTypes {
		if len(r.Data[recordType]) == 0 {
			continue
		}
		color.Yellow("  [ %s ]", dns.TypeToString[recordType])
		length := len(r.Data[recordType])
		for i, record := range r.Data[recordType] {
			if i == length-1 {
				fmt.Printf("  |_____%s\n\n", record)
				break
			}
			fmt.Printf("  | \t%s\n", record)
		}
	}
}

func (r *Records) CheckAllRecords(domain, nameserver string, recordsToCheck []uint16) {
	tasks := make(chan uint16, 100)
	var wg sync.WaitGroup

	go func() {
		for _, t := range recordsToCheck {
			tasks <- t
		}
		close(tasks)
	}()
	for i := 0; i < DEFAULT_THREAD_COUNT; i++ {
		wg.Add(1)
		go r.checkRecords(domain, nameserver, tasks, &wg)
	}

	wg.Wait()

	color.Blue("[ Record Check Results ]")
	r.Print()
	if slices.Contains(recordsToCheck, dns.TypeAXFR) {
		axfr := newAXFR()
		for _, ns := range r.Data[dns.TypeNS] {
			last := strings.Split(ns.String(), "\t")
			nsEntry := last[len(last)-1]
			axfr.AddTask(DNSTask{domain, nsEntry}, true)
			// select {
			// case axfr.tasks <- DNSTask{domain, nsEntry}:
			// 	atomic.AddInt32(&counter, 1)
			// default:
			// 	panic("channel closed")
			// }

		}
		axfr.ZoneTransfer(domain, nameserver)
	}
}

func (r *Records) checkRecords(domain, nameserver string, tasks chan uint16, wg *sync.WaitGroup) {
	defer wg.Done()
	for recordType := range tasks {
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(domain), recordType)
		client := new(dns.Client)
		in, _, err := client.Exchange(msg, net.JoinHostPort(nameserver, "53"))
		if err != nil {
			fmt.Printf("[ERROR] DNS Lookup Failure: %v\n", err)
			return
		}
		if len(in.Answer) > 0 {
			for _, answer := range in.Answer {
				r.mu.Lock()
				r.Data[recordType] = append(r.Data[recordType], answer)
				r.mu.Unlock()
			}
		}

	}
}

type Transfers map[string]*Records

type AXFR struct {
	transfers map[string]*Records
	tasks     chan DNSTask
	results   chan DNSTask
	visited   *sync.Map
	failed    []string
	mu        sync.Mutex
	Counter   *int32
}

func (a *AXFR) printTransfers() {
	for d, rec := range a.transfers {
		rec.mu.Lock()
		if len(rec.Data) == 0 {
			continue
		}
		color.Red("[------ %s ------]", d)
		rec.Print()
		rec.mu.Unlock()
	}

}

func newAXFR() AXFR {
	return AXFR{make(Transfers), make(chan DNSTask, 200), make(chan DNSTask, 200), &sync.Map{}, make([]string, 0), sync.Mutex{}, new(int32)}
}

func (a *AXFR) ZoneTransfer(domain, ns string) {
	var wg sync.WaitGroup
	const WORKERS = 5

	// atomic.AddInt32(taskCounter, 1)
	a.AddTask(DNSTask{domain, ns}, true)
	// a.tasks <- DNSTask{domain, ns}

	for i := 0; i < WORKERS; i++ {
		wg.Add(1)
		go a.recurseTransfer(&wg)
	}

	// Add new tasks
	go func() {
		for newTask := range a.results {
			a.tasks <- newTask

		}
	}()

	var closeOnce sync.Once
	go func() {
		for {
			if atomic.LoadInt32(a.Counter) == 0 {
				closeOnce.Do(func() {
					close(a.tasks)
					close(a.results)

				})
				break
			}

		}
	}()
	wg.Wait()
	color.Blue("[ Zone Transfer Results ]")
	a.printTransfers()

}

func (a *AXFR) recurseTransfer(wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range a.tasks {
		dom := dns.Fqdn(task.Domain)
		ns := dns.Fqdn(task.Nameserver)
		// fmt.Printf("Starting %s @ %s\n ", dom, ns)
		DN := dom + "@" + ns
		a.mu.Lock()
		if _, visit := a.visited.Load(DN); visit {
			atomic.AddInt32(a.Counter, -1)
			continue
		}
		a.mu.Unlock()
		a.visited.Store(DN, true)
		a.transfers[DN] = NewRecords()
		t := new(dns.Transfer)
		msg := new(dns.Msg)
		msg.SetAxfr(dom)

		stream, err := t.In(msg, net.JoinHostPort(ns, "53"))
		if err != nil {
			a.mu.Lock()
			a.failed = append(a.failed, fmt.Sprintf("[AXFR Fail] - %s @ %s | %v", dom, ns, err))
			a.mu.Unlock()
			atomic.AddInt32(a.Counter, -1)
			continue
		}
		for r := range stream {
			if r.Error != nil {
				a.mu.Lock()
				a.failed = append(a.failed, fmt.Sprintf("[AXFR Error] - %s @ %s | %v", dom, ns, err))
				a.mu.Unlock()
				continue
			}
			for _, answer := range r.RR {
				a.transfers[DN].mu.Lock()
				a.transfers[DN].Data[answer.Header().Rrtype] = append(a.transfers[DN].Data[answer.Header().Rrtype], answer)
				a.transfers[DN].mu.Unlock()
			}
		}

		nsRecs, hasNS := a.transfers[DN].Data[dns.TypeNS]
		for _, types := range domainTypes {
			a.transfers[DN].mu.Lock()
			for _, r := range a.transfers[DN].Data[types] {
				if hasNS {
					for _, names := range nsRecs {
						a.AddTask(DNSTask{dom, names.Header().Name}, false)
						// atomic.AddInt32(counter, 1)
						// a.results <- DNSTask{dom, names.Header().Name}
					}
				}
				a.AddTask(DNSTask{r.Header().Name, ns}, false)
				// atomic.AddInt32(counter, 1)
				// a.results <- DNSTask{r.Header().Name, ns}
			}
			a.transfers[DN].mu.Unlock()
		}
		atomic.AddInt32(a.Counter, -1)
	}

}

func (a *AXFR) AddTask(task DNSTask, taskchan bool) {
	if taskchan {
		select {
		case a.tasks <- task:
			atomic.AddInt32(a.Counter, 1)
		default:
			return
		}
	} else {
		select {
		case a.results <- task:
			atomic.AddInt32(a.Counter, 1)
		default:
			return
		}
	}

}
