// valider testar ſ/s-kontekstmodellen med leave-one-side-out: for kvar
// side byggjer vi oppslagstabellen frå dei andre sidene, og listar opp
// kvar spådommen bommar på den verkelege skanna sida, med sidetal og
// tekstsamanheng, slik at avvika kan sjekkast manuelt mot faksimilen.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const langS = 'ſ' // U+017F LATIN SMALL LETTER LONG S

type teljing struct {
	Rund int
	Lang int
}

type side struct {
	Nummer int
	Runar  []rune
}

type avvik struct {
	Side     int
	Kontekst string
	Venta    rune
	Fekk     rune
}

var sidemønster = regexp.MustCompile(`side-(\d+)\.txt$`)

var vindauge = 1

func main() {
	mappe := "bøker/sider"
	if len(os.Args) > 1 {
		mappe = os.Args[1]
	}
	if len(os.Args) > 2 {
		if v, err := strconv.Atoi(os.Args[2]); err == nil {
			vindauge = v
		}
	}

	sider, err := lastSider(mappe)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(sider) == 0 {
		fmt.Fprintln(os.Stderr, "fann ingen sider i", mappe)
		os.Exit(1)
	}

	ordbok, err := lastOrdbok("bøker/norsk_grammatik.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "kunne ikkje laste ordliste:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "ordliste: %d ord\n", len(ordbok))

	var avvika []avvik
	rettTal, altTal := 0, 0

	for _, s := range sider {
		tabellVid := byggTabellMed(sider, s.Nummer, vindauge)
		tabellSmal := byggTabellMed(sider, s.Nummer, 1)
		for i, r := range s.Runar {
			if r != 's' && r != langS {
				continue
			}
			altTal++
			etter1 := kontekstStr(s.Runar, i+1, i+2)
			keyVid := kontekstStr(s.Runar, i-vindauge, i) + "_" + kontekstStr(s.Runar, i+1, i+1+vindauge)
			keySmal := kontekstStr(s.Runar, i-1, i) + "_" + etter1
			ordFør := ordDelFør(s.Runar, i)
			ordEtter := ordDelEtter(s.Runar, i)
			spådd := spå(tabellVid, tabellSmal, keyVid, keySmal, etter1, ordbok, ordFør, ordEtter)
			if spådd == r {
				rettTal++
				continue
			}
			avvika = append(avvika, avvik{
				Side:     s.Nummer,
				Kontekst: kontekstStr(s.Runar, maxi(i-6, 0), i) + "[" + string(r) + "]" + kontekstStr(s.Runar, i+1, i+7),
				Venta:    r,
				Fekk:     spådd,
			})
		}
	}

	sort.Slice(avvika, func(i, j int) bool { return avvika[i].Side < avvika[j].Side })

	for _, a := range avvika {
		fmt.Printf("side %3d: venta %q fekk %q  …%s…\n", a.Side, string(a.Venta), string(a.Fekk), a.Kontekst)
	}
	fmt.Printf("\n%d/%d rette (%.1f%%), %d avvik\n", rettTal, altTal, 100*float64(rettTal)/float64(altTal), len(avvika))
}

func lastSider(mappe string) ([]side, error) {
	filer, err := filepath.Glob(filepath.Join(mappe, "side-*.txt"))
	if err != nil {
		return nil, err
	}
	var sider []side
	for _, fil := range filer {
		match := sidemønster.FindStringSubmatch(fil)
		if match == nil {
			continue
		}
		nr, _ := strconv.Atoi(match[1])
		bytes, err := os.ReadFile(fil)
		if err != nil {
			continue
		}
		sider = append(sider, side{Nummer: nr, Runar: []rune(string(bytes))})
	}
	sort.Slice(sider, func(i, j int) bool { return sider[i].Nummer < sider[j].Nummer })
	return sider, nil
}

func byggTabellMed(sider []side, ekskluder, vindauge int) map[string]*teljing {
	tal := map[string]*teljing{}
	for _, s := range sider {
		if s.Nummer == ekskluder {
			continue
		}
		for i, r := range s.Runar {
			if r != 's' && r != langS {
				continue
			}
			før := kontekstStr(s.Runar, i-vindauge, i)
			etter := kontekstStr(s.Runar, i+1, i+1+vindauge)
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
	return tal
}

// minStemmer er kor mange røyster ein kontekst i det breie vindauget må ha
// før vi stolar på han, i staden for å falle attende til det smalare
// (meir robuste, men mindre spesifikke) vindauget.
const minStemmer = 3

// spå avgjer rund/lang s i denne rekkjefylgja:
//  1. Kompound-bindings-s: det som følgjer s-en i same ord er sjølv eit
//     kjent, sjølvstendig ord i Aasens eige vokabular (t.d. "Syns[s]maade"
//     -> "maade" finst i ordlista) -> rund, som ved eit vanleg ordskifte.
//  2. Det breie vindauget (t.d. 2 bokstavar før/etter), viss han har nok
//     røyster til å stolast på - fangar opp ord som "Konſonant" der eitt
//     bokstav kontekst er for upresist.
//  3. Elles det smale vindauget (1 bokstav), som er meir robust når data
//     er sparsame.
//  4. Ukjend kontekst: ordsluttande (etter tomt/mellomrom) -> rund,
//     elles -> lang (dei to mest pålitelege einskildreglane i data).
func spå(tabellVid, tabellSmal map[string]*teljing, keyVid, keySmal, etter1 string, ordbok map[string]bool, ordFør, ordEtter string) rune {
	if len(ordFør) >= 1 && len(ordEtter) >= 2 {
		// Sjekk fyrst om s-en høyrer til andre rotordet sjølv (t.d.
		// Til+Stand: "stand" er eit ord). Berre om DET ikkje stemmer,
		// men resten utan s er sitt eige ord (t.d. Syns+maade: "maade"),
		// reknar vi s-en som ein ekte bindings-s mellom to rotord.
		if ordbok["s"+ordEtter] {
			return langS
		}
		if len(ordEtter) >= 3 && ordbok[ordEtter] {
			return 's'
		}
	}
	if t := tabellVid[keyVid]; t != nil && t.Lang+t.Rund >= minStemmer && t.Lang != t.Rund {
		if t.Lang > t.Rund {
			return langS
		}
		return 's'
	}
	if t := tabellSmal[keySmal]; t != nil && t.Lang != t.Rund {
		if t.Lang > t.Rund {
			return langS
		}
		return 's'
	}
	if etter1 == "" || etter1 == " " {
		return 's'
	}
	return langS
}

// ordDelEtter hentar bokstavane i same ord etter posisjon i (ikkje inkl.
// sjølve s-en), normalisert til små bokstavar med ſ->s, for oppslag i
// ordlista.
func ordDelEtter(runar []rune, i int) string {
	var sb strings.Builder
	for j := i + 1; j < len(runar) && erBokstav(runar[j]); j++ {
		sb.WriteRune(normaliser(runar[j]))
	}
	return sb.String()
}

// ordDelFør hentar bokstavane i same ord før posisjon i (ikkje inkl.
// sjølve s-en).
func ordDelFør(runar []rune, i int) string {
	frå := i
	for frå > 0 && erBokstav(runar[frå-1]) {
		frå--
	}
	var sb strings.Builder
	for j := frå; j < i; j++ {
		sb.WriteRune(normaliser(runar[j]))
	}
	return sb.String()
}

func erBokstav(r rune) bool {
	return unicode.IsLetter(r)
}

func normaliser(r rune) rune {
	if r == langS {
		r = 's'
	}
	return unicode.ToLower(r)
}

// lastOrdbok byggjer eit sett av alle ord (2+ bokstavar) som faktisk
// finst i doc-transkripsjonen av Aasens tekst - vårt eige, kjeldeforankra
// vokabular, ikkje ei ekstern ordliste.
func lastOrdbok(sti string) (map[string]bool, error) {
	fil, err := os.Open(sti)
	if err != nil {
		return nil, err
	}
	defer fil.Close()

	ordbok := map[string]bool{}
	skannar := bufio.NewScanner(fil)
	skannar.Buffer(make([]byte, 1024*1024), 1024*1024)
	for skannar.Scan() {
		var ord strings.Builder
		for _, r := range skannar.Text() {
			if erBokstav(r) || r == '-' {
				ord.WriteRune(normaliser(r))
				continue
			}
			leggTilOrd(ordbok, ord.String())
			ord.Reset()
		}
		leggTilOrd(ordbok, ord.String())
	}
	return ordbok, skannar.Err()
}

func leggTilOrd(ordbok map[string]bool, ord string) {
	if len([]rune(ord)) >= 2 {
		ordbok[ord] = true
	}
}

func kontekstStr(runar []rune, frå, til int) string {
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

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
