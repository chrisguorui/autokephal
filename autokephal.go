package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	irc "github.com/fluffle/goirc/client"
)

// Struct for the toml config
type TomlConfig struct {
	Bot        botInfo
	Connection connectionInfo
}
type botInfo struct {
	Nick string
	User string
}

type connectionInfo struct {
	ServerName string
	ServerPort string
	SSL        bool
	Channel    string
}

// constant in order to find the config
const (
	defaultConfigPath = "config.toml"
)

// Simple error checking
func checkErr(e error, msg string) {
	if e != nil {
		log.Fatalln(msg, e)
	}
}

func main() {

	configPath := flag.String("config", defaultConfigPath, "Path to the toml configuration file")
	flag.Parse()

	// Decoding the config
	var config TomlConfig
	_, err := toml.DecodeFile(*configPath, &config)
	checkErr(err, "Could not decode the configuration file!")

	// Creating the IRC client with goirc/client config and filling the values from our config
	ircfg := irc.NewConfig(config.Bot.Nick)
	ircfg.SSL = config.Connection.SSL
	ircfg.SSLConfig = &tls.Config{ServerName: config.Connection.ServerName}
	ircfg.Server = fmt.Sprintf("%s:%s", config.Connection.ServerName, config.Connection.ServerPort)
	ircfg.NewNick = func(n string) string { return n + "^" }
	client := irc.Client(ircfg)

	// Handle connection
	client.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { conn.Join(config.Connection.Channel) })
	// Handle disconnection
	quit := make(chan bool)
	client.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	err = client.Connect()
	checkErr(err, fmt.Sprintf("Cannot etablish a connection to the server: %s:%s", config.Connection.ServerName, config.Connection.ServerPort))

	// Wait for disconnect
	<-quit
}
