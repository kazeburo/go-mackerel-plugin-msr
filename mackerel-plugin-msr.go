package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
)

// Version by Makefile
var version string

type Opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
	Version bool          `short:"v" long:"version" description:"Show version"`
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := Opts{}
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

	behinds := map[string]int64{}
	go func() {
		rows, e := db.Query("SHOW SLAVE STATUS")
		if e != nil {
			ch <- e
			return
		}
		defer rows.Close()

		cols, e := rows.Columns()
		if e != nil {
			ch <- e
			return
		}
		vals := make([]interface{}, len(cols))
		idxChannelName := -1
		idxSecondsBehindMaster := -1
		for i, v := range cols {
			vals[i] = new(sql.RawBytes)
			if v == "Channel_Name" {
				idxChannelName = i
			}
			if v == "Seconds_Behind_Master" {
				idxSecondsBehindMaster = i
			}
		}
		if idxSecondsBehindMaster < 0 {
			ch <- fmt.Errorf("Could not find Seconds_Behind_Master in columns")
			return
		}

		i := 0
		for rows.Next() {
			i++
			e = rows.Scan(vals...)
			if e != nil {
				ch <- e
				return
			}
			channelName := "-"
			if idxChannelName >= 0 {
				channelName = string(*vals[idxChannelName].(*sql.RawBytes))
			}
			strSecondsBehindMaster := string(*vals[idxSecondsBehindMaster].(*sql.RawBytes))
			secondsBehindMaster, e := strconv.ParseInt(strSecondsBehindMaster, 10, 64)
			if e != nil {
				ch <- e
				return
			}
			behinds[channelName] = secondsBehindMaster
		}
		if err := rows.Err(); err != nil {
			ch <- err
			return
		}

		if i == 0 {
			ch <- fmt.Errorf("No replication settings")
			return
		}

		ch <- nil
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

	for n, behind := range behinds {
		fmt.Printf("mysql-msr.behind.%s\t%d\t%d\n", n, behind, now)
	}

	return 0
}
