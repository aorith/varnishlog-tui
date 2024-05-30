package tx

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/aorith/varnishlog-tui/assets"
)

type report struct {
	TxsTotalTimeHistogram string
	TxsStateDiagram       string
	AccountingReceived    string
	AccountingTransmitted string
	Txs                   []reportTx
}

type reportTx struct {
	Txid               string
	RecordType         string
	RawTx              string
	TimestampHistogram string
	TransitionsDiagram string
	TxInfoTable        []verticalTableRow
	TTLTable           horizontalTable
}

type horizontalTable struct {
	Headers []string
	Rows    [][]string
}

type verticalTableRow struct {
	Header string
	Values []string
}

// generateAllTxsDiagram generates a string representing a mermaid diagram of all the Txs
func (t Tx) generateAllTxsDiagram(parent *Tx) string {
	var s strings.Builder

	relationships := make(map[string]struct{})
	subgraphs := make(map[string]string)
	collectDiagramRelationships(parent, relationships, subgraphs)

	s.WriteString("\nflowchart TD\n")
	for _, sub := range subgraphs {
		s.WriteString(fmt.Sprintf("%s\n", sub))
	}
	for rel := range relationships {
		s.WriteString(fmt.Sprintf("%s\n", rel))
	}

	return s.String()
}

// collectDiagramRelationships is a helper to construct the mermaid diagram
func collectDiagramRelationships(tx *Tx, relationships map[string]struct{}, subgraphs map[string]string) {
	// A relationship is represented like this:
	// TxId1 ==> TxId2
	if tx.Parent != nil {
		rel := fmt.Sprintf(
			"    %s== \"%s\" ==>%s",
			tx.Parent.Txid,
			tx.Reason,
			tx.Txid,
		)
		relationships[rel] = struct{}{}
	}

	// The subgraphs map contains the TxID as the key and the subgraph as the value
	if _, exists := subgraphs[tx.Txid]; !exists {
		var style string
		if tx.RecordType == "sess" {
			style = "    style " + tx.Txid + " fill:#fafce6,stroke:#666666,stroke-width:1px"
		} else if tx.RecordType == "req" {
			style = "    style " + tx.Txid + " fill:#fcf2e6,stroke:#666666,stroke-width:1px"
		} else {
			style = "    style " + tx.Txid + " fill:#fce7e6,stroke:#666666,stroke-width:1px"
		}

		// Timestamp events: Start --> Fetch --> Connected ...
		var (
			tsEvents                       []string
			event, tsEventsRel, eventStyle string
		)

		if len(tx.Timestamps) > 0 {
			for i, ts := range tx.Timestamps {
				event = fmt.Sprintf("%s_%d_%s(%s\n%s)", ts.EventLabel, i, tx.Txid, ts.EventLabel, ts.SinceLast)
				tsEvents = append(tsEvents, event)
				eventStyle += fmt.Sprintf("    style %s_%d_%s fill:#fafafa,color:#333333,stroke:#111111,stroke-width:1px;\n", ts.EventLabel, i, tx.Txid)
			}
		}
		tsEventsRel = strings.Join(tsEvents, "-->")

		subgraph := fmt.Sprintf(
			"%s\n%s\n%s\n%s\n%s\n%s\n",
			fmt.Sprintf("subgraph %s[\"`&nbsp;**%s %s**&nbsp;\n`\"]", tx.Txid, tx.Txid, tx.RecordType),
			"    direction LR",
			"    "+tsEventsRel,
			"    end",
			eventStyle,
			style,
		)
		subgraphs[tx.Txid] = subgraph
	}

	for _, child := range tx.Children {
		if child != nil {
			collectDiagramRelationships(child, relationships, subgraphs)
		}
	}
}

// newTxInfoTable generates an HTML table with basic info about the tx.
// if the tx is a session nothing is returned
func (t Tx) newTxInfoTable() []verticalTableRow {
	if t.RecordType == "sess" {
		return nil
	}

	// values
	var parentTxid string
	if t.Parent != nil {
		parentTxid = t.Parent.Txid
	} else {
		parentTxid = "-"
	}

	var children []string
	for c := range t.Children {
		children = append(children, c)
	}
	var childrenStr string = "-"
	if len(children) > 0 {
		childrenStr = strings.Join(children, ", ")
	}

	var rows []verticalTableRow
	rows = append(rows,
		verticalTableRow{Header: "Parent", Values: []string{parentTxid}},
		verticalTableRow{Header: "Reason", Values: []string{t.Reason}},
		verticalTableRow{Header: "Request", Values: []string{fmt.Sprintf("%s %s%s", t.Method, t.Host, t.Url)}},
		verticalTableRow{Header: "Status", Values: []string{fmt.Sprintf("%d %s", t.StatusCode, t.StatusReason)}},
		verticalTableRow{Header: "Children", Values: []string{childrenStr}},
	)

	return rows
}

func (t Tx) newTxTTLTable() (headers []string, rows [][]string) {
	if t.RecordType == "ses" || len(t.TTL) <= 0 {
		return nil, nil
	}

	headers = append(headers, "Source", "TTL", "Grace", "Keep", "Reference", "Age", "Date", "Expires", "MaxAge", "CacheStatus")

	for _, ttl := range t.TTL {
		var row []string
		if ttl.Source == "HFP" {
			row = append(row,
				ttl.Source,
				fmt.Sprintf("%d", ttl.TTL),
				fmt.Sprintf("%d", ttl.Grace),
				fmt.Sprintf("%d", ttl.Keep),
				ttl.Reference.String(),
				fmt.Sprintf("%d", ttl.Age),
				ttl.Date.String(),
				ttl.Expires.String(),
				fmt.Sprintf("%d", ttl.MaxAge),
				ttl.CacheStatus,
			)
		} else {
			row = append(row,
				ttl.Source,
				fmt.Sprintf("%d", ttl.TTL),
				fmt.Sprintf("%d", ttl.Grace),
				fmt.Sprintf("%d", ttl.Keep),
				ttl.Reference.String(),
				"-",
				"-",
				"-",
				"-",
				ttl.CacheStatus,
			)
		}
		rows = append(rows, row)
	}

	return headers, rows
}

// generateTransitionsDiagram generates a mermaid diagram with the VCL transition states
func (t Tx) generateTransitionsDiagram() string {
	if t.RecordType == "sess" || len(t.Transitions) <= 0 {
		return ""
	}

	lastTransitionId := ""
	lastTransitionReturn := ""
	transitions := ""
	for _, tr := range t.Transitions {
		if lastTransitionId != "" {
			transitions += fmt.Sprintf("%s --> %s: <em>%s</em>\n", lastTransitionId, tr.Call, lastTransitionReturn)
		}
		lastTransitionId = tr.Call
		lastTransitionReturn = tr.Return
	}
	transitions += fmt.Sprintf("%s --> [*]: <em>%s</em>\n", lastTransitionId, lastTransitionReturn)

	return fmt.Sprintf(
		"stateDiagram\ndirection LR\n%s",
		transitions,
	)
}

func (t Tx) GenerateHtmlReport() ([]string, error) {
	// All Txs starting from the parent Tx
	parent := t.FindRootParent()
	txs := []*Tx{parent}
	txs = append(txs, parent.GetSortedChildren()...)

	repTxs := make([]reportTx, len(txs))
	for i, tx := range txs {
		repTx := reportTx{
			Txid:               tx.Txid,
			RecordType:         tx.RecordType,
			TimestampHistogram: tx.GenerateTimestampHistogram(),
			RawTx:              strings.Join(tx.RawTx, "\n"),
			TxInfoTable:        tx.newTxInfoTable(),
			TransitionsDiagram: tx.generateTransitionsDiagram(),
		}

		if tx.RecordType != "sess" {
			ttlHeaders, ttlRows := tx.newTxTTLTable()
			repTx.TTLTable = horizontalTable{
				Headers: ttlHeaders,
				Rows:    ttlRows,
			}
		}

		repTxs[i] = repTx
	}

	report := report{
		TxsTotalTimeHistogram: t.GenerateAllTxsHistogram(txs),
		TxsStateDiagram:       t.generateAllTxsDiagram(parent),
		AccountingReceived:    t.GenerateAccountingHistogram(txs, false),
		AccountingTransmitted: t.GenerateAccountingHistogram(txs, true),
		Txs:                   repTxs,
	}

	tmpl, err := template.New("report").Parse(assets.ReportTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, report)
	if err != nil {
		return nil, err
	}

	html := strings.Split(buf.String(), "\n")

	return html, nil
}
