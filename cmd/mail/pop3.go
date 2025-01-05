/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package mail

import (
	"fmt"
	"sync"
	"time"

	"github.com/phaze228/genum/utils"
	"github.com/spf13/cobra"
)

// Pop3Cmd represents the pop3 command
var Pop3Cmd = &cobra.Command{
	Use:   "pop3",
	Short: "Enumerates POP3 User Emails",
	Long: `
	[POP3 EMAIL ENUMERATION/BRUTE FORCER]
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Unimplemented")
	},
}

const (
	//Defaults
	POP3_PORT = 110
	POP3_SSL  = 995

	//POP3 COMMANDS
	USER = "USER %s"
	PASS = "PASS %s"
)

type POP3_Options struct {
	utils.Options
	Users     string
	Passwords string
	Hosts     string
	Mode      string
	Domain    string
	From      string
	Port      int
	Threads   int
	Time      utils.Duration
	Verbose   bool
	SSL       bool
}

func init() {
	var duration utils.Duration = utils.Duration(time.Duration(3) * time.Second)
	Pop3Cmd.Flags().StringP("users", "U", "", "username or file with list of usernames")
	Pop3Cmd.Flags().StringP("passwords", "P", "", "hosts or file with list of hosts")
	Pop3Cmd.Flags().StringP("hosts", "H", "", "hosts or file with list of hosts")
	Pop3Cmd.Flags().StringP("domain", "D", "", "Domain to append to usernames: user@domain.com")
	Pop3Cmd.Flags().StringP("from", "F", DEFAULT_EMAIL, "For use with RCPT, address of FROM sender: user@meow.com")
	Pop3Cmd.Flags().IntP("port", "p", POP3_PORT, "Port the SMTP Service runs on")
	Pop3Cmd.Flags().IntP("threads", "T", THREAD_COUNT, "Thread Count: Default: 10")
	Pop3Cmd.Flags().VarP(&duration, "duration", "d", "Timeout: 3s, 10s...etc")
	Pop3Cmd.Flags().BoolP("verbose", "v", false, "Verbose output, prints nearly everything")
	Pop3Cmd.Flags().BoolP("ssl", "S", false, "Enable SSL")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pop3Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pop3Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func executePOP3(cmd *cobra.Command, args []string) error {
	validatedArgs := cmd.Context().Value(Key{})
	if validatedArgs == nil {
		return fmt.Errorf("[Command Line Options Error]")
	}
	opts, ok := validatedArgs.(*POP3_Options)
	if !ok {
		return fmt.Errorf("Invalid Type: %T", validatedArgs)
	}

	usernames := make([]string, 0)
	hostnames := make([]string, 0)
	passwords := make([]string, 0)

	utils.AppendFileContentsOrString(opts.Users, &usernames)
	utils.AppendFileContentsOrString(opts.Hosts, &hostnames)
	utils.AppendFileContentsOrString(opts.Passwords, &passwords)

	var wg sync.WaitGroup
	wg.Add(opts.Threads)

	// queryChan := make(chan string)
	// startTime := time.Now()

	return nil
}

func generatePOP3Query(qChan chan string, users, hosts []string, domain string) {
	defer close(qChan)
	for _, u := range users {
		for _, h := range hosts {
			if domain != "" {
				u = u + "@" + domain
			}
			q := fmt.Sprintf("%s\t%s", h, u)
			qChan <- q
		}
	}
}
