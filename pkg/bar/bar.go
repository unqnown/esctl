package bar

import (
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func Docs(total int64, name string) (*mpb.Bar, func()) {
	p := mpb.New(mpb.WithWidth(80))
	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			decor.Elapsed(decor.ET_STYLE_HHMMSS, decor.WC{W: 9}),
			decor.Name("/"),
			decor.AverageETA(decor.ET_STYLE_HHMMSS, decor.WC{W: 9, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(decor.CountersNoUnit("%d / %d")),
	)

	return bar, p.Wait
}

func Percent(total int64, name string) (*mpb.Bar, func()) {
	p := mpb.New(mpb.WithWidth(80))
	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			decor.Elapsed(decor.ET_STYLE_HHMMSS, decor.WC{W: 9}),
			decor.Name("/"),
			decor.AverageETA(decor.ET_STYLE_HHMMSS, decor.WC{W: 9, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)

	return bar, p.Wait
}
