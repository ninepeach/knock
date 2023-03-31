package main

import (
	"flag"
	"fmt"
	log "github.com/ninepeach/go-clog"
	"net"
	"net/http"
	"os"
	"os/exec"
)

// Define the default values for the command-line arguments or flags
const (
	defaultIPSet         = "knock"
	defaultAPIRequest    = "ipset"
	defaultTimeout       = "7200"
	defaultListenAddress = ":8080"
)

var (
	ipsetName      string
	ipsetV4Name    string
	ipsetV6Name    string
	apiRequestName string
	timeout        string
	listenAddress  string
)

func init() {
	// Set the default values for the command-line arguments or flags
	flag.StringVar(&ipsetName, "knock", defaultIPSet, "name of the IPSet")
	flag.StringVar(&apiRequestName, "request", defaultAPIRequest, "name of the API request")
	flag.StringVar(&timeout, "timeout", defaultTimeout, "timeout for IP addresses in seconds")
	flag.StringVar(&listenAddress, "address", defaultListenAddress, "listen address for the HTTP server")
}

func main() {
	// Parse the command-line arguments or flags
	flag.Parse()

	ipsetV4Name = ipsetName + "v4"
	ipsetV6Name = ipsetName + "v6"

	// Define routes
	http.HandleFunc("/", defaultHandler)

	// Define the HTTP handler function for the API request
	http.HandleFunc("/"+apiRequestName, ipsetHandler)

	// Set custom 404 handler
	http.HandleFunc("/404", notFoundHandler)
	http.HandleFunc("/favicon.ico", http.NotFound)

	// Start the HTTP server on the specified listen address
	log.Info("Listen on %s", listenAddress)
	log.Info("apiname is %s", apiRequestName)
	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
		os.Exit(1)
	}

}

func ipsetAdd(ip string) error {

	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		log.Error("Invalid IP address %s", ip)
		return fmt.Errorf("Invalid IP address %s", ip)
	}

	tbName := ipsetV4Name
	if parsedIP.To4() == nil {
		tbName = ipsetV6Name
	} else {
		tbName = ipsetV4Name
	}

	// Add the IP address to the IPSet with the specified timeout
	cmd := exec.Command("ipset", "add", tbName, ip, "timeout", timeout)
	err := cmd.Run()
	if err != nil {
		log.Error("Error adding IP address %s to IPSet %s", ip, tbName)
		return err
	}
	return nil
}

func ipsetHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the IP address from the remote connection
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Error parsing remote IP address", http.StatusInternalServerError)
		return
	}

	err = ipsetAdd(remoteIP)
	if err != nil {
		fmt.Fprintf(w, "Failed")
		return
	}

	fmt.Fprintf(w, "OK")
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, ".")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, ".")
}
