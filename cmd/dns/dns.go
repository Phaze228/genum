/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package dns

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	// "github.com/fatih/color"
	"github.com/miekg/dns"
	"github.com/phaze228/genum/utils"
	"github.com/spf13/cobra"
)

const (
	//		Defaults
	TIME_FORMAT          = "Mon, 2006-01-02 15:04:05"
	DEFAULT_THREAD_COUNT = 10
	//		DNS
	DEFAULT_DNS_PORT    = 53
	DEFAULT_NAME_SERVER = "8.8.8.8" // Googles DNS
	DEFAULT_OPTION      = "ANY"
)

var (
	gMu      = &sync.Mutex{}
	gResults = make([]string, 0)
	gErrata  = make([]error, 0)
)

const DNS_START_STRING = `
[DNS ENUMERATION]
 Domain: %s
 Nameserver: %s
 Time Start: %s
`
const DNS_END_STRING = `
[---FINISHED---]
 Time End: %s
 Took: %s
`

var DNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS Enumeration",
	Long: `
[DNS Enumeration]
	[-- REQUIRED --]
	-d <Domain to query DNS records>

	[-- OPTIONAL --]
	-n <Nameserver to resolve DNS queries>
	-t <DNS Record type>
	-T <Thread Count>
	-d <Timeout Duration>
	-p <Port for service>

	[-- EXAMPLES --]
	goEnum dns -d zonetransfer.me -n nsztm1.digi.ninja. -t AXFR
`,
	PreRunE: validateDNS,
	RunE:    executeDNS,
}

type DNS_Options struct {
	utils.Options
	Domain     string
	Nameserver string
	Type       string
	Port       int
	Threads    int
	Time       utils.Duration
	Verbose    bool
	SSL        bool
}

type Key struct{}

func init() {
	var duration utils.Duration = utils.Duration(time.Duration(3) * time.Second)
	DNSCmd.Flags().StringP("domain", "d", "", "domain name to check DNS of")
	DNSCmd.Flags().StringP("nameserver", "n", DEFAULT_NAME_SERVER, "nameserver to resolve queries")
	DNSCmd.Flags().StringP("type", "t", DEFAULT_OPTION, "DNS Enumeration Modes: [ANY, AXFR, A, AAAA... etc,]")
	// dnsCmd.Flags().StringP("domain", "D", "", "Domain to append to usernames: user@domain.com")
	// dnsCmd.Flags().StringP("from", "F", DEFAULT_EMAIL, "For use with RCPT, address of FROM sender: user@meow.com")
	DNSCmd.Flags().IntP("port", "p", DEFAULT_DNS_PORT, "Port the DNS Service runs on")
	DNSCmd.Flags().IntP("threads", "T", DEFAULT_THREAD_COUNT, "Thread Count: Default: 10")
	DNSCmd.Flags().VarP(&duration, "duration", "D", "Timeout: 3s, 10s...etc")
	DNSCmd.Flags().BoolP("verbose", "v", false, "Verbose output, prints nearly everything")
	DNSCmd.Flags().BoolP("ssl", "S", false, "Enable SSL")
}

func validateDNS(cmd *cobra.Command, args []string) error {
	var options = new(DNS_Options)
	err := options.AddRequired(cmd,
		"domain", &options.Domain,
	)
	if err != nil {
		return err
	}

	err = options.Add(cmd,
		"nameserver", &options.Nameserver,
		"type", &options.Type,
		"port", &options.Port,
		"threads", &options.Threads,
		"verbose", &options.Verbose,
		"ssl", &options.SSL,
	)
	if err != nil {
		return err
	}
	// options.Mode = parseMode(options.Mode)

	cmd.SetContext(context.WithValue(cmd.Context(), Key{}, options))
	return nil
}

func executeDNS(cmd *cobra.Command, args []string) error {
	validatedArgs := cmd.Context().Value(Key{})
	if validatedArgs == nil {
		return fmt.Errorf("[Command Line Options Error]")
	}
	opts, ok := validatedArgs.(*DNS_Options)
	if !ok {
		return fmt.Errorf("Invalid Type: %T", validatedArgs)

	}
	start_time := time.Now()
	fmt.Printf(DNS_START_STRING, opts.Domain, opts.Nameserver, start_time.Format(TIME_FORMAT))
	domain := dns.Fqdn(opts.Domain)
	ns := opts.Nameserver
	recordTypes := func() []uint16 {
		buf := make([]uint16, 0)
		test := strings.Split(opts.Type, ",")
		for _, t := range test {
			buf = append(buf, dns.StringToType[strings.TrimSpace(t)])
		}
		return buf
	}()
	fmt.Println("\n------------[PROGRESS]---------------------")
	recs := NewRecords()

	if slices.Contains(recordTypes, dns.TypeANY) {
		recs.CheckAllRecords(domain, ns, DNSRecTypes[:])
	} else {
		recs.CheckAllRecords(domain, ns, recordTypes)
	}
	end_time := time.Now()
	fmt.Printf(DNS_END_STRING, end_time.Format(TIME_FORMAT), end_time.Sub(start_time).String())
	return nil

}
