// vev lagar ei lesbar nettutgåve av teksten: tolkar blokkstrukturen og
// set han i fraktur, med skriftfila baka inn i sida so ho står på eiga
// hand utan nettkopling.
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"
)

type Side struct {
	Verk      string
	Undertitl string
	Skrift    template.CSS
	Bolkar    []Bolk
	Nake      bool // utan html/head/body, til innbygging
}

func main() {
	inn := flag.String("inn", "bøker/norsk_grammatik_sats.txt", "tekst med ſ sett inn")
	ut := flag.String("ut", "vev/grammatik.html", "html-fil")
	skriftfil := flag.String("skrift", "skrift/UnifrakturMaguntia-Book.ttf", "frakturskrift")
	frå := flag.String("frå", "", "start ved fyrste blokk som inneheld denne teksten")
	til := flag.String("til", "", "stopp ved fyrste blokk som inneheld denne teksten")
	nake := flag.Bool("nake", false, "utan html/head/body, til innbygging")
	flag.Parse()

	rå, err := os.ReadFile(*inn)
	if err != nil {
		stopp(err)
	}
	skrift, err := os.ReadFile(*skriftfil)
	if err != nil {
		stopp(err)
	}

	bolkar := avgrens(Tolk(string(rå)), *frå, *til)
	if len(bolkar) == 0 {
		stopp(fmt.Errorf("ingen blokker att etter avgrensing"))
	}

	side := Side{
		Verk:      "Norſk Grammatik",
		Undertitl: "Ivar Aasen · Chriſtiania 1864",
		Skrift: template.CSS("data:font/ttf;base64," +
			base64.StdEncoding.EncodeToString(skrift)),
		Bolkar: bolkar,
		Nake:   *nake,
	}

	if err := os.MkdirAll(mappeAv(*ut), 0o755); err != nil {
		stopp(err)
	}
	fil, err := os.Create(*ut)
	if err != nil {
		stopp(err)
	}
	defer fil.Close()

	if err := mal.Execute(fil, side); err != nil {
		stopp(err)
	}

	fmt.Printf("%d blokker skrivne til %s\n", len(bolkar), *ut)
}

func avgrens(bolkar []Bolk, frå, til string) []Bolk {
	start, slutt := 0, len(bolkar)
	if frå != "" {
		for i, b := range bolkar {
			if strings.Contains(b.Tittel, frå) || strings.Contains(b.Tekst, frå) {
				start = i
				break
			}
		}
	}
	if til != "" {
		for i := start + 1; i < len(bolkar); i++ {
			if strings.Contains(bolkar[i].Tittel, til) || strings.Contains(bolkar[i].Tekst, til) {
				slutt = i
				break
			}
		}
	}
	return bolkar[start:slutt]
}

func mappeAv(sti string) string {
	if i := strings.LastIndex(sti, "/"); i > 0 {
		return sti[:i]
	}
	return "."
}

func stopp(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

var mal = template.Must(template.New("side").Parse(malTekst))
