package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "netdebug",
	Short: "netdebug tool",
}

func init() {
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run the netdebug server",
		Run: func(cmd *cobra.Command, args []string) {
			ln, err := net.Listen("tcp", ":8080")
			if err != nil {
				log.Fatal(err)
			}
			defer ln.Close()
			log.Println("Server started and accepting connections")
			for {
				// Accept a connection
				conn, err := ln.Accept()
				if err != nil {
					log.Println(err)
				}
				log.Println("Accepted a new connection")

				go func(c net.Conn) {
					// Close the connection when done with it
					defer c.Close()

					bufReader := bufio.NewReader(c)

					for {
						// Set a 30 second timeout
						c.SetDeadline(time.Now().Add(30 * time.Second))

						// Read from the connection
						buf, err := bufReader.ReadBytes('\n')
						if err != nil {
							log.Println(err)
							break
						}

						// Send a response back to person contacting us.
						_, err = c.Write([]byte(fmt.Sprintf("You said: %s", buf)))
						if err != nil {
							log.Println(err)
							break
						}
					}
				}(conn)
			}
		},
	}

	rootCmd.AddCommand(serverCmd)

	var clientHost string
	var clientPort string
	var period string
	var clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Run the netdebug client",
		Run: func(cmd *cobra.Command, args []string) {
			address := net.JoinHostPort(clientHost, clientPort)
			duration, err := time.ParseDuration(period)
			if err != nil {
				log.Fatal(err)
			}

			for {
				conn, err := net.Dial("tcp", address)
				if err != nil {
					log.Printf("Error connecting to server: %s", err)
					time.Sleep(1 * time.Second)
					continue
				}
				log.Printf("Connected to server at: %s", address)

				func(c net.Conn) {
					defer c.Close()
					bufReader := bufio.NewReader(c)
					ticker := time.Tick(duration)
					for ; true; <-ticker {
						// send message
						_, err = c.Write([]byte("Soy todo oídos!\n"))
						if err != nil {
							log.Println(err)
							break
						}

						// read response
						buf, err := bufReader.ReadString('\n')
						if err != nil {
							log.Println(err)
							break
						}

						log.Printf("Response from server: \"%s\"", strings.TrimSuffix(buf, "\n"))
					}
				}(conn)
			}
		},
	}

	clientCmd.Flags().StringVarP(&clientHost, "host", "H", "netdebug", "Host to send messages to")
	clientCmd.Flags().StringVarP(&clientPort, "port", "p", "8080", "Port to send messages to")
	clientCmd.Flags().StringVarP(&period, "period", "t", "5s", "Duration with which to send messages")
	rootCmd.AddCommand(clientCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
