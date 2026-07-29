// hent plukkar teksten ut av Word-fila med tabellar og kursiv i behald,
// og legg han i ei json-fil som resten av verktøya arbeider på.
//
// Vegen går om wvHtml, av di det er den einaste konverteraren vi har
// funne som tek vare på både tabellane og kursiven i den binære
// Word 95-fila. textutil kastar begge deler.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"aasen/internal/tekst"
)

func main() {
	doc := flag.String("doc", "bøker/Norsk Grammatik.doc", "Word-fila")
	ut := flag.String("ut", "bøker/grammatik.json", "json-fil")
	behald := flag.String("html", "", "ta vare på mellomsteget som html her")
	flag.Parse()

	htmlsti := *behald
	if htmlsti == "" {
		mappe, err := os.MkdirTemp("", "aasen")
		if err != nil {
			stopp(err)
		}
		defer os.RemoveAll(mappe)
		htmlsti = filepath.Join(mappe, "ng.html")
	}

	cmd := exec.Command("wvHtml", "--charset=utf-8", *doc, htmlsti)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		stopp(fmt.Errorf("wvHtml: %w (er han installert? brew install wv)", err))
	}

	fil, err := os.Open(htmlsti)
	if err != nil {
		stopp(err)
	}
	defer fil.Close()

	blokker, err := tekst.HentFråHTML(fil)
	if err != nil {
		stopp(err)
	}

	skriv(*ut, blokker)
	meld(blokker, *ut)
}

func skriv(sti string, blokker []tekst.Blokk) {
	fil, err := os.Create(sti)
	if err != nil {
		stopp(err)
	}
	defer fil.Close()
	enc := json.NewEncoder(fil)
	enc.SetIndent("", " ")
	if err := enc.Encode(blokker); err != nil {
		stopp(err)
	}
}

func meld(blokker []tekst.Blokk, sti string) {
	var oppsett, kursiv, feit int
	for _, b := range blokker {
		if b.Tabell != nil {
			oppsett++
		}
		for _, s := range b.Stumpar {
			if s.Kursiv {
				kursiv++
			}
			if s.Feit {
				feit++
			}
		}
	}
	fmt.Printf("%d blokker, %d oppsett, %d kursivstumpar, %d feite -> %s\n",
		len(blokker), oppsett, kursiv, feit, sti)
}

func stopp(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
