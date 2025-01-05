/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package mail

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/phaze228/genum/utils"
	"github.com/spf13/cobra"
)

const (
	// Defaults
	SMTP_PORT     = 25
	THREAD_COUNT  = 10
	DEFAULT_EMAIL = "user@meow.com"
	TIME_FORMAT   = "Mon, 2006-01-02 15:04:05"
	// SMTP
	VRFY      = "VRFY"
	EXPN      = "EXPN"
	RCPT      = "RCPT"
	RN        = "\r\n"
	HELO      = "HELO x" + RN
	FROM_MAIL = "MAIL FROM:"
	RCPT_TO   = "RCPT TO:"
)

var (
	gMu      = &sync.Mutex{}
	gResults = make([]string, 0)
	gErrata  = make([]error, 0)
)

const START_STRING = `
[SMTP USER ENUMERATION]
 Mode: %s
 Hosts: %s
 Users: %s
 Time Start: %s
`
const END_STRING = `
[---FINISHED---]
 Time End: %s
 Took: %s
`

// SmtpCmd represents the smtp command
var SmtpCmd = &cobra.Command{
	Use:   "smtp",
	Short: "Enumerates valid users for SMTP mail servers",
	Long: `
[SMTP USER ENUMERATION]
	[-- REQUIRED --]
	-U <User or file of users>
	-H <Host or file of hosts>

	[-- OPTIONAL --]
	-M <Method of Enumeration>
	-F <From Address (for RCPT)>
	-D <Domain to append to usernames>
	-T <Amount of Threads to run>
	-S <Toggle SSL>
	-d <Duration of timeout>
	-p <Port which Service acts on>

	[-- EXAMPLES --]
	goEnum smpt -U john -H 10.129.90.102 -M VRFY
`,
	PreRunE: validateSMTP,
	RunE:    executeSMTP,
}

type SMTP_Options struct {
	utils.Options
	Users   string
	Hosts   string
	Mode    string
	Domain  string
	From    string
	Port    int
	Threads int
	Time    utils.Duration
	Verbose bool
	SSL     bool
}
type Key struct{}

func init() {
	var duration utils.Duration = utils.Duration(time.Duration(3) * time.Second)
	SmtpCmd.Flags().StringP("users", "U", "", "username or file with list of usernames")
	SmtpCmd.Flags().StringP("hosts", "H", "", "hosts or file with list of hosts")
	SmtpCmd.Flags().StringP("mode", "M", VRFY, "SMTP Enumeration Modes: [VRFY/V/v | EXPN/E/e | RCPT/R/r]")
	SmtpCmd.Flags().StringP("domain", "D", "", "Domain to append to usernames: user@domain.com")
	SmtpCmd.Flags().StringP("from", "F", DEFAULT_EMAIL, "For use with RCPT, address of FROM sender: user@meow.com")
	SmtpCmd.Flags().IntP("port", "p", SMTP_PORT, "Port the SMTP Service runs on")
	SmtpCmd.Flags().IntP("threads", "T", THREAD_COUNT, "Thread Count: Default: 10")
	SmtpCmd.Flags().VarP(&duration, "duration", "d", "Timeout: 3s, 10s...etc")
	SmtpCmd.Flags().BoolP("verbose", "v", false, "Verbose output, prints nearly everything")
	SmtpCmd.Flags().BoolP("ssl", "S", false, "Enable SSL")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// smtpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// smtpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func executeSMTP(cmd *cobra.Command, args []string) error {
	validatedArgs := cmd.Context().Value(Key{})
	if validatedArgs == nil {
		return fmt.Errorf("[Command Line Options Error]")
	}
	opts, ok := validatedArgs.(*SMTP_Options)
	if !ok {
		return fmt.Errorf("Invalid Type: %T", validatedArgs)

	}
	usernames := make([]string, 0)
	hostnames := make([]string, 0)

	utils.AppendFileContentsOrString(opts.Users, &usernames)
	utils.AppendFileContentsOrString(opts.Hosts, &hostnames)

	var wg sync.WaitGroup
	wg.Add(opts.Threads)

	queryChan := make(chan string)

	start_time := time.Now()
	fmt.Printf(START_STRING, opts.Mode, opts.Hosts, opts.Users, start_time.Format(TIME_FORMAT))
	fmt.Println("\n------------[PROGRESS]---------------------")

	go generateQuery(queryChan, usernames, hostnames, opts.Domain)

	for i := 0; i < opts.Threads; i++ {
		go qWorker(queryChan, opts, &wg, probeSMPT)
	}

	wg.Wait()
	sort.Strings(gResults)

	fmt.Println("\n---- [Results] ----")
	for _, r := range gResults {
		fmt.Println(r)
	}
	end_time := time.Now()
	fmt.Printf(END_STRING, end_time.Format(TIME_FORMAT), end_time.Sub(start_time).String())

	return nil

}

func validateSMTP(cmd *cobra.Command, args []string) error {
	var options = new(SMTP_Options)
	err := options.AddRequired(cmd,
		"users", &options.Users,
		"hosts", &options.Hosts,
	)
	if err != nil {
		return err
	}

	err = options.Add(cmd,
		"mode", &options.Mode,
		"domain", &options.Domain,
		"from", &options.From,
		"port", &options.Port,
		"threads", &options.Threads,
		"verbose", &options.Verbose,
		"ssl", &options.SSL,
	)
	if err != nil {
		return err
	}
	options.Mode = parseMode(options.Mode)

	cmd.SetContext(context.WithValue(cmd.Context(), Key{}, options))
	return nil
}

func parseMode(m string) string {
	m = strings.ToUpper(m)
	if strings.HasPrefix(m, "E") {
		return EXPN
	}
	if strings.HasPrefix(m, "R") {
		return RCPT
	}
	return VRFY
}

func generateQuery(qChan chan string, users, hosts []string, domain string) {
	defer close(qChan)
	for _, u := range users {
		for _, h := range hosts {
			if domain != "" {
				u = u + "@" + domain
			}
			query := fmt.Sprintf("%s\t%s", h, u)
			qChan <- query
		}
	}
}

func qWorker(qChan chan string, args *SMTP_Options, wg *sync.WaitGroup, qFunc utils.QueryFunction) {
	defer wg.Done()
	for q := range qChan {
		parts := strings.Split(q, "\t")
		if len(parts) != 2 {
			fmt.Printf("[WARNING] Invalid Query Format: %s\n", q)
			continue
		}
		h, u := parts[0], parts[1]
		res, err := qFunc(h, u, args.Mode, args.From, args.Port, args.Time.ToTime(), args.SSL)
		if err != nil {
			fmt.Println(err)
			continue
		}
		gMu.Lock()
		gResults = append(gResults, res)
		gMu.Unlock()
	}
}

func createSMTPConnection(ssl bool, host string, port int, timeout time.Duration) (net.Conn, error) {
	hostPort := fmt.Sprintf("%s:%d", host, port)
	if !ssl {
		return net.DialTimeout("tcp", hostPort, timeout)
	}
	tlsConfig := &tls.Config{ServerName: host}
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", hostPort, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS Connection Error: %v", err)
	}
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	_, err = reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Banner Read Failure: %v", err)
	}
	fmt.Fprintf(writer, "STARTTLS"+RN)
	writer.Flush()

	resp, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Error with TLS: %v", err)
	}

	if !strings.HasPrefix(resp, "220") {
		return nil, fmt.Errorf("STARTTLS Failed: %v", resp)
	}
	tlsCon := tls.Client(conn, tlsConfig)
	if err := tlsCon.Handshake(); err != nil {
		return nil, fmt.Errorf("TLS Handshake Error: %v", err)
	}
	return tlsCon, nil
}

func probeSMPT(host, user, mode, from string, port int, timeout time.Duration, ssl bool) (string, error) {
	conn, err := createSMTPConnection(ssl, host, port, timeout)
	if err != nil {
		return "", fmt.Errorf("Connection Error: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_, err = reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("Banner Read Failure: %v", err)
	}

	fmt.Fprintf(writer, HELO)
	writer.Flush()

	_, err = reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("HELO Response Error: %v", err)
	}

	switch mode {
	case VRFY, EXPN:
		fmt.Fprintf(writer, "%s %s%s", mode, user, RN)
	case RCPT:
		fmt.Fprintf(writer, "%s %s%s", FROM_MAIL, from, RN)
		writer.Flush()

		_, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("RCPT Error: %v", err)
		}
		fmt.Fprintf(writer, "%s %s%s", RCPT_TO, user, RN)
	default:
		return "", fmt.Errorf("%s: %s [Mode seems invalid]", host, user)
	}
	writer.Flush()

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}

	return processResponse(response, host, user, mode)
}

func processResponse(response, host, user, mode string) (string, error) {
	found := fmt.Sprintf("[%s] %s", host, user)
	if strings.HasPrefix(response, "2") {
		return found, nil
	} else if strings.HasPrefix(response, "5") && strings.Contains(response, "authentication") {
		return found, nil
	} else if strings.HasPrefix(response, "5") && strings.Contains(response, "disallowed") {
		fmt.Printf("%s NOT IMPLEMENTED! Exiting...", mode)
		os.Exit(1)
	}
	return "", fmt.Errorf("%v", response)
}
