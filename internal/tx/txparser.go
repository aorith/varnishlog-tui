package tx

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aorith/varnishlog-tui/internal/ui/state"
	"github.com/aorith/varnishlog-tui/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type NewTxMsg Tx

type FetchEndMsg struct {
	Err error
}

func ExecVarnishlogAndFetchTxs(script state.NewVarnishlogScriptMsg, cancelChan chan struct{}, txChan chan Tx) tea.Cmd {
	tmpCmdScript, err := os.CreateTemp("", "varnishlog-tui-command-*.sh")
	if err != nil {
		return func() tea.Msg {
			return FetchEndMsg{Err: fmt.Errorf("Error creating temporary command file: %s", err.Error())}
		}
	}
	tmpCmdScriptName := tmpCmdScript.Name()
	go func() {
		time.Sleep(time.Second * 1) // Give some time until the script is executed
		if err := os.Remove(tmpCmdScriptName); err != nil {
			log.Debug(fmt.Sprintf("Error removing temporary file: %s", err.Error()))
		}
	}()

	return func() tea.Msg {
		defer close(txChan)

		cmdString := fmt.Sprintf("exec %s", script)
		if _, err := tmpCmdScript.Write([]byte(cmdString)); err != nil {
			return FetchEndMsg{Err: fmt.Errorf("Error writing command to temporary file: %s", err.Error())}
		}
		if err := tmpCmdScript.Close(); err != nil {
			return FetchEndMsg{Err: fmt.Errorf("Error closing temporary command file: %s", err.Error())}
		}

		log.Debug(fmt.Sprintf("Executing: sh %s", tmpCmdScriptName))
		log.Debug(fmt.Sprintf("Command: %s", cmdString))

		cmd := exec.Command("sh", tmpCmdScriptName)
		out, err := cmd.StdoutPipe()
		if err != nil {
			return FetchEndMsg{Err: fmt.Errorf("Error creating StdoutPipe: %s", err.Error())}
		}
		defer out.Close()

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return FetchEndMsg{Err: fmt.Errorf("Error creating StderrPipe: %s", err.Error())}
		}
		defer stderr.Close()

		if err := cmd.Start(); err != nil {
			return FetchEndMsg{Err: fmt.Errorf("Error starting program: %s", err.Error())}
		}

		scanner := bufio.NewScanner(out)
		errScanner := bufio.NewScanner(stderr)

		// Channel to collect stderr output
		errChan := make(chan string)
		go func() {
			var stderrContent strings.Builder
			for errScanner.Scan() {
				line := errScanner.Text()
				stderrContent.WriteString(line + "\n")
				log.Debug(fmt.Sprintf("stderr: %s", line))
			}
			if err := errScanner.Err(); err != nil {
				log.Debug(fmt.Sprintf("Error reading from stderr: %s", err))
			}
			errChan <- stderrContent.String()
			close(errChan)
		}()

		for scanner.Scan() {
			select {
			case <-cancelChan:
				err := cmd.Process.Kill()
				if err != nil {
					return FetchEndMsg{Err: fmt.Errorf("Could not kill the process: %s", err.Error())}
				}
				return FetchEndMsg{}
			default:
				// Look for the start of a transaction, eg:
				// *   << Session  >> 16812342
				// **  << Request  >> 4
				line := strings.TrimSpace(scanner.Text())
				parts := strings.Fields(line)

				if len(parts) != 5 || parts[0][0] != '*' {
					continue
				}

				found := false
				rawTx := []string{line}

				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					rawTx = append(rawTx, line)
					parts := strings.Fields(line)

					if len(parts) >= 2 && parts[1] == "End" {
						// Tx ended (or an VSL store overflow was encountered)
						newTx := parseTx(rawTx)
						if newTx != nil {
							txChan <- *newTx
						}
						found = true
						break
					}
				}

				if !found {
					return FetchEndMsg{Err: fmt.Errorf("Incomplete tx, stopping")}
				}
			}
		}

		var endMsg = FetchEndMsg{}
		if err := scanner.Err(); err != nil {
			endMsg.Err = fmt.Errorf("Error reading from stdout: %s", err.Error())
		}

		cmd.WaitDelay = time.Second
		if err := cmd.Wait(); err != nil {
			stderrOutput := <-errChan
			endMsg.Err = fmt.Errorf("Error: %s %s", err.Error(), stderrOutput)
		}

		return endMsg
	}
}

func ListenForTxsCmd(txChan chan Tx) tea.Cmd {
	return func() tea.Msg {
		for itm := range txChan {
			return NewTxMsg(itm)
		}
		return nil
	}
}

func parseTx(rawTx []string) *Tx {
	currentTx := Tx{
		RawTx:    rawTx,
		Children: make(map[string]*Tx),
	}

	var transition VCLTransition

	for _, s := range rawTx {
		parts := strings.Fields(s)
		partsLen := len(parts)

		// New tx
		// *   << Session  >> 16812342
		// **  << Request  >> 4
		if partsLen >= 5 && parts[0][0] == '*' {
			vxid, err := strconv.ParseUint(parts[4], 10, 64)
			if err != nil {
				return nil
			}
			currentTx.Vxid = vxid
			continue
		}

		// RecordType & reason - the parent cannot be set here because we don't know if it was an ESI
		// -   Begin          bereq 21334348 fetch
		// --  Begin          req 2 esi 1
		// -4- Begin          bereq 6 fetch
		if partsLen > 4 && parts[1] == "Begin" {
			if parts[4] == "esi" && len(parts) > 5 {
				currentTx.Txid = fmt.Sprintf("%d_%s", currentTx.Vxid, parts[5])
			} else {
				currentTx.Txid = fmt.Sprintf("%d", currentTx.Vxid)
			}

			currentTx.RecordType = parts[2]
			currentTx.Reason = parts[4]
			continue
		}

		// Children
		// --  Link           bereq 12 fetch
		// --  Link           req 13 esi 2
		if partsLen > 4 && parts[1] == "Link" {
			var childId string
			if parts[4] == "esi" && len(parts) > 5 {
				childId = fmt.Sprintf("%s_%s", parts[3], parts[5])
			} else {
				childId = parts[3]
			}
			// Add the children as an empty tx for now
			// relationships will be updated later in logview.addNewTx
			currentTx.Children[childId] = &Tx{Txid: childId}
			continue
		}

		// URL (want the first value in the VSL so check if we don't have a URL yet)
		// -   ReqURL         /path/to/resource
		if currentTx.Url == "" && partsLen == 3 && parts[1] == "ReqURL" {
			currentTx.Url = parts[2]
			continue
		}

		// BereqURL (want the last value since it's what the backend receives)
		// -   BereqURL         /path/to/resource
		if partsLen == 3 && parts[1] == "BereqURL" {
			currentTx.Url = parts[2]
			continue
		}

		// Host (want the first value)
		// -   ReqHeader      host: www.example.com
		// -   ReqHeader      Host:www.example.com
		if currentTx.Host == "" && partsLen >= 3 && parts[1] == "ReqHeader" && strings.HasPrefix(strings.ToLower(parts[2]), "host:") {
			if partsLen == 4 {
				currentTx.Host = parts[3]
				continue
			} else {
				hostParts := strings.Split(parts[2], ":")
				if len(hostParts) >= 2 {
					currentTx.Host = hostParts[1]
					continue
				}
			}
		}

		// BereqHost (want the last value)
		// -   BereqHeader      host: www.example.com
		// -   BereqHeader      Host:www.example.com
		if partsLen >= 3 && parts[1] == "BereqHeader" && strings.HasPrefix(strings.ToLower(parts[2]), "host:") {
			if partsLen == 4 {
				currentTx.Host = parts[3]
				continue
			} else {
				hostParts := strings.Split(parts[2], ":")
				if len(hostParts) >= 2 {
					currentTx.Host = hostParts[1]
					continue
				}
			}
		}

		// Method
		// -   ReqMethod      GET
		if partsLen == 3 && (parts[1] == "ReqMethod" || parts[1] == "BereqMethod") {
			currentTx.Method = parts[2]
			continue
		}

		// Status (last one)
		// --  RespStatus     301
		// --- BerespStatus   200
		if partsLen == 3 && (parts[1] == "RespStatus" || parts[1] == "BerespStatus") {
			statusCode, err := strconv.Atoi(parts[2])
			if err == nil {
				currentTx.StatusCode = statusCode
			}
			continue
		}

		// StatusReason (last one)
		// --  RespReason     OK
		// --- BerespReason   OK
		if partsLen == 3 && (parts[1] == "RespReason" || parts[1] == "BerespReason") {
			currentTx.StatusReason = strings.Join(parts[2:], " ")
			continue
		}

		// Timestamp
		//                    label absolute        sinceStart sinceLast
		// -   Timestamp      Resp: 1714823222.274262 0.003330 0.000015
		if partsLen > 4 && parts[1] == "Timestamp" {
			ts := newTimestamp(
				parts[2][:len(parts[2])-1], // Remove ':' from the label
				parts[3],
				parts[4],
				parts[5],
			)
			if ts != nil {
				currentTx.Timestamps = append(currentTx.Timestamps, *ts)
			}
			continue
		}

		// TTL
		// --  TTL            RFC 120 10 0 1606398419 1606398419 1606398419 0 0 cacheable
		// --  TTL            VCL 120 10 0 1606400537 uncacheable
		// --  TTL            HFP 10 0 0 1606402666 uncacheable
		if parts[1] == "TTL" && (partsLen == 12 || partsLen == 8) {
			ttl := newTTL(parts[2:])
			if ttl != nil {
				currentTx.TTL = append(currentTx.TTL, *ttl)
			}
			continue
		}

		// Transitions
		// --  VCL_call       RECV
		if partsLen == 3 && parts[1] == "VCL_call" {
			transition = VCLTransition{Call: parts[2]}
			continue
		}
		// --  VCL_return     synth
		if partsLen == 3 && parts[1] == "VCL_return" {
			transition.Return = parts[2]
			currentTx.Transitions = append(currentTx.Transitions, transition)
			continue
		}

		// Accounting information
		// h=header, b=body, t=total
		//                    received-|-transmitted
		//                    h   b t   h   b t
		// --  ReqAcct        611 0 611 287 0 287
		// --- BereqAcct      619 0 619 536 935300 935836
		if partsLen == 8 && (parts[1] == "ReqAcct" || parts[1] == "BereqAcct") {
			headerBytesReceived, err := strconv.Atoi(parts[2])
			if err != nil {
				continue
			}
			bodyBytesReceived, err := strconv.Atoi(parts[3])
			if err != nil {
				continue
			}
			headerBytesTransmitted, err := strconv.Atoi(parts[5])
			if err != nil {
				continue
			}
			bodyBytesTransmitted, err := strconv.Atoi(parts[6])
			if err != nil {
				continue
			}
			acct := RequestAccounting{
				HeaderBytesReceived:    util.SizeValue(headerBytesReceived),
				BodyBytesReceived:      util.SizeValue(bodyBytesReceived),
				HeaderBytesTransmitted: util.SizeValue(headerBytesTransmitted),
				BodyBytesTransmitted:   util.SizeValue(bodyBytesTransmitted),
			}
			currentTx.Accounting = acct

			continue
		}

		// Special handling for sessions
		if currentTx.RecordType == "sess" {
			// Timestamp
			// -   SessClose      REM_CLOSE 0.000
			if partsLen == 4 && parts[1] == "SessClose" {
				ts := newTimestamp(
					parts[2],
					"0",
					parts[3],
					parts[3],
				)
				if ts != nil {
					currentTx.Timestamps = append(currentTx.Timestamps, *ts)
				}
			}

			if currentTx.Method == "" {
				currentTx.Method = "-"
			}
			if currentTx.Host == "" {
				currentTx.Host = "-"
			}
			if currentTx.Url == "" {
				currentTx.Url = "-"
			}
			// Looking for:
			// -   VSL            store overflow
			// -   SessOpen       1.2.3.4 5673 a0 2.3.4.5 80 1714899853.663185 3704
			if len(parts) >= 4 {
				if parts[1] == "VSL" && parts[3] == "overflow" {
					currentTx.Host = "store overflow"
					continue
				}
				if parts[1] == "SessOpen" {
					currentTx.Url = parts[1] + "  " + strings.Join(parts[2:], " ")
					continue
				}
			}
		}
	}

	return &currentTx
}

func newTimestamp(label, absStr, sinceStartStr, sinceLastStr string) *Timestamp {
	absoluteTime, err := util.ConvertUnixTimestamp(absStr)
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid Timestamp: unparsable absoluteTime field: %s\n%s", err.Error(), absStr))
		return nil
	}

	// Convert sinceStart and sinceLast from string to float64
	sinceStartFloat, err := strconv.ParseFloat(sinceStartStr, 64)
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid Timestamp: unparsable sinceStart field: %s\n%s", err.Error(), sinceStartStr))
		return nil
	}
	sinceLastFloat, err := strconv.ParseFloat(sinceLastStr, 64)
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid Timestamp: unparsable sinceLast field: %s\n%s", err.Error(), sinceLastStr))
		return nil
	}

	// Convert the float64 values to time.Duration
	sinceStart := time.Duration(sinceStartFloat * float64(time.Second))
	sinceLast := time.Duration(sinceLastFloat * float64(time.Second))

	return &Timestamp{
		EventLabel: label,
		Absolute:   absoluteTime,
		SinceStart: sinceStart,
		SinceLast:  sinceLast,
	}
}

func newTTL(parts []string) *TTLData {
	// RFC 120 10 0 1606398419 1606398419 1606398419 0 0 cacheable
	// VCL 120 10 0 1606400537 uncacheable
	// HFP 10 0 0 1606402666 uncacheable

	ttlData := TTLData{Source: parts[0]}

	// First 5 parts are common
	ttl, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable ttl field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.TTL = ttl

	grace, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable grace field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.Grace = grace

	keep, err := strconv.Atoi(parts[3])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable keep field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.Keep = keep

	ref, err := util.ConvertUnixTimestamp(parts[4])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable reference field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.Reference = ref

	// Check if we are parsing a VCL or HFP source (6 fields) or a HFP
	if len(parts) == 6 {
		ttlData.CacheStatus = parts[5]
		return &ttlData
	}

	if len(parts) != 10 {
		log.Debug(fmt.Sprintf("Invalid TTL: len does not match: %s", parts))
		return nil
	}

	age, err := strconv.Atoi(parts[5])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable age field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.Age = age

	date, err := util.ConvertUnixTimestamp(parts[6])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable date field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.Date = date

	expires, err := util.ConvertUnixTimestamp(parts[7])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable expires field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.Expires = expires

	maxAge, err := strconv.Atoi(parts[8])
	if err != nil {
		log.Debug(fmt.Sprintf("Invalid TTL: unparsable maxAge field: %s\n%s", err.Error(), parts))
		return nil
	}
	ttlData.MaxAge = maxAge

	// Last field
	ttlData.CacheStatus = parts[9]

	return &ttlData
}
