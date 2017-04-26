package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"
)

// реализация сортировки hosts, необходимо чтобы вниз сваливались пофейленные
type sortHostStates struct {
	slice []*hostState
}

func (h sortHostStates) Len() int {
	return len(h.slice)
}
func (h sortHostStates) Swap(i, j int) {
	h.slice[i], h.slice[j] = h.slice[j], h.slice[i]
}

func (h sortHostStates) Less(i, j int) bool {
	hostI := h.slice[i]
	hostJ := h.slice[j]
	if hostI.err != nil && hostJ.err == nil {
		return false
	}
	if hostI.err != nil && hostJ.err != nil {
		return hostI.err.Error() < hostJ.err.Error()
	}
	return true
}

// отчет по запуску
func report(hosts []*hostState) {

	time.Sleep(1 * time.Second)

	fmt.Fprintf(os.Stdout, "--- Report --------------------------------\n")

	sortedHostStates := &sortHostStates{slice: hosts}
	sort.Sort(sortedHostStates)

	// подсчет
	var count, succ, notStarted, connFailed, execFailed int
	badHosts := ""
	for _, h := range sortedHostStates.slice {
		count++
		if h.startedAt == nil {
			notStarted++
			fmt.Fprintf(os.Stderr, "%s\t\t%v\n", h.visibleHostName, "< not started >")
			badHosts += h.connectionAddress + "\n"
			continue
		}
		if h.connectedAt == nil {
			connFailed++
			fmt.Fprintf(os.Stderr, "%s\t\t%v\t%v\n", h.visibleHostName, h.err, "< not connected >")
			badHosts += h.connectionAddress + "\n"
			continue
		}
		if h.endedAt == nil {
			execFailed++
			fmt.Fprintf(os.Stderr, "%s\t\t%v,\t< time %v >\n", h.visibleHostName, h.err, h.connectedAt.Sub(*h.startedAt))
			badHosts += h.connectionAddress + "\n"
			continue
		}
		if h.err == nil {
			fmt.Printf("%s\t\t%v,\t< time %v >\n", h.visibleHostName, "< successfully completed >", h.endedAt.Sub(*h.startedAt))
			succ++
		} else {
			fmt.Fprintf(os.Stderr, "%s\t\t%v,\t< time: %v >\n", h.visibleHostName, h.err, h.endedAt.Sub(*h.startedAt))
			badHosts += h.connectionAddress + "\n"
		}
	}

	// подвал
	fmt.Fprintf(os.Stdout, "--------------------------------\n")
	if notStarted == 0 && connFailed == 0 && execFailed == 0 {
		fmt.Fprintf(os.Stdout, "Total: %v Success: %v\n", count, succ)
	} else {
		fmt.Fprintf(os.Stderr, "Total: %d Success: %d Connect failed: %d Execute failed: %d\n", count, succ, connFailed+notStarted, execFailed)
		if file, err := ioutil.TempFile(os.TempDir(), "knife-sh"); err == nil {
			file.WriteString(badHosts)
			fmt.Fprintf(os.Stderr, "File with bad hosts: %s\n", file.Name())
			file.Sync()
			file.Close()
		}
		os.Exit(count - succ)
	}
}
