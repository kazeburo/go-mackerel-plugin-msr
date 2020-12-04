package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
)

// Version by Makefile
var version string

type opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
	Version bool          `short:"v" long:"version" description:"Show version"`
}

type slave struct {
	ChannelName *string `mysqlvar:"Channel_Name"`
	Behind      int64   `mysqlvar:"Seconds_Behind_Master"`
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := opts{}
	psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opts.Version {
		fmt.Fprintf(os.Stderr, "Version: %s\nCompiler: %s %s\n",
			version,
			runtime.Compiler,
			runtime.Version())
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	now := time.Now().Unix()

	db, err := mysqlflags.OpenDB(opts.MyOpts, opts.Timeout, false)
	if err != nil {
		log.Printf("couldn't connect DB: %v", err)
		return 1
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	ch := make(chan error, 1)

	var slaves []slave
	go func() {
		ch <- mysqlflags.Query(db, "SHOW SLAVE STATUS").Scan(&slaves)
	}()

	select {
	case err = <-ch:
		// nothing
	case <-ctx.Done():
		err = fmt.Errorf("Connection or query timeout")
	}

	if err != nil {
		log.Printf("%v", err)
		return 1
	}

	if len(slaves) == 0 {
		log.Printf("No replication settings")
		return 1
	}

	for _, slave := range slaves {
		if slave.ChannelName == nil {
			*slave.ChannelName = "-"
		}
		fmt.Printf("mysql-msr.behind.%s\t%d\t%d\n", *slave.ChannelName, slave.Behind, now)
	}

	return 0
}
