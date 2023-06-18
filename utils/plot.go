package utils

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"
)

func PlotLine(base []float64, prediction []float64, fileName string) error {
	p := plot.New()
	baseLine, err := plotter.NewLine(make(plotter.XYs, len(base)))
	if err != nil {
		return err
	}
	predtionLine, err := plotter.NewLine(make(plotter.XYs, len(prediction)))
	if err != nil {
		return err
	}
	connectionLine, err := plotter.NewLine(make(plotter.XYs, 2))
	if err != nil {
		return err
	}
	connectionLine.XYs[0].X = float64(len(base) - 1)
	connectionLine.XYs[0].Y = float64(base[len(base)-1])

	connectionLine.XYs[1].X = float64(len(base))
	connectionLine.XYs[1].Y = float64(prediction[0])
	for i, v := range base {
		baseLine.XYs[i].X = float64(i)
		baseLine.XYs[i].Y = float64(v)
	}

	for i, v := range prediction {
		predtionLine.XYs[i].X = float64(i + len(base))
		predtionLine.XYs[i].Y = float64(v)
	}
	baseLine.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	connectionLine.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	predtionLine.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	p.Add(baseLine)
	p.Add(predtionLine)
	p.Add(connectionLine)
	p.X.Label.Text = "time stamp"
	p.Y.Label.Text = "value"
	if err := p.Save(4*vg.Inch, 4*vg.Inch, fmt.Sprintf("%s.png", fileName)); err != nil {
		return err
	}
	return nil
}
