package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	ftpserver "github.com/oarkflow/ftp-server"
	"github.com/oarkflow/ftp-server/config"
	"github.com/oarkflow/ftp-server/ftp"
	"github.com/oarkflow/ftp-server/log/oarklog"
)

var (
	// BuildVersion is the current version of the program
	BuildVersion = ""

	// BuildDate is the time the program was built
	BuildDate = ""

	// Commit is the git hash of the program
	Commit = ""
)

var (
	ftpServer *ftp.Server
	driver    *ftpserver.Server
)

func main() {
	// Arguments vars
	var confFile string
	var onlyConf bool

	// Parsing arguments
	flag.StringVar(&confFile, "conf", "", "Configuration file")
	flag.BoolVar(&onlyConf, "conf-only", false, "Only create the conf")
	flag.Parse()

	// Setting up the logger
	logger := oarklog.Default()

	logger.Info("FTP server", "version", BuildVersion, "date", BuildDate, "commit", Commit)

	autoCreate := onlyConf

	// The general idea here is that if you start it without any arg, you're probably doing a local quick&dirty run
	// possibly on a windows machine, so we're better of just using a default file name and create the file.
	if confFile == "" {
		confFile = "ftpserver.json"
		autoCreate = true
	}

	if autoCreate {
		if _, err := os.Stat(confFile); err != nil && os.IsNotExist(err) {
			logger.Warn("No conf file, creating one", "confFile", confFile)

			if err := os.WriteFile(confFile, confFileContent(), 0600); err != nil { //nolint: gomnd
				logger.Warn("Couldn't create conf file", "confFile", confFile)
			}
		}
	}

	conf, errConfig := config.NewConfig(confFile, logger)
	if errConfig != nil {
		logger.Error("Can't load conf", "err", errConfig)

		return
	}

	// Loading the driver
	var errNewServer error
	driver, errNewServer = ftpserver.NewServer(conf, logger.With("component", "driver"))

	if errNewServer != nil {
		logger.Error("Could not load the driver", "err", errNewServer)

		return
	}

	// Instantiating the server by passing our driver implementation
	ftpServer = ftp.New(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	ftpServer.Logger = logger.With("component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe
	if onlyConf {
		logger.Warn("Only creating conf")

		return
	}

	if err := ftpServer.ListenAndServe(); err != nil {
		logger.Error("Problem listening", "err", err)
	}

	// We wait at most 1 minutes for all clients to disconnect
	if err := driver.WaitGracefully(time.Minute); err != nil {
		ftpServer.Logger.Warn("Problem stopping server", "err", err)
	}
}

func stop() {
	driver.Stop()

	if err := ftpServer.Stop(); err != nil {
		ftpServer.Logger.Error("Problem stopping server", "err", err)
	}
}

func signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		sig := <-ch

		if sig == syscall.SIGTERM {
			stop()

			break
		}
	}
}

func confFileContent() []byte {
	str := `{
  "version": 1,
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "os",
      "params": {
        "basePath": "/tmp"
      }
    }
  ],
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  }
}`

	return []byte(str)
}
