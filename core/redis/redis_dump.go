package redis

import (
	"fmt"
	"github.com/yannh/redis-dump-go/pkg/config"
	"github.com/yannh/redis-dump-go/pkg/redisdump"
	"log"
	"os"
)

func RedisDump(host string, port int, user, passwd, file string) error {
	s := redisdump.Host{
		Host:     host,
		Port:     port,
		Username: user,
		Password: passwd,
		//TlsHandler: tlshandler,
	}
	db := redisdump.AllDBs
	c := config.Config{
		Filter:    "*",
		NWorkers:  10,
		WithTTL:   true,
		BatchSize: 1000,
		Noscan:    true,
	}
	f, _ := os.Create(file)
	defer f.Close()
	serializer := redisdump.RESPSerializer
	progressNotifs := make(chan redisdump.ProgressNotification)

	defer close(progressNotifs)
	go func() {
		for _ = range progressNotifs {
			// log something
		}
	}()

	logger := log.New(f, "", 0)

	if err := redisdump.DumpServer(s, db, c.Filter, c.NWorkers, c.WithTTL, c.BatchSize, c.Noscan, logger, serializer, progressNotifs); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return err
	}
	return nil
}
