// Copyright 2012-2022 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

//go:generate go run server/errors_gen.go

import (
	"flag"
	"fmt"
	"strings"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/http_server"
	"github.com/memphisdev/memphis/server"

	"os"

	"go.uber.org/automaxprocs/maxprocs"
)

var usageStr = `
Usage: nats-server [options]

Server Options:
    -a, --addr, --net <host>         Bind to host address (default: 0.0.0.0)
    -p, --port <port>                Use port for clients (default: 4222)
    -n, --name
        --server_name <server_name>  Server name (default: auto)
    -P, --pid <file>                 File to store PID
    -m, --http_port <port>           Use port for http monitoring
    -ms,--https_port <port>          Use port for https monitoring
    -c, --config <file>              Configuration file
    -t                               Test configuration and exit
    -sl,--signal <signal>[=<pid>]    Send signal to nats-server process (ldm, stop, quit, term, reopen, reload)
                                     pid> can be either a PID (e.g. 1) or the path to a PID file (e.g. /var/run/nats-server.pid)
        --client_advertise <string>  Client URL to advertise to other servers
        --ports_file_dir <dir>       Creates a ports file in the specified directory (<executable_name>_<pid>.ports).

Logging Options:
    -l, --log <file>                 File to redirect log output
    -T, --logtime                    Timestamp log entries (default: true)
    -s, --syslog                     Log to syslog or windows event log
    -r, --remote_syslog <addr>       Syslog server addr (udp://localhost:514)
    -D, --debug                      Enable debugging output
    -V, --trace                      Trace the raw protocol
    -VV                              Verbose trace (traces system account as well)
    -DV                              Debug and trace
    -DVV                             Debug and verbose trace (traces system account as well)
        --log_size_limit <limit>     Logfile size limit (default: auto)
        --max_traced_msg_len <len>   Maximum printable length for traced messages (default: unlimited)

JetStream Options:
    -js, --jetstream                 Enable JetStream functionality
    -sd, --store_dir <dir>           Set the storage directory

Authorization Options:
        --user <user>                User required for connections
        --pass <password>            Password required for connections
        --auth <token>               Authorization token required for connections

TLS Options:
        --tls                        Enable TLS, do not verify clients (default: false)
        --tlscert <file>             Server certificate file
        --tlskey <file>              Private key for server certificate
        --tlsverify                  Enable TLS, verify client certificates
        --tlscacert <file>           Client certificate CA for verification

Cluster Options:
        --routes <rurl-1, rurl-2>    Routes to solicit and connect
        --cluster <cluster-url>      Cluster URL for solicited routes
        --cluster_name <string>      Cluster Name, if not set one will be dynamically generated
        --no_advertise <bool>        Do not advertise known cluster information to clients
        --cluster_advertise <string> Cluster URL to advertise to other servers
        --connect_retries <number>   For implicit routes, number of connect retries
        --cluster_listen <url>       Cluster url from which members can solicit routes

Profiling Options:
        --profile <port>             Profiling HTTP port

Common Options:
    -h, --help                       Show this message
    -v, --version                    Show version
        --help_tls                   TLS help
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func runMemphis(s *server.Server) {
	err := analytics.InitializeAnalytics(s.MemphisVersion(), s.GetCustomDeploymentId())
	if err != nil {
		s.Errorf("Failed initializing analytics: " + err.Error())
	}

	isUserPassBased := os.Getenv("USER_PASS_BASED_AUTH") == "true"

	if isUserPassBased {
		// For backward compatibility data from old account to memphis default account
		storeDir := s.Opts().StoreDir
		if storeDir == "" {
			storeDir = os.TempDir()
			storeDir = strings.TrimSuffix(storeDir, "/")
		}

		folderName := fmt.Sprintf("%s%s%s", storeDir, "/jetstream/", server.DEFAULT_GLOBAL_ACCOUNT)
		f, _ := os.Stat(folderName)
		if f != nil {
			err = s.MoveResourcesFromOldToNewDefaultAcc()
			if err != nil {
				s.Errorf("Data from global account to memphis account failed: %s", err.Error())
			}
		}
	}

	err = s.CreateDefaultEntitiesOnMemphisAccount()
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		s.Errorf("Failed create default entities: " + err.Error())
	}

	go http_server.InitializeHttpServer(s)

	var env string
	var message string
	if os.Getenv("DOCKER_ENV") != "" {
		env = "Docker"
		if isUserPassBased {
			message = "\n**********\n\nDashboard/CLI: http://localhost:" + fmt.Sprint(s.Opts().UiPort) + "\nBroker: localhost:" + fmt.Sprint(s.Opts().Port) + " (client connections)\nREST gateway: localhost:" + fmt.Sprint(s.Opts().RestGwPort) + " (Data and management via HTTP)\nUI/CLI/SDK root username - root\nUI/CLI/SDK root password - memphis\n\nDocs: https://docs.memphis.dev/memphis/getting-started/2-hello-world  \n\n**********"
		} else {
			message = "\n**********\n\nDashboard/CLI: http://localhost:" + fmt.Sprint(s.Opts().UiPort) + "\nBroker: localhost:" + fmt.Sprint(s.Opts().Port) + " (client connections)\nREST gateway: localhost:" + fmt.Sprint(s.Opts().RestGwPort) + " (Data and management via HTTP)\nUI/CLI/SDK root username - root\nUI/CLI root password - memphis\nSDK connection token - " + s.Opts().Authorization + "\n\nDocs: https://docs.memphis.dev/memphis/getting-started/2-hello-world  \n\n**********"
		}
		s.Noticef(message)
	} else if os.Getenv("LOCAL_CLUSTER_ENV") != "" {
		env = "Local cluster"
		if isUserPassBased {
			message = "\n**********\n\nDashboard/CLI: http://localhost:9000/9001/9002\nBroker: localhost:6666/6667/6668 (client connections)\nREST gateway: localhost:" + fmt.Sprint(s.Opts().RestGwPort) + " (Data and management via HTTP)\nUI/CLI/SDK root username - root\nUI/CLI/SDK root password - memphis\n\nDocs: https://docs.memphis.dev/memphis/getting-started/2-hello-world  \n\n**********"
		} else {
			message = "\n**********\n\nDashboard/CLI: http://localhost:9000/9001/9002\nBroker: localhost:6666/6667/6668 (client connections)\nREST gateway: localhost:" + fmt.Sprint(s.Opts().RestGwPort) + " (Data and management via HTTP)\nUI/CLI/SDK root username - root\nUI/CLI root password - memphis\nSDK connection token - " + s.Opts().Authorization + "\n\nDocs: https://docs.memphis.dev/memphis/getting-started/2-hello-world  \n\n**********"
		}
		s.Noticef(message)
	} else {
		env = "K8S"
	}

	s.Noticef("*** Memphis broker is ready, ENV: %s :-) ***", env)
}

func main() {
	exe := "nats-server"

	// Create a FlagSet and sets the usage
	fs := flag.NewFlagSet(exe, flag.ExitOnError)
	fs.Usage = usage

	metadataDb, rootUserCreated, err := server.InitializeMetadataStorage()
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("%s: %s", exe, err))
	}

	err = server.InitializeCloudComponents()
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed initializing InitializeCloudComponents: %s: %s", exe, err))
	}

	// create users from env or config file
	var lenUsers int
	var errCreateUsers error
	fmt.Println("rootUserCreated", rootUserCreated)
	// create only on the first system load
	if rootUserCreated {
		lenUsers, errCreateUsers = server.CreateUsersFromConfigOnFirstSystemLoad()
		fmt.Println("check", lenUsers, errCreateUsers)
	}

	// Configure the options from the flags/config file
	opts, err := server.ConfigureOptions(fs, os.Args[1:],
		server.PrintServerAndExit,
		fs.Usage,
		server.PrintTLSHelpAndDie)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("%s: %s", exe, err))
	} else if opts.CheckConfig {
		fmt.Fprintf(os.Stderr, "%s: configuration file %s is valid\n", exe, opts.ConfigFile)
		os.Exit(0)
	}

	// Create the server with appropriate options.
	s, err := server.NewServer(opts)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("%s: %s", exe, err))
	}

	// we do this check here and not below the function creating the users - CreateUsersFromConfigOnFirstSystemLoad because we need the s *Server for logs
	fmt.Println("check2", errCreateUsers)
	if errCreateUsers != nil {
		s.Warnf("[tenant: %v]Failed create users from config file", s.MemphisGlobalAccountString(), errCreateUsers.Error())
	}
	if lenUsers > 0 {
		s.Noticef("[tenant: %v]loaded %d users from config file", s.MemphisGlobalAccountString(), lenUsers)
	}
	// Configure the logger based on the flags
	s.ConfigureLogger()

	// Start things up. Block here until done.
	if err := server.Run(s); err != nil {
		server.PrintAndDie(err.Error())
	}

	// Adjust MAXPROCS if running under linux/cgroups quotas.
	undo, err := maxprocs.Set(maxprocs.Logger(s.Debugf))
	if err != nil {
		s.Warnf("Failed to set GOMAXPROCS: %v", err)
	} else {
		defer undo()
	}
	s.Noticef("Established connection with the meta-data storage")

	runMemphis(s)
	defer db.CloseMetadataDb(metadataDb, s)
	defer analytics.Close()
	s.WaitForShutdown()
}
