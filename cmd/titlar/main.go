// titlar lagar ei redigerbar liste over overskriftene i verket.
//
// Grunnlaget er boka si eiga innhaldsliste. Ho ligg i Word-fila som
// felt (GOTOBUTTON/PAGEREF), og wvHtml slepper dei - difor hentar vi
// nettopp denne biten med textutil, som gjengjev dei som tekst.
//
// Fila er meint til å rettast for hand. Koden merkjer berre kva han
// sjølv har funne, so det er lett å sjå kvar dei to er usamde.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"aasen/internal/tekst"
)

type oppf struct {
	Tittel string
	Side   string
	Nivå   int
}

var (
	reLinje     = regexp.MustCompile(`^(?:.*?\\P\s+"\s*"\s*)?(.+?)\s+GOTOBUTTON.*?PAGEREF\s+\S+\s+(\S+)\s*$`)
	reRoman     = regexp.MustCompile(`^[IVX]+\.\s`)
	reTal       = regexp.MustCompile(`^\d+\.\s`)
	reBokst     = regexp.MustCompile(`^[a-z][.)]\s`)
	reOver      = regexp.MustCompile(`(Afdeling|Tillæg)\.`)
	reFramanfor = regexp.MustCompile(`^(Fortale|Indledning|Indholdsliste|Forklaring af nogle Forkortninger|Enkelte Tillæg og Rettelser|Register)\.?$`)
)

func main() {
	doc := flag.String("doc", "bøker/Norsk Grammatik.doc", "Word-fila")
	sats := flag.String("sats", "bøker/grammatik-sats.json", "blokkene med ſ")
	ut := flag.String("ut", "bøker/overskrifter.md", "lista som skal rettast")
	flag.Parse()

	liste, err := lesInnhaldsliste(*doc)
	if err != nil {
		stopp(err)
	}
	funne, err := lesFunne(*sats)
	if err != nil {
		stopp(err)
	}

	skriv(*ut, liste, funne)
}

// lesInnhaldsliste hentar boka si eiga liste ut av Word-fila.
func lesInnhaldsliste(doc string) ([]oppf, error) {
	cmd := exec.Command("textutil", "-convert", "txt", "-stdout", doc)
	rå, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("textutil: %w", err)
	}

	var ut []oppf
	for _, linje := range strings.Split(string(rå), "\n") {
		if !strings.Contains(linje, "PAGEREF") {
			continue
		}
		m := reLinje.FindStringSubmatch(linje)
		if m == nil {
			continue
		}
		t := strings.TrimSpace(m[1])
		ut = append(ut, oppf{Tittel: t, Side: m[2], Nivå: nivå(t)})
	}
	return ut, nil
}

// nivå les hierarkiet ut av sjølve nummereringa, slik lista syner han.
func nivå(t string) int {
	switch {
	case reOver.MatchString(t):
		return 1
	case reRoman.MatchString(t), reTal.MatchString(t):
		return 2
	case reBokst.MatchString(t):
		return 3
	case reFramanfor.MatchString(t):
		return 1
	default:
		return 4 // ei overskrift utan nummer
	}
}

// lesFunne gjev overskriftene koden sjølv har kome fram til.
func lesFunne(sti string) (map[string]string, error) {
	rå, err := os.ReadFile(sti)
	if err != nil {
		return nil, err
	}
	var blokker []tekst.Blokk
	if err := json.Unmarshal(rå, &blokker); err != nil {
		return nil, err
	}

	// Koden lagrar nummer og namn i kvar sine felt, og Afdeling-namnet
	// på ei eiga linje. Vi må setje dei saman att for å kunne samanlikne
	// med boka si liste, som har heile tittelen på éi linje.
	ut := map[string]string{}
	klassa := tekst.Klassifiser(blokker)
	for i, b := range klassa {
		var heil string
		switch b.Slag {
		case tekst.Afdeling:
			heil = b.Tittel
			if i+1 < len(klassa) && klassa[i+1].Slag == tekst.Mellomtittel {
				heil += ". " + klassa[i+1].Tittel
			}
		case tekst.Seksjon, tekst.Underseksjon:
			heil = b.Nummer + ". " + b.Tittel
		case tekst.Mellomtittel:
			heil = b.Tittel
		default:
			continue
		}
		ut[nykel(heil)] = string(b.Slag)
	}
	return ut, nil
}

// nykel gjer titlar samanliknbare på tvers av lang ſ og småskilnader.
func nykel(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer("ſ", "s", ".", "", ",", "", ")", "", "`", "", "'", "", "’", "").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func skriv(sti string, liste []oppf, funne map[string]string) {
	fil, err := os.Create(sti)
	if err != nil {
		stopp(err)
	}
	defer fil.Close()
	w := bufio.NewWriter(fil)
	defer w.Flush()

	fmt.Fprint(w, `# Overskrifter i Norſk Grammatik

Denne lista er boka si eiga innhaldsliste, henta ut av Word-fila. Ho er
fasiten for kva som er ei overskrift og kva nivå ho ligg på.

Rett fritt. Talet framfor er nivået:

    1  Afdeling, Tillæg, og bolkane framanfor
    2  I. II. III. og 1. 2. 3.
    3  a) b) c)
    4  overskrifter utan nummer

Merket bak seier kva koden min gjorde med same tittelen:

    =  same nivå
    ~  fann henne, men på eit anna nivå
    !  fann henne ikkje - ho blir ikkje synt som overskrift
    +  koden har ei overskrift som ikkje står i boka si liste

`)

	tilNivå := map[string]int{"afdeling": 1, "seksjon": 2, "underseksjon": 3, "mellomtittel": 4}
	sett := map[string]bool{}
	var manglar, ekstra int

	fmt.Fprintln(w, "## Boka si liste")
	fmt.Fprintln(w)
	for _, o := range liste {
		n := nykel(o.Tittel)
		sett[n] = true
		merke := "!"
		if slag, ok := funne[n]; ok {
			if tilNivå[slag] == o.Nivå {
				merke = "="
			} else {
				merke = fmt.Sprintf("~ %s", slag)
			}
		} else {
			manglar++
		}
		fmt.Fprintf(w, "%d  %-52s  %-8s %s\n", o.Nivå, o.Tittel, "s. "+o.Side, merke)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Overskrifter koden har funne som ikkje står i lista")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Boka listar ikkje alt - somme mellomoverskrifter står berre i")
	fmt.Fprintln(w, "teksten. Stryk dei som ikkje er overskrifter.")
	fmt.Fprintln(w)
	for n, slag := range funne {
		if sett[n] {
			continue
		}
		ekstra++
		fmt.Fprintf(w, "%d  %-52s  %-8s +\n", tilNivå[slag], n, "")
	}

	fmt.Printf("%d oppføringar i boka si liste, %d fann koden ikkje, %d ekstra -> %s\n",
		len(liste), manglar, ekstra, sti)
}

func stopp(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
