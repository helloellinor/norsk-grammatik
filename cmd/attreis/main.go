// attreis set lang ſ attende inn i den reine doc-teksten, ved å slå opp
// korleis kvart ordform faktisk vart sett i den trykte 1864-utgåva.
// Ord vi ikkje finn i trykken fell attende på trykkregelen: rund s i
// slutten av ordet, lang ſ elles.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"aasen/internal/tekst"
)

const langS = 'ſ' // U+017F LATIN SMALL LETTER LONG S

var sidemønster = regexp.MustCompile(`side-(\d+)\.txt$`)

// oppslag er det vi har lært om eitt ordform frå trykken.
type oppslag struct {
	Form    string // s/ſ-mønsteret i trykken, like langt som ordet
	Tal     int    // kor mange gonger den forma stod der
	Ialt    int    // kor mange gonger ordet stod der i det heile
	Usikker bool   // fleire skrivemåtar med jamn fordeling
}

// leksikon slår opp på to måtar: eksakt, og med vokalane folda saman.
// Det siste trengst av di OCR-modellen er tysk og les æ som e og ø som o
// (ſedvanlig for sædvanlig, Hunkjonsord for Hunkjønsord), så ord med
// norske vokalar elles aldri ville finne seg sjølve i trykken.
type leksikon struct {
	Eksakt map[string]oppslag
	Folda  map[string]oppslag
}

func (l leksikon) slåOpp(ord []rune) (oppslag, bool) {
	if o, ok := l.Eksakt[normaliser(ord)]; ok && len([]rune(o.Form)) == len(ord) {
		return o, true
	}
	if o, ok := l.Folda[foldnykel(ord)]; ok && len([]rune(o.Form)) == len(ord) {
		return o, true
	}
	return oppslag{}, false
}

func main() {
	sidemappe := flag.String("sider", "bøker/sider", "mappe med OCR-tekst frå trykken")
	tekstfil := flag.String("tekst", "bøker/grammatik.json", "blokker frå Word-fila")
	utfil := flag.String("ut", "bøker/grammatik-sats.json", "blokker med ſ sett inn")
	rapportfil := flag.String("rapport", "bøker/sats-rapport.txt", "liste over ord som treng ettersyn")
	flag.Parse()

	leks, err := byggLeksikon(*sidemappe)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("leksikon frå trykken: %d ordformer (%d med folda vokalar)\n",
		len(leks.Eksakt), len(leks.Folda))

	rå, err := os.ReadFile(*tekstfil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var blokker []tekst.Blokk
	if err := json.Unmarshal(rå, &blokker); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	stat := nyStatistikk()
	for i := range blokker {
		for j := range blokker[i].Stumpar {
			blokker[i].Stumpar[j].Tekst = settInnLangS(blokker[i].Stumpar[j].Tekst, leks, stat)
		}
		if blokker[i].Tabell == nil {
			continue
		}
		for r := range blokker[i].Tabell.Rader {
			for c := range blokker[i].Tabell.Rader[r].Celler {
				celle := blokker[i].Tabell.Rader[r].Celler[c]
				for k := range celle {
					celle[k].Tekst = settInnLangS(celle[k].Tekst, leks, stat)
				}
			}
		}
	}

	utfila, err := os.Create(*utfil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(utfila)
	enc.SetIndent("", " ")
	if err := enc.Encode(blokker); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	utfila.Close()

	skrivRapport(*rapportfil, stat)
	stat.skriv(*utfil, *rapportfil)
}

type statistikk struct {
	OrdIalt       int
	FråLeksikon   int
	FråRegel      int
	SIalt         int
	SVartLang     int
	UkjendeOrd    map[string]int
	UsikreOrd     map[string]int
	RegelGavLangS map[string]int
}

func nyStatistikk() *statistikk {
	return &statistikk{
		UkjendeOrd:    map[string]int{},
		UsikreOrd:     map[string]int{},
		RegelGavLangS: map[string]int{},
	}
}

// settInnLangS går gjennom teksten teikn for teikn og byter berre ut
// små s-ar. Alt anna - teiknsetjing, linjeskift, store bokstavar - står
// urørt, slik at teksten elles er nøyaktig den same.
func settInnLangS(tekst string, leks leksikon, stat *statistikk) string {
	runar := []rune(tekst)
	ut := make([]rune, len(runar))
	copy(ut, runar)

	i := 0
	for i < len(runar) {
		if !erBokstav(runar[i]) {
			i++
			continue
		}
		start := i
		for i < len(runar) && erBokstav(runar[i]) {
			i++
		}
		ord := runar[start:i]
		if !harLitenS(ord) {
			continue
		}

		stat.OrdIalt++
		opp, funne := leks.slåOpp(ord)

		switch {
		case funne:
			stat.FråLeksikon++
			if opp.Usikker {
				stat.UsikreOrd[string(ord)]++
			}
			form := []rune(opp.Form)
			for j, r := range ord {
				if r != 's' {
					continue
				}
				stat.SIalt++
				if form[j] == langS {
					ut[start+j] = langS
					stat.SVartLang++
				}
			}
		default:
			stat.FråRegel++
			if !funne {
				stat.UkjendeOrd[string(ord)]++
			}
			for j, r := range ord {
				if r != 's' {
					continue
				}
				stat.SIalt++
				if j != len(ord)-1 {
					ut[start+j] = langS
					stat.SVartLang++
					stat.RegelGavLangS[string(ord)]++
				}
			}
		}
	}

	return string(ut)
}

// byggLeksikon les alle OCR-sidene og finn den vanlegaste skrivemåten
// for kvart ordform. Ord der skrivemåtane er jamt fordelte blir merkte
// som usikre, så dei kan sjåast over manuelt.
func byggLeksikon(mappe string) (leksikon, error) {
	filer, err := filepath.Glob(filepath.Join(mappe, "side-*.txt"))
	if err != nil {
		return leksikon{}, err
	}
	if len(filer) == 0 {
		return leksikon{}, fmt.Errorf("fann ingen sider i %s", mappe)
	}

	eksakt := map[string]map[string]int{}
	folda := map[string]map[string]int{}
	for _, fil := range filer {
		if !sidemønster.MatchString(fil) {
			continue
		}
		b, err := os.ReadFile(fil)
		if err != nil {
			continue
		}
		runar := []rune(string(b))
		i := 0
		for i < len(runar) {
			if !erBokstav(runar[i]) {
				i++
				continue
			}
			start := i
			for i < len(runar) && erBokstav(runar[i]) {
				i++
			}
			ord := runar[start:i]
			// Vi lagrar berre s/ſ-mønsteret, ikkje store/små bokstavar,
			// så "Sag" og "sag" kan hjelpe kvarandre.
			tel(eksakt, normaliser(ord), sMønster(ord))
			tel(folda, foldnykel(ord), sMønster(ord))
		}
	}

	return leksikon{Eksakt: velVanlegaste(eksakt), Folda: velVanlegaste(folda)}, nil
}

func tel(m map[string]map[string]int, nykel, form string) {
	if m[nykel] == nil {
		m[nykel] = map[string]int{}
	}
	m[nykel][form]++
}

func velVanlegaste(tal map[string]map[string]int) map[string]oppslag {
	ut := make(map[string]oppslag, len(tal))
	for nykel, former := range tal {
		beste, besteTal, ialt := "", 0, 0
		for form, n := range former {
			ialt += n
			if n > besteTal {
				beste, besteTal = form, n
			}
		}
		ut[nykel] = oppslag{
			Form:    beste,
			Tal:     besteTal,
			Ialt:    ialt,
			Usikker: ialt >= 2 && float64(besteTal)/float64(ialt) < 0.75,
		}
	}
	return ut
}

// sMønster gjev ein streng like lang som ordet, med ſ eller s på kvar
// s-plass og punktum elles. Då kan same mønsteret brukast om att på ord
// som er stava litt ulikt i OCR-en.
func sMønster(runar []rune) string {
	ut := make([]rune, len(runar))
	for i, r := range runar {
		switch {
		case r == langS:
			ut[i] = langS
		case unicode.ToLower(r) == 's':
			ut[i] = 's'
		default:
			ut[i] = '.'
		}
	}
	return string(ut)
}

// foldnykel gjer alle vokalar like, så tyskleste former som "ſedvanlig"
// og "Hunkjonsord" finn att "sædvanlig" og "Hunkjønsord".
func foldnykel(runar []rune) string {
	var sb strings.Builder
	for _, r := range runar {
		if r == langS {
			r = 's'
		}
		r = unicode.ToLower(r)
		if erVokal(r) {
			sb.WriteRune('·')
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func erVokal(r rune) bool {
	return strings.ContainsRune("aeiouyæøåäöüàáâèéêìíîòóôùúû", r)
}

func normaliser(runar []rune) string {
	var sb strings.Builder
	for _, r := range runar {
		if r == langS {
			r = 's'
		}
		sb.WriteRune(unicode.ToLower(r))
	}
	return sb.String()
}

func harLitenS(runar []rune) bool {
	for _, r := range runar {
		if r == 's' {
			return true
		}
	}
	return false
}

func erBokstav(r rune) bool { return unicode.IsLetter(r) }

func (s *statistikk) skriv(utfil, rapportfil string) {
	fmt.Printf("ord med liten s: %d\n", s.OrdIalt)
	fmt.Printf("  frå leksikon:  %d (%.1f%%)\n", s.FråLeksikon, prosent(s.FråLeksikon, s.OrdIalt))
	fmt.Printf("  frå trykkregel: %d (%.1f%%)\n", s.FråRegel, prosent(s.FråRegel, s.OrdIalt))
	fmt.Printf("s-førekomstar: %d, vart lang ſ: %d (%.1f%%)\n", s.SIalt, s.SVartLang, prosent(s.SVartLang, s.SIalt))
	fmt.Printf("ulike ukjende ord: %d, usikre oppslag: %d\n", len(s.UkjendeOrd), len(s.UsikreOrd))
	fmt.Printf("\nskrive: %s\nrapport: %s\n", utfil, rapportfil)
}

func prosent(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return 100 * float64(a) / float64(b)
}

func skrivRapport(sti string, stat *statistikk) {
	fil, err := os.Create(sti)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer fil.Close()
	w := bufio.NewWriter(fil)
	defer w.Flush()

	fmt.Fprintln(w, "# Ord som treng ettersyn")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Ikkje funne i trykken - trykkregelen er brukt")
	fmt.Fprintln(w, "# (ord, kor ofte i teksten)")
	for _, r := range sorter(stat.UkjendeOrd) {
		fmt.Fprintf(w, "%-40s %d\n", r.Ord, r.Tal)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Funne, men trykken er ikkje eintydig")
	for _, r := range sorter(stat.UsikreOrd) {
		fmt.Fprintf(w, "%-40s %d\n", r.Ord, r.Tal)
	}
}

type rad struct {
	Ord string
	Tal int
}

func sorter(m map[string]int) []rad {
	var rader []rad
	for o, t := range m {
		rader = append(rader, rad{o, t})
	}
	sort.Slice(rader, func(i, j int) bool {
		if rader[i].Tal != rader[j].Tal {
			return rader[i].Tal > rader[j].Tal
		}
		return rader[i].Ord < rader[j].Ord
	})
	return rader
}
