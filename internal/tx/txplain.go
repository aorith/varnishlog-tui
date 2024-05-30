package tx

import (
	"fmt"
	"strings"
	"time"

	"github.com/aorith/varnishlog-tui/internal/util"
)

// printTimestampsFlow returns a string represening the sequence of the timestamps and their durations
func (t Tx) printTimestampsFlow() string {
	var (
		tsParts []string
		tsFlow  string
	)

	if len(t.Timestamps) <= 0 {
		return ""
	}

	for _, ts := range t.Timestamps {
		tsParts = append(tsParts, fmt.Sprintf("%s(%s)", ts.EventLabel, ts.SinceLast))
	}
	tsFlow = strings.Join(tsParts, " → ")

	return tsFlow
}

// PrintTree prints the tree structure starting from the current transaction
func (t Tx) PrintTree(prefix, selectedTxid string, isTail bool) string {
	var (
		builder strings.Builder
		txInfo  string
		spacing string = "    "
	)

	if selectedTxid == t.Txid {
		txInfo = fmt.Sprintf("%s* %s\n", t.Txid, t.Reason)
	} else {
		txInfo = fmt.Sprintf("%s %s\n", t.Txid, t.Reason)
	}

	if t.Parent == nil {
		spacing = ""
		builder.WriteString(txInfo)
	} else if isTail || t.Parent == nil {
		builder.WriteString(prefix + "└── " + txInfo)
	} else {
		builder.WriteString(prefix + "├── " + txInfo)
	}

	children := make([]*Tx, 0, len(t.Children))
	for _, child := range t.Children {
		children = append(children, child)
	}

	for i := 0; i < len(children)-1; i++ {
		builder.WriteString(children[i].PrintTree(prefix+spacing, selectedTxid, false))
	}
	// Last children
	if len(children) > 0 {
		builder.WriteString(children[len(children)-1].PrintTree(prefix+spacing, selectedTxid, true))
	}

	return builder.String()
}

// GenerateTimestampHistogram generates an ASCII string with the histogram of the timestamps of this tx
func (t Tx) GenerateTimestampHistogram() string {
	headers := []string{"Event", "Duration"}

	rows := make([]util.RowValue, len(t.Timestamps))

	var total, current time.Duration
	for i, ts := range t.Timestamps {
		rows[i].Row = []string{ts.EventLabel}
		current = ts.SinceLast
		total += current
		rows[i].Value = util.DurationValue(current)
	}

	rowValues := util.RowValues{
		Rows:  rows,
		Total: util.DurationValue(total),
	}

	return util.GenerateHistogram(headers, rowValues)
}

// GenerateAccountingHistogram generates an ASCII string with the accounting histogram of all the Txs
func (t Tx) GenerateAccountingHistogram(txs []*Tx, transmitted bool) string {
	headers := []string{"TxId", "Type", "Reason", "Header", "Body", "Sum"}

	rows := make([]util.RowValue, len(txs))

	var total, rowSum int64
	for i, tx := range txs {
		var row []string
		if t.Txid == tx.Txid {
			row = append(row, fmt.Sprintf("%s*", tx.Txid))
		} else {
			row = append(row, tx.Txid)
		}

		row = append(row, tx.RecordType)
		row = append(row, tx.Reason)

		if tx.RecordType == "sess" {
			row = append(row, "-")
			row = append(row, "-")
			rowSum = 0
		} else {
			if transmitted {
				row = append(row, tx.Accounting.HeaderBytesTransmitted.String())
				row = append(row, tx.Accounting.BodyBytesTransmitted.String())
				rowSum = tx.Accounting.HeaderBytesTransmitted.Value() + tx.Accounting.BodyBytesTransmitted.Value()
			} else {
				row = append(row, tx.Accounting.HeaderBytesReceived.String())
				row = append(row, tx.Accounting.BodyBytesReceived.String())
				rowSum = tx.Accounting.HeaderBytesReceived.Value() + tx.Accounting.BodyBytesReceived.Value()
			}
			total += rowSum
		}

		rows[i].Row = row
		rows[i].Value = util.SizeValue(rowSum)
	}

	rowValues := util.RowValues{
		Rows:  rows,
		Total: util.SizeValue(total),
	}

	return util.GenerateHistogram(headers, rowValues)
}

// GenerateAllTxsHistogram generates an ASCII string with the histogram of all the Txs
func (t Tx) GenerateAllTxsHistogram(txs []*Tx) string {
	headers := []string{"TxId", "Type", "Reason", "Duration"}

	rows := make([]util.RowValue, len(txs))

	var total, current time.Duration
	for i, tx := range txs {
		var row []string
		if t.Txid == tx.Txid {
			row = append(row, fmt.Sprintf("%s*", tx.Txid))
		} else {
			row = append(row, tx.Txid)
		}
		row = append(row, tx.RecordType)
		row = append(row, tx.Reason)
		rows[i].Row = row
		current = tx.SumOfSinceLast()
		rows[i].Value = util.DurationValue(current)
		total += current
	}

	rowValues := util.RowValues{
		Rows:  rows,
		Total: util.DurationValue(total),
	}

	return util.GenerateHistogram(headers, rowValues)
}
