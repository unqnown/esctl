package highlighting

import "github.com/olekukonko/tablewriter"

var Health = map[string]tablewriter.Colors{
	"green":  {tablewriter.FgGreenColor},
	"yellow": {tablewriter.FgYellowColor},
	"red":    {tablewriter.FgRedColor},
}

var State = map[string]tablewriter.Colors{
	"STARTED":      {tablewriter.FgGreenColor},
	"INITIALIZING": {tablewriter.FgYellowColor},
	"RELOCATING":   {tablewriter.FgYellowColor},
	"UNASSIGNED":   {tablewriter.FgRedColor},
}
