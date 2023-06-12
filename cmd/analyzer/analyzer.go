package main

import (
	"io"
	"log"
	"os"
	"time"

	e "github.com/NickBabakin/ipiad/res/elasticgo"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func drawProfessions(devs int, analysts int, architects int) {
	groupA := plotter.Values{float64(devs), float64(analysts),
		float64(architects)}

	p := plot.New()

	p.Title.Text = "Распределение вакансий по профессиям"
	p.Y.Label.Text = "Количество вакансий"

	w := vg.Points(40)

	barsA, err := plotter.NewBarChart(groupA, w)
	if err != nil {
		panic(err)
	}
	barsA.LineStyle.Width = vg.Length(0)
	barsA.Color = plotutil.Color(0)
	barsA.Offset = -w + 50

	p.Add(barsA)
	p.Legend.Top = true
	p.NominalX("Разработчик", "Аналитик", "Архитектор")

	if err := p.Save(5*vg.Inch, 3*vg.Inch, "barchart.png"); err != nil {
		panic(err)
	}
}

func main() {

	os.MkdirAll("logs", 0750)
	logFile, err := os.OpenFile("logs/log_analyzer.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	log.Println("Analyzer waits for data")
	time.Sleep(time.Second * 90)
	for {
		err := e.Init_elastic()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 30)
	}

	log.Println("Total:")
	e.SearchAllVacancies()
	log.Println("MTS developers:")
	e.SearchMtsDevVacancies()
	log.Println("Developers:")
	devs := e.SearchProfessionVacancies(e.Developer)
	log.Println("Analysts:")
	analysts := e.SearchProfessionVacancies(e.Analyst)
	log.Println("Architects:")
	architects := e.SearchProfessionVacancies(e.Architect)
	drawProfessions(devs, analysts, architects)

}
