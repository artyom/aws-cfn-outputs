// Command aws-cfn-outputs prints AWS CloudFormation [stack outputs],
// optionally substituting values in a text template.
//
// [stack outputs]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	args := runArgs{}
	flag.StringVar(&args.stackName, "s", args.stackName, "CloudFormation `stack` name or ARN")
	flag.StringVar(&args.tplFile, "t", args.tplFile, "(optional) template `file`;\nfor syntax see https://pkg.go.dev/text/template")
	flag.StringVar(&args.outFile, "o", args.outFile, "(optional) output `file`")
	flag.BoolVar(&args.json, "j", args.json, "output result as JSON (only works if -t is not set)")
	flag.Parse()
	if err := run(ctx, args); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

type runArgs struct {
	stackName string
	tplFile   string
	outFile   string
	json      bool
}

func run(ctx context.Context, args runArgs) error {
	if args.stackName == "" {
		return errors.New("stack name must be set")
	}
	var tpl *template.Template
	if args.tplFile != "" {
		if args.json {
			return errors.New("-j and -t cannot be set at the same time")
		}
		var err error
		if tpl, err = template.ParseFiles(args.tplFile); err != nil {
			return err
		}
		tpl = tpl.Option("missingkey=error")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	out, err := cloudformation.NewFromConfig(cfg).DescribeStacks(ctx,
		&cloudformation.DescribeStacksInput{StackName: &args.stackName})
	if err != nil {
		return err
	}
	if l := len(out.Stacks); l != 1 {
		return fmt.Errorf("got results for %d stacks, want 1", l)
	}
	m := make(map[string]string)
	for _, o := range out.Stacks[0].Outputs {
		if o.OutputKey == nil || o.OutputValue == nil {
			continue
		}
		m[*o.OutputKey] = *o.OutputValue
	}
	buf := new(bytes.Buffer)
	switch {
	case args.json:
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(m); err != nil {
			return err
		}
	case tpl != nil:
		if err := tpl.Execute(buf, m); err != nil {
			return err
		}
	default:
		for _, k := range keys(m) {
			fmt.Fprintf(buf, "%s\t%s\n", k, m[k])
		}
	}
	if args.outFile != "" {
		return os.WriteFile(args.outFile, buf.Bytes(), 0666)
	}
	_, err = os.Stdout.Write(buf.Bytes())
	return err
}

func keys(m map[string]string) []string {
	out := make([]string, len(m))
	var i int
	for k := range m {
		out[i] = k
		i++
	}
	sort.Strings(out)
	return out
}
