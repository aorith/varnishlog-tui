package assets

import _ "embed"

//go:embed queries/built-in.yaml
var BuiltInQueries string

//go:embed templates/report.html
var ReportTemplate string
