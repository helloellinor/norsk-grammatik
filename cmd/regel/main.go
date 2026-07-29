// regel prøver to prinsipielle tilnærmingar til ſ/s, i staden for rein
// bokstavstatistikk, og måler dei mot den trykte boka med leave-one-out:
//
//  1. Den klassiske trykkregelen: rund s står i slutten av eit ord eller
//     eit ordledd, lang ſ overalt elles.
//  2. Boka sitt eige leksikon: slå opp korleis kvart ordform faktisk vart
//     sett i trykken, med regelen som reserveløysing for ukjende ord.
package main

import (
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

type side struct {
	Nummer int
	Runar  []rune
}

// ordform er eitt ord slik det står i teksten, med posisjonane til kvar
// s/ſ inne i ordet.
type ordform struct {
	Nykel       string // normalisert (små bokstavar, ſ->s)
	Runar       []rune // ordet slik det faktisk står
	Startindeks int    // kvar ordet byrjar i sida
}

var sidemønster = regexp.MustCompile(`side-(\d+)\.txt$`)

func main() {
	mappe := "bøker/sider"
	if len(os.Args) > 1 {
		mappe = os.Args[1]
	}

	sider, err := lastSider(mappe)
	if err != nil || len(sider) == 0 {
		fmt.Fprintln(os.Stderr, "fann ingen sider i", mappe, err)
		os.Exit(1)
	}

	evaluer(sider, "berre trykkregel", false, false)
	evaluer(sider, "leksikon + trykkregel", true, false)
	evaluer(sider, "leksikon + kontekst + trykkregel", true, true)
}

func evaluer(sider []side, namn string, brukLeksikon, brukKontekst bool) {
	rett, alt := 0, 0
	feilOrd := map[string]int{}

	for _, s := range sider {
		var leksikon map[string]string
		if brukLeksikon {
			leksikon = byggLeksikon(sider, s.Nummer)
		}
		var vid, smal map[string]*teljing
		if brukKontekst {
			vid = byggKontekst(sider, s.Nummer, 2)
			smal = byggKontekst(sider, s.Nummer, 1)
		}
		for _, ord := range finnOrd(s.Runar) {
			spådd := spåOrd(ord.Nykel, leksikon)
			kjendILeksikon := false
			if leksikon != nil {
				_, kjendILeksikon = leksikon[ord.Nykel]
			}
			faktisk := ord.Runar
			for i, r := range faktisk {
				if r != 's' && r != langS {
					continue
				}
				alt++
				gjett := rune(0)
				if i < len(spådd) {
					gjett = spådd[i]
				}
				// Når ordet ikkje står i boka frå før, er bokstav-
				// konteksten eit betre haldepunkt enn den enkle regelen.
				if brukKontekst && !kjendILeksikon {
					gjett = spåKontekst(vid, smal, s.Runar, ord.Startindeks+i)
				}
				if gjett == r {
					rett++
				} else {
					feilOrd[string(faktisk)]++
				}
			}
		}
	}

	fmt.Printf("=== %s ===\n", namn)
	fmt.Printf("%d/%d rette (%.1f%%)\n", rett, alt, 100*float64(rett)/float64(alt))
	skrivToppfeil(feilOrd, 15)
	fmt.Println()
}

// spåOrd gjev ordet med ſ/s sett inn. Med leksikon slår vi fyrst opp
// korleis ordet faktisk vart sett i resten av boka.
func spåOrd(nykel string, leksikon map[string]string) []rune {
	if leksikon != nil {
		if form, ok := leksikon[nykel]; ok {
			return []rune(form)
		}
	}
	return []rune(trykkregel(nykel))
}

// trykkregel er den klassiske regelen frå fraktursetjinga: rund s i
// slutten av ordet, lang ſ elles. (Ordledd-grensene inne i samansette
// ord fangar denne enkle forma ikkje opp - det er nettopp det leksikonet
// skal dekkje.)
func trykkregel(nykel string) string {
	runar := []rune(nykel)
	ut := make([]rune, len(runar))
	for i, r := range runar {
		if r != 's' {
			ut[i] = r
			continue
		}
		if i == len(runar)-1 {
			ut[i] = 's'
		} else {
			ut[i] = langS
		}
	}
	return string(ut)
}

// byggLeksikon lærer frå alle sidene utanom den vi testar: for kvart
// ordform, kva ſ/s-mønster som er vanlegast i trykken.
func byggLeksikon(sider []side, ekskluder int) map[string]string {
	tal := map[string]map[string]int{}
	for _, s := range sider {
		if s.Nummer == ekskluder {
			continue
		}
		for _, ord := range finnOrd(s.Runar) {
			if tal[ord.Nykel] == nil {
				tal[ord.Nykel] = map[string]int{}
			}
			tal[ord.Nykel][string(ord.Runar)]++
		}
	}

	leksikon := make(map[string]string, len(tal))
	for nykel, former := range tal {
		beste, besteTal := "", 0
		for form, n := range former {
			if n > besteTal {
				beste, besteTal = form, n
			}
		}
		leksikon[nykel] = beste
	}
	return leksikon
}

type teljing struct{ Rund, Lang int }

// byggKontekst tel opp ſ/s etter kva bokstavar som står rundt, som
// reserveløysing for ord vi ikkje har sett i trykken før.
func byggKontekst(sider []side, ekskluder, vindauge int) map[string]*teljing {
	tal := map[string]*teljing{}
	for _, s := range sider {
		if s.Nummer == ekskluder {
			continue
		}
		for i, r := range s.Runar {
			if r != 's' && r != langS {
				continue
			}
			key := kontekstNykel(s.Runar, i, vindauge)
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

const minStemmer = 3

func spåKontekst(vid, smal map[string]*teljing, runar []rune, i int) rune {
	if t := vid[kontekstNykel(runar, i, 2)]; t != nil && t.Lang+t.Rund >= minStemmer && t.Lang != t.Rund {
		if t.Lang > t.Rund {
			return langS
		}
		return 's'
	}
	if t := smal[kontekstNykel(runar, i, 1)]; t != nil && t.Lang != t.Rund {
		if t.Lang > t.Rund {
			return langS
		}
		return 's'
	}
	if i+1 >= len(runar) || !erBokstav(runar[i+1]) {
		return 's'
	}
	return langS
}

func kontekstNykel(runar []rune, i, vindauge int) string {
	return utsnitt(runar, i-vindauge, i) + "_" + utsnitt(runar, i+1, i+1+vindauge)
}

func utsnitt(runar []rune, frå, til int) string {
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

func finnOrd(runar []rune) []ordform {
	var ord []ordform
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
		bit := runar[start:i]
		ord = append(ord, ordform{
			Nykel:       normaliserOrd(bit),
			Runar:       bit,
			Startindeks: start,
		})
	}
	return ord
}

func normaliserOrd(runar []rune) string {
	var sb strings.Builder
	for _, r := range runar {
		if r == langS {
			r = 's'
		}
		sb.WriteRune(unicode.ToLower(r))
	}
	return sb.String()
}

func erBokstav(r rune) bool { return unicode.IsLetter(r) }

func skrivToppfeil(feilOrd map[string]int, n int) {
	type rad struct {
		Ord string
		Tal int
	}
	var rader []rad
	for o, t := range feilOrd {
		rader = append(rader, rad{o, t})
	}
	sort.Slice(rader, func(i, j int) bool { return rader[i].Tal > rader[j].Tal })
	if len(rader) > n {
		rader = rader[:n]
	}
	fmt.Print("  vanlegaste feil: ")
	for _, r := range rader {
		fmt.Printf("%s(%d) ", r.Ord, r.Tal)
	}
	fmt.Println()
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
		b, err := os.ReadFile(fil)
		if err != nil {
			continue
		}
		sider = append(sider, side{Nummer: nr, Runar: []rune(string(b))})
	}
	sort.Slice(sider, func(i, j int) bool { return sider[i].Nummer < sider[j].Nummer })
	return sider, nil
}
