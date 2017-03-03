package ssh

import (
	"fmt"
	"os"
	"time"
)

// отчет по запуску
func report(hosts []*hostState) {

	time.Sleep(1 * time.Second)

	fmt.Fprintf(os.Stdout, "--- Report --------------------------------\n")

	// подсчет
	var count, succ, notStarted, connFailed, execFailed int
	for _, h := range hosts {
		count++
		if h.startedAt == nil {
			notStarted++
			fmt.Printf("%s\t\t%v\n", h.hostname, "< not started >")
			continue
		}
		if h.connectedAt == nil {
			connFailed++
			fmt.Printf("%s\t\t%v\t%v\n", h.hostname, h.err, "< not connected >")
			continue
		}
		if h.endedAt == nil {
			fmt.Printf("%s\t\t%v,\t< time %v >\n", h.hostname, h.err, h.connectedAt.Sub(*h.startedAt))
			execFailed++
			continue
		}
		if h.err == nil {
			fmt.Printf("%s\t\t%v,\t< time %v >\n", h.hostname, "< successfully completed >", h.endedAt.Sub(*h.startedAt))
			succ++
		} else {
			fmt.Printf("%s\t\t%v,\t< time: %v >\n", h.hostname, h.err, h.endedAt.Sub(*h.startedAt))
		}
	}

	// подвал
	fmt.Fprintf(os.Stdout, "--------------------------------\n")
	if notStarted == 0 && connFailed == 0 && execFailed == 0 {
		fmt.Fprintf(os.Stdout, "Total: %v Success: %v\n", count, succ)
	} else {
		fmt.Fprintf(os.Stderr, "Total: %d Success: %d Connect failed: %d Execute failed: %d\n", count, succ, connFailed+notStarted, execFailed)
		os.Exit(count - succ)
	}
}
