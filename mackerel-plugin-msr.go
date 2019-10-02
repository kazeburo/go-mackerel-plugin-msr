package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

type connectionOpts struct {
	Host string `short:"H" long:"host" default:"localhost" description:"Hostname"`
	Port string `short:"p" long:"port" default:"3306" description:"Port"`
	User string `short:"u" long:"user" default:"root" description:"Username"`
	Pass string `short:"P" long:"password" default:"" description:"Password"`
}

func main() {
	os.Exit(_main())
}

func _main() (st int) {
	opts := connectionOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		os.Exit(1)
	}

	db := mysql.New("tcp", "", fmt.Sprintf("%s:%s", opts.Host, opts.Port), opts.User, opts.Pass, "")
	err = db.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't connect DB\n")
		return
	}
	defer db.Close()

	rows, _, err := db.Query("SELECT CHANNEL_NAME FROM performance_schema.replication_connection_status")
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't execute query: %s\n", err)
		return
	}

	if len(rows) == 0 {
		fmt.Fprintf(os.Stderr, "no replication channels found\n")
		return
	}

	now := int32(time.Now().Unix())
	for _, row := range rows {
		channelName := row.Str(0)
		query := fmt.Sprintf("SHOW SLAVE STATUS FOR CHANNEL '%s'\n", channelName)
		slaveRows, slaveRes, err := db.Query(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't execute query")
			return
		}

		idxSecondsBehindMaster := slaveRes.Map("Seconds_Behind_Master")
		secondsBehindMaster := slaveRows[0].Int64(idxSecondsBehindMaster)

		fmt.Printf("mysql-msr.behind.%s\t%d\t%d\n", channelName, secondsBehindMaster, now)
	}

	st = 0
	return
}
