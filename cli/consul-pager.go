package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/consul-api"
	"github.com/codegangsta/cli"
	cp "github.com/jmcarbo/consul-pager"
	"os"
	"os/signal"
	"time"
)

func addCheck(name, interval, script string) {
	client := cp.Connect()
	client.Agent().CheckRegister(&consulapi.AgentCheckRegistration{name, name, "",
		consulapi.AgentServiceCheck{Interval: interval, Script: script}})
}

func main() {
	app := cli.NewApp()
	app.Name = "consul-pager"
	app.Usage = "consul alarms on check failures!"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:      "version",
			ShortName: "v",
			Usage:     "consul-externalservice version",
			Action: func(c *cli.Context) {
				fmt.Println(app.Version)
			},
		},
		{
			Name:      "start",
			ShortName: "s",
			Usage:     "start alarm watcher",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: "",
					Usage: "config file",
				},
			},
			Action: func(c *cli.Context) {
				if c.String("config") == "" {
					fmt.Print("Must supply config file")
				}
				client := cp.Connect()
				watcher := cp.LoadPagerFromYAML(c.String("config"), client)
				if watcher != nil {
					log.Printf("Starting pager watcher ...\n")
					stopCh := make(chan struct{})
					go func() {
					TRY_LEADERSHIP:
						watcher.Run()
						if watcher.IsLeader() {
							log.Info("I am the leader now ...")
						}
					WAIT_FOR_EVENT:
						select {
						case <-stopCh:
							watcher.Destroy()
							return
						case <-time.After(10 * time.Second):
							if watcher.IsLeader() {
								log.Info("I am still the leader ...")
								goto WAIT_FOR_EVENT
							} else {
								log.Info("Trying to be leader ...")
								goto TRY_LEADERSHIP
							}
						}
					}()

					// Wait for termination
					signalCh := make(chan os.Signal, 1)
					signal.Notify(signalCh, os.Interrupt, os.Kill)
					select {
					case <-signalCh:
						log.Warn("Received signal, stopping pager watch ...")
						close(stopCh)
					}
				} else {
					log.Error("Error starting pager watcher. Check consul agent is running on localhost:8500. Exiting ...")
				}
			},
		},
		{
			Name:      "addcheck",
			ShortName: "a",
			Usage:     "addcheck name interval script",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Value: "acheck",
					Usage: "check name",
				},
				cli.StringFlag{
					Name:  "interval",
					Value: "10s",
					Usage: "check interval (ex: 10s)",
				},
				cli.StringFlag{
					Name:  "script",
					Value: "",
					Usage: "check script",
				},
			},
			Action: func(c *cli.Context) {
				if c.String("name") == "" || c.String("interval") == "" || c.String("script") == "" {
					fmt.Printf("Needs name, interval and script\n")
				} else {
					addCheck(c.String("name"), c.String("interval"), c.String("script"))
				}
			},
		},
	}
	app.Run(os.Args)
}
