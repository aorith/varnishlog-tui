package tx

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aorith/varnishlog-tui/internal/ui/styles"
	"github.com/aorith/varnishlog-tui/internal/util"
	"github.com/charmbracelet/lipgloss"
)

// Tx represents a single transaction
type Tx struct {
	Txid         string
	Vxid         uint64
	RecordType   string // req, bereq, sess
	Reason       string // rxreq, fetch, esi, ...
	Method       string
	Host         string
	Url          string
	StatusCode   int
	StatusReason string
	Timestamps   []Timestamp
	Transitions  []VCLTransition
	TTL          []TTLData
	Accounting   RequestAccounting
	Parent       *Tx
	Children     map[string]*Tx
	RawTx        []string
}

type Timestamp struct {
	EventLabel string        // Start, Req, Fetch, Process, Resp, ...
	Absolute   time.Time     // Absolute time of the timestamp
	SinceStart time.Duration // Duration since the start of the tx
	SinceLast  time.Duration // Duration since the last timestamp
}

type TTLData struct {
	Source      string    // "RFC", "VCL" or "HFP"
	TTL         int       // Time-to-live
	Grace       int       // Grace period
	Keep        int       // Keep period
	Reference   time.Time // Reference time for TTL
	Age         int       // Age (incl Age: header value)
	Date        time.Time // Date header
	Expires     time.Time // Expires header
	MaxAge      int       // Max-Age from Cache-Control header
	CacheStatus string    // "cacheable" or "uncacheable"
}

type VCLTransition struct {
	Call   string // RECV, HASH
	Return string // synth, lookup
}

type RequestAccounting struct {
	HeaderBytesReceived    util.SizeValue
	BodyBytesReceived      util.SizeValue
	HeaderBytesTransmitted util.SizeValue
	BodyBytesTransmitted   util.SizeValue
}

// FilterValue satisfaces list.Item interface
func (t Tx) FilterValue() string {
	return strings.ReplaceAll(t.AsString(nil, false, false), "\n", " ")
}

// AsString
//
//	123 req 122 rxreq (200 OK)
//	GET www.example.com/path/to/asset
func (t Tx) AsString(matchedRunes []int, highlight, highlightMatches bool) string {
	var (
		txid         string = t.Txid
		recordType   string = t.RecordType
		parentId     string = "-"
		reason       string = t.Reason
		host         string = t.Host
		method       string = t.Method
		url          string = t.Url
		statusCode   string = fmt.Sprintf("(%d", t.StatusCode)
		statusReason string = t.StatusReason + ")"
		offset       int    = 0
	)
	if t.Parent != nil {
		parentId = t.Parent.Txid
	}

	if t.RecordType == "sess" {
		// In sessions the Host is either empty or an store overflow
		// method is always "-" and Url is SessionOpen
		statusCode = ""
		statusReason = ""
		if host != "-" && host != "" {
			host = fmt.Sprintf("(%s)", host)
		}
		host = host + " "
	}

	if highlightMatches && highlight {
		txid, offset = styleRunesWithOffset(txid, offset, matchedRunes, styles.TxidStyle)
		recordType, offset = styleRunesWithOffset(recordType, offset, matchedRunes, styles.RecordTypeStyle)
		parentId, offset = styleRunesWithOffset(parentId, offset, matchedRunes, styles.TxidStyle)
		reason, offset = styleRunesWithOffset(reason, offset, matchedRunes, styles.ReasonStyle)
		statusCode, offset = styleRunesWithOffset(statusCode, offset, matchedRunes, styles.ReasonStyle)
		statusReason, offset = styleRunesWithOffset(statusReason, offset, matchedRunes, styles.ReasonStyle)
		method, offset = styleRunesWithOffset(method, offset, matchedRunes, styles.MethodStyle)
		host, offset = styleRunesWithOffset(host, offset, matchedRunes, styles.HostStyle)
		if t.RecordType != "sess" {
			offset -= 1 // Host & Url are together
		}
		url, _ = styleRunesWithOffset(url, offset, matchedRunes, styles.UrlStyle)
	} else if highlight {
		txid = styles.TxidStyle.Render(txid)
		recordType = styles.RecordTypeStyle.Render(recordType)
		parentId = styles.TxidStyle.Render(parentId)
		reason = styles.ReasonStyle.Render(reason)
		statusCode = styles.ReasonStyle.Render(statusCode)
		statusReason = styles.ReasonStyle.Render(statusReason)
		method = styles.MethodStyle.Render(method)
		host = styles.HostStyle.Render(host)
		url = styles.UrlStyle.Render(url)
	}

	return fmt.Sprintf(
		"%s %s %s %s %s %s\n%s %s%s",
		txid,
		recordType,
		parentId,
		reason,
		statusCode,
		statusReason,
		method,
		host,
		url,
	)
}

// AsItem
//
//	123 req 122 rxreq (200 OK)
//	GET www.example.com/path/to/asset
//	178µs total for Start(0s) → Fetch(140µs) → Process(6µs) → Resp(32µs)
func (t Tx) AsItem(matchedRunes []int, highlight, highlightMatches bool) string {
	var (
		tsFlow     string
		tsTotalDur time.Duration = 0
	)

	for _, ts := range t.Timestamps {
		tsTotalDur += ts.SinceLast
	}
	tsFlow = fmt.Sprintf("%s total for %s", tsTotalDur, t.printTimestampsFlow())

	if highlight {
		tsFlow = styles.TsFlowStyle.Render(tsFlow)
	}

	return fmt.Sprintf(
		"%s\n%s",
		t.AsString(matchedRunes, highlight, highlightMatches),
		tsFlow,
	)
}

func styleRunesWithOffset(s string, offset int, matchedRunes []int, style lipgloss.Style) (string, int) {
	var (
		length        int = len(s)
		newOffset     int
		filteredRunes []int

		unmatched lipgloss.Style = style
		matched   lipgloss.Style = style.Inherit(styles.MatchedItemStyle)
	)

	// Adjust the indices by subtracting the offset and filter out invalid indices
	for _, index := range matchedRunes {
		adjustedIndex := index - int(offset)
		if adjustedIndex >= 0 && adjustedIndex < length {
			filteredRunes = append(filteredRunes, adjustedIndex)
		}
	}

	// current offset + string len + space
	newOffset = offset + length + 1

	// Return the styled string
	return lipgloss.StyleRunes(s, filteredRunes, matched, unmatched), newOffset
}

// FindRootParent returns the parent of the chain of txs
func (t Tx) FindRootParent() *Tx {
	if t.Parent == nil {
		return &t
	}
	return t.Parent.FindRootParent()
}

// SumOfSinceLast returns the sum of all the SinceLast of this Tx Timestamps
func (t Tx) SumOfSinceLast() time.Duration {
	var total time.Duration
	for _, ts := range t.Timestamps {
		total += ts.SinceLast
	}
	return total
}

// GetSortedChildren retrieves all children and their descendants and returns them sorted by Txid.
func (t *Tx) GetSortedChildren() []*Tx {
	allChildren := make(map[string]*Tx)
	t.collectChildren(allChildren)

	childrenSlice := make([]*Tx, 0, len(allChildren))
	for _, child := range allChildren {
		if child != nil {
			childrenSlice = append(childrenSlice, child)
		}
	}

	sort.Slice(childrenSlice, func(i, j int) bool {
		return childrenSlice[i].Txid < childrenSlice[j].Txid
	})

	return childrenSlice
}

// collectChildren is a helper function to recursively collect all children and their descendants.
func (t *Tx) collectChildren(allChildren map[string]*Tx) {
	if t.Children == nil {
		return
	}
	for _, child := range t.Children {
		if child != nil {
			allChildren[child.Txid] = child
			child.collectChildren(allChildren)
		}
	}
}
