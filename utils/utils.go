package utils

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

const (
	RN          = "\r\n"
	TIME_FORMAT = "Mon, 2006-01-02 15:04:05"
)

type QueryFunction func(string, string, string, string, int, time.Duration, bool) (string, error)

type Duration time.Duration

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func (d *Duration) ToTime() time.Duration {
	return time.Duration(*d)
}

func (d *Duration) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

func (d *Duration) Type() string {
	return "Duration"
}

func AppendFileContentsOrString(name string, arr *[]string) {
	file, err := os.Open(name)
	if err != nil {
		*arr = append(*arr, strings.ReplaceAll(name, " ", ""))
		//fmt.Printf("String: %s\n", name)
		return

	}
	defer file.Close()
	//fmt.Printf("Found File: %s\n", name)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		text = strings.ReplaceAll(text, " ", "")
		*arr = append(*arr, text)
	}
	if err := scanner.Err(); err != nil {
		fmt.Errorf("%v\n", err)
	}

}

type Flaggable interface {
	AddRequired(cmd *cobra.Command, flags ...interface{}) error
	Add(cmd *cobra.Command, flags ...interface{}) error
}

type Options struct{}

func (opt *Options) AddRequired(cmd *cobra.Command, flags ...interface{}) error {
	parseError := "%s parameter was parsed incorrectly: %v"
	requiredError := "%s parameter is necessary"
	if len(flags)%2 != 0 {
		return fmt.Errorf("Invalid Number of arguments")
	}
	for i := 0; i < len(flags); i += 2 {
		name, ok := flags[i].(string)
		if !ok {
			return fmt.Errorf("Invalid argument name: %s -- %v", name, ok)
		}
		value := flags[i+1]
		switch v := value.(type) {
		case *string:
			field, err := cmd.Flags().GetString(name)
			if err != nil {
				return fmt.Errorf(parseError, name, err)
			}
			if field == "" {
				return fmt.Errorf(requiredError, name)
			}
			*v = field

		case *int:
			field, err := cmd.Flags().GetInt(name)
			if err != nil {
				return fmt.Errorf(parseError, err)
			}
			*v = field
		case *bool:
			field, err := cmd.Flags().GetBool(name)
			if err != nil {
				return fmt.Errorf(parseError, name, err)
			}
			*v = field
		default:
			fmt.Println(v)
			return fmt.Errorf("Unsupported flag type: %T: %T", value, v)
		}
	}
	return nil

}
func (opt *Options) Add(cmd *cobra.Command, flags ...interface{}) error {
	parseError := "%s parameter was parsed incorrectly: %v"
	if len(flags)%2 != 0 {
		return fmt.Errorf("Invalid argument count")
	}
	for i := 0; i < len(flags); i += 2 {
		name, ok := flags[i].(string)
		if !ok {
			return fmt.Errorf("Invalid parameter name! Got %s", name)
		}

		value := flags[i+1]
		switch v := value.(type) {
		case *string:
			field, err := cmd.Flags().GetString(name)
			if err != nil {
				return fmt.Errorf(parseError, name, err)
			}
			*v = field

		case *int:
			field, err := cmd.Flags().GetInt(name)
			if err != nil {
				return fmt.Errorf(parseError, err)
			}
			*v = field
		case *bool:
			field, err := cmd.Flags().GetBool(name)
			if err != nil {
				return fmt.Errorf(parseError, name, err)
			}
			*v = field
		default:
			fmt.Println(v)
			return fmt.Errorf("Unsupported flag type: %T: %T", value, v)
		}

	}
	return nil

}
