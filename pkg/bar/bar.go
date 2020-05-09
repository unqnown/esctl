package bar

import (
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func Docs(total int64, name string) (*mpb.Bar, func()) {
	p := mpb.New(mpb.WithWidth(100))
	done := "ʕ•ᴥ•ʔ"
	bar := p.AddBar(total,
		// override DefaultBarStyle, which is "[=>-]<+"
		//mpb.BarStyle("╢▌▌░╟"),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 9}), done,
			),
		),
		mpb.AppendDecorators(decor.CountersNoUnit("%d / %d")),
	)
	return bar, p.Wait
}

func Percent(total int64, name string) (*mpb.Bar, func()) {
	p := mpb.New(mpb.WithWidth(100))
	done := "ʕ•ᴥ•ʔ"
	bar := p.AddBar(total,
		// override DefaultBarStyle, which is "[=>-]<+"
		//mpb.BarStyle("╢▌▌░╟"),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 9}), done,
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	return bar, p.Wait
}
