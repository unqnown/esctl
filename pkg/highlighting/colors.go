package highlighting

import "github.com/olekukonko/tablewriter"

var Health = map[string]tablewriter.Colors{
	"green":  {tablewriter.FgGreenColor},
	"yellow": {tablewriter.FgYellowColor},
	"red":    {tablewriter.FgRedColor},
}
