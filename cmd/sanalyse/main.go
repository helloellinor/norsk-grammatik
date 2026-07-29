// sanalyse tel opp i kva bokstavkontekst rund s og lang ſ dukkar opp,
// basert på OCR-tekst frå faktiske trykte sider. Brukast til å byggje
// eit oppslagsgrunnlag for å setje rett s-form inn i rein transkribert tekst.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const langS = 'ſ' // U+017F LATIN SMALL LETTER LONG S

type teljing struct {
	Rund int
	Lang int
}

func main() {
	mappe := flag.String("mappe", "sider", "mappe med OCR-tekstfiler (.txt)")
	vindauge := flag.Int("vindauge", 1, "kor mange bokstavar før/etter s som skal brukast som kontekst")
	flag.Parse()

	filer, err := filepath.Glob(filepath.Join(*mappe, "*.txt"))
	if err != nil || len(filer) == 0 {
		fmt.Fprintln(os.Stderr, "fann ingen .txt-filer i", *mappe)
		os.Exit(1)
	}

	tal := map[string]*teljing{}
	for _, fil := range filer {
		bytes, err := os.ReadFile(fil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kunne ikkje lese %s: %v\n", fil, err)
			continue
		}
		handsamTekst(string(bytes), *vindauge, tal)
	}

	skrivRapport(tal)
}

func handsamTekst(tekst string, vindauge int, tal map[string]*teljing) {
	runar := []rune(tekst)
	for i, r := range runar {
		if r != 's' && r != langS {
			continue
		}
		før := kontekst(runar, i-vindauge, i)
		etter := kontekst(runar, i+1, i+1+vindauge)
		key := før + "_" + etter
		if tal[key] == nil {
			tal[key] = &teljing{}
		}
		if r == langS {
			tal[key].Lang++
		} else {
			tal[key].Rund++
		}
	}
}

func kontekst(runar []rune, frå, til int) string {
	if frå < 0 {
		frå = 0
	}
	if til > len(runar) {
		til = len(runar)
	}
	if frå >= til {
		return ""
	}
	return string(runar[frå:til])
}

type rad struct {
	Kontekst string `json:"kontekst"`
	Rund     int    `json:"rund"`
	Lang     int    `json:"lang"`
}

func skrivRapport(tal map[string]*teljing) {
	rader := make([]rad, 0, len(tal))
	for k, v := range tal {
		rader = append(rader, rad{k, v.Rund, v.Lang})
	}
	sort.Slice(rader, func(i, j int) bool {
		return rader[i].Rund+rader[i].Lang > rader[j].Rund+rader[j].Lang
	})
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(rader)
}
