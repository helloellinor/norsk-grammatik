package main

import (
	"regexp"
	"strings"
)

// Bolkslag er dei blokkene teksten er bygd av. Alle saman er henta frå
// boka sitt eige oppsett, ikkje frå formateringa i doc-fila: Afdeling ->
// romartal -> bokstav er hierarkiet i innhaldslista, §-nummera er dei
// einingane teksten siterer seg sjølv med, og Anm. er Aasens eige merke
// for ein merknad.
type Bolkslag string

const (
	Afdeling     Bolkslag = "afdeling"
	Seksjon      Bolkslag = "seksjon"      // I. II. III.
	Underseksjon Bolkslag = "underseksjon" // a) b) c)
	Mellomtittel Bolkslag = "mellomtittel" // Enkelte Vokaler.
	Paragraf     Bolkslag = "paragraf"     // 8. Det norſke Sprog ...
	Merknad      Bolkslag = "merknad"      // Anm. ...
	Avsnitt      Bolkslag = "avsnitt"      // innrykt framhald
	Oppslag      Bolkslag = "oppslag"      // Ang. — Angelſachſiſk.
	Brødtekst    Bolkslag = "brødtekst"
)

type Bolk struct {
	Slag   Bolkslag
	Nummer string // §-nummer, romartal eller bokstav
	Tittel string
	Tekst  string
}

var (
	reAfdeling  = regexp.MustCompile(`^(Førſte|Første|Anden|Tredie|Fjerde|Femte) Afdeling\.?$`)
	reSeksjon   = regexp.MustCompile(`^([IVX]+)\.\s+(.+?)\.?$`)
	reUnderseks = regexp.MustCompile(`^([a-z])\)\s+(.+?)\.?$`)
	reParagraf  = regexp.MustCompile(`^(\d+)\.?\s+(.+)$`)
	reMerknad   = regexp.MustCompile(`^\s*Anm(?:\.|ærkning\.)\s*`)
	reOppslag   = regexp.MustCompile(`^(\S[^—]{0,24}?)\s+—\s+(.+)$`)
	reSøppel    = regexp.MustCompile(`GOTOBUTTON|PAGEREF|_Toc\d+|^\s*TOC\s`)
)

// Tolk deler teksten opp i blokker. Linjer som held fram ei setning frå
// linja over blir sette saman att: doc-fila bryt avsnitt der 1864-utgåva
// hadde sideskift, og det er ei side i papiret, ikkje eit avsnittsskifte
// i teksten.
func Tolk(tekst string) []Bolk {
	var bolkar []Bolk
	linjer := strings.Split(tekst, "\n")

	leggTil := func(b Bolk) { bolkar = append(bolkar, b) }

	for _, rå := range linjer {
		if reSøppel.MatchString(rå) {
			continue
		}
		linje := strings.TrimRight(rå, " \t\r")
		if strings.TrimSpace(linje) == "" {
			continue
		}

		innrykt := strings.HasPrefix(linje, "\t") || strings.HasPrefix(linje, "    ")
		trimma := strings.TrimSpace(linje)

		switch {
		case reMerknad.MatchString(trimma):
			leggTil(Bolk{Slag: Merknad, Tekst: trimma})

		case !innrykt && reAfdeling.MatchString(trimma):
			leggTil(Bolk{Slag: Afdeling, Tittel: strings.TrimSuffix(trimma, ".")})

		case !innrykt && erKort(trimma) && reSeksjon.MatchString(trimma):
			m := reSeksjon.FindStringSubmatch(trimma)
			leggTil(Bolk{Slag: Seksjon, Nummer: m[1], Tittel: m[2]})

		case !innrykt && erKort(trimma) && reUnderseks.MatchString(trimma):
			m := reUnderseks.FindStringSubmatch(trimma)
			leggTil(Bolk{Slag: Underseksjon, Nummer: m[1], Tittel: m[2]})

		case !innrykt && reParagraf.MatchString(trimma) && !erKort(trimma):
			m := reParagraf.FindStringSubmatch(trimma)
			leggTil(Bolk{Slag: Paragraf, Nummer: m[1], Tekst: m[2]})

		case !innrykt && erKort(trimma) && reOppslag.MatchString(trimma):
			// Forkortingslista: stikkord — forklaring.
			m := reOppslag.FindStringSubmatch(trimma)
			leggTil(Bolk{Slag: Oppslag, Nummer: m[1], Tekst: m[2]})

		case !innrykt && erKort(trimma):
			// Ei kort linje åleine mellom avsnitt er ei mellomoverskrift
			// (Enkelte Vokaler., Tvelyd (Diftonger)., Overſigt.).
			leggTil(Bolk{Slag: Mellomtittel, Tittel: strings.TrimSuffix(trimma, ".")})

		case innrykt:
			leggTil(Bolk{Slag: Avsnitt, Tekst: trimma})

		default:
			// Held denne linja fram setninga frå blokka over, er det eit
			// sideskift i papiret og ikkje eit nytt avsnitt.
			if n := len(bolkar); n > 0 && heldFram(bolkar[n-1].Tekst) {
				bolkar[n-1].Tekst += " " + trimma
				continue
			}
			leggTil(Bolk{Slag: Brødtekst, Tekst: trimma})
		}
	}

	return bolkar
}

// heldFram seier om ein tekst er avbroten midt i ei setning.
func heldFram(tekst string) bool {
	t := strings.TrimSpace(tekst)
	if t == "" {
		return false
	}
	siste := t[len(t)-1:]
	return !strings.Contains(".:;!?»)—", siste)
}

func erKort(s string) bool { return len([]rune(s)) < 60 }
