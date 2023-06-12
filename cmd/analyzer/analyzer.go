package main

import (
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	e "github.com/NickBabakin/ipiad/res/elasticgo"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/cdipaolo/goml/base"
	goml "github.com/cdipaolo/goml/text"
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

func shuffleSlice(sl []e.VacancieProfession) {
	rand.Seed(time.Now().UnixNano())
	for i := range sl {
		j := rand.Intn(i + 1)
		sl[i], sl[j] = sl[j], sl[i]
	}
}

func ai(allVacancieProfessions []e.VacancieProfession) {
	testLimit := 500
	stream := make(chan base.TextDatapoint, testLimit)
	errors := make(chan error)

	model := goml.NewNaiveBayes(stream, 4, base.OnlyWordsAndNumbers)

	go model.OnlineLearn(errors)

	trainVacancieProfessions := allVacancieProfessions[:len(allVacancieProfessions)-20]
	testVacancieProfessions := allVacancieProfessions[len(allVacancieProfessions)-20:]

	for _, profession := range trainVacancieProfessions {
		stream <- base.TextDatapoint{
			X: profession.Title,
			Y: uint8(profession.Profession),
		}
	}

	close(stream)

	for {
		err := <-errors
		if err != nil {
			log.Printf("Error passed: %v", err)
		} else {
			// training is done!
			break
		}
	}

	successCases := 0
	for _, vacancieProfession := range testVacancieProfessions {
		predictedProfession := e.Profession(model.Predict(vacancieProfession.Title))
		log.Printf("Title: %s\n   Actual profession    : %s\n   Predicted profession : %s\n",
			vacancieProfession.Title,
			e.ProffessionToString[vacancieProfession.Profession],
			e.ProffessionToString[predictedProfession])
		if vacancieProfession.Profession == predictedProfession {
			successCases++
		}
	}
	log.Printf("\n\n*** Number of train data : %d\n*** Prediction attempts  : %d\n*** Matches              : %d\n*** Mismatches           : %d\n*** Success rate         : %.1f%%\n",
		len(trainVacancieProfessions),
		len(testVacancieProfessions),
		successCases,
		len(testVacancieProfessions)-successCases,
		100*float64(successCases)/float64(len(testVacancieProfessions)))
}

func analyze() {
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
	log.Println("Other:")
	other := e.SearchProfessionVacancies(e.Other)
	drawProfessions(len(devs), len(analysts), len(architects))

	allVacancieProfessions := append(analysts, architects...)
	allVacancieProfessions = append(allVacancieProfessions, devs...)
	allVacancieProfessions = append(allVacancieProfessions, other...)
	shuffleSlice(allVacancieProfessions)
	log.Println("\n\nListing all professions:")
	for _, profession := range allVacancieProfessions {
		log.Printf("%-10s: %s", e.ProffessionToString[profession.Profession], profession.Title)
	}

	ai(allVacancieProfessions)
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

	analyze()

}
