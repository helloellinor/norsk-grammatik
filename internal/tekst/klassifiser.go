package tekst

import (
	"strconv"
	"strings"
)

// Klassifiser gjev kvar blokk sitt slag ut frå boka sitt eige oppsett.
// Reglane er dei same som før, men no arbeider dei på blokkene frå
// Word-fila i staden for på linjer i ei tekstfil.
func Klassifiser(inn []Blokk) []Blokk {
	var ut []Blokk

	// Boka nummererer seksjonane med romartal, men på eitt punkt brukar
	// ho arabartal der «I.» skulle ha stått: «1. Subſtantivernes Bøining»
	// står i hennar eiga innhaldsliste jamsides «II. Adjektivernes
	// Bøining». Difor reknar vi ei kort, arabisk nummerert overskrift som
	// ein seksjon når Afdelinga enno ikkje har fått nokon.
	harSeksjon := false

	for _, b := range inn {
		if b.Tabell != nil {
			// Eit oppsett som står rett etter ein merknad høyrer til
			// han, og skal setjast som ein del av merknaden.
			b.Slag = Oppsett
			if n := len(ut); n > 0 && (ut[n-1].Slag == Merknad || (ut[n-1].Slag == Oppsett && ut[n-1].IMerknad)) {
				b.IMerknad = true
			}
			ut = append(ut, b)
			continue
		}

		rå := b.Tekst()
		if reSøppel.MatchString(rå) {
			continue
		}
		t := strings.TrimSpace(rå)
		if t == "" {
			continue
		}

		switch {
		case reMerknad.MatchString(t):
			b.Slag = Merknad

		case reAfdeling.MatchString(t):
			harSeksjon = false
			b.Slag, b.Tittel = Afdeling, strings.TrimSuffix(t, ".")

		case erKort(t) && reSeksjon.MatchString(t):
			m := reSeksjon.FindStringSubmatch(t)
			harSeksjon = true
			b.Slag, b.Nummer, b.Tittel = Seksjon, m[1], m[2]

		case erKort(t) && erTalOverskrift(t):
			m := reTalseks.FindStringSubmatch(t)
			if !harSeksjon {
				harSeksjon = true
				b.Slag, b.Nummer, b.Tittel = Seksjon, m[1], m[2]
			} else {
				b.Slag, b.Nummer, b.Tittel = Underseksjon, m[1], m[2]
			}

		case erKort(t) && reUnderseks.MatchString(t):
			m := reUnderseks.FindStringSubmatch(t)
			b.Slag, b.Nummer, b.Tittel = Underseksjon, m[1], m[2]

		// Alt anna som byrjar med eit tal er eit §-avsnitt, same kor
		// kort det er: sideskifta i papiret deler somme paragrafar i to,
		// so fyrste luten kan vera berre ei halv linje (t.d. § 153).
		case reParagraf.MatchString(t):
			m := reParagraf.FindStringSubmatch(t)
			b.Slag, b.Nummer = Paragraf, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, len(t)-len(m[2]))

		case erKort(t) && reOppslag.MatchString(t):
			m := reOppslag.FindStringSubmatch(t)
			b.Slag, b.Nummer = Oppslag, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, len(t)-len(m[2]))

		case erKort(t):
			b.Slag, b.Tittel = Mellomtittel, strings.TrimSuffix(t, ".")

		default:
			// Held blokka fram setninga frå blokka over, er det eit
			// sideskift i papiret og ikkje eit nytt avsnitt.
			if n := len(ut); n > 0 && kanHaldeFram(ut[n-1].Slag) && heldFram(ut[n-1].Tekst()) {
				ut[n-1].Stumpar = append(ut[n-1].Stumpar, Stump{Tekst: " "})
				ut[n-1].Stumpar = append(ut[n-1].Stumpar, b.Stumpar...)
				continue
			}
			b.Slag = Brødtekst
		}

		ut = append(ut, b)
	}

	return ut
}

func kanHaldeFram(s Bolkslag) bool {
	return s == Brødtekst || s == Paragraf || s == Merknad
}

// skjerFramme fjernar dei n fyrste teikna, som ved eit §-nummer vi alt
// har teke vare på i eit eige felt, utan å øydeleggje utheivinga i det
// som står att.
func skjerFramme(st []Stump, n int) []Stump {
	var ut []Stump
	for _, s := range st {
		if n <= 0 {
			ut = append(ut, s)
			continue
		}
		r := []rune(s.Tekst)
		if len(r) <= n {
			n -= len(r)
			continue
		}
		s.Tekst = string(r[n:])
		n = 0
		ut = append(ut, s)
	}
	return ut
}

func erKort(s string) bool { return len([]rune(s)) < 60 }

// heldFram seier om ein tekst er avbroten midt i ei setning.
func heldFram(tekst string) bool {
	t := strings.TrimSpace(tekst)
	if t == "" {
		return false
	}
	return !strings.ContainsAny(t[len(t)-1:], ".:;!?»)—")
}

// erTalOverskrift skil ei arabisk nummerert overskrift frå eit kort
// §-avsnitt. Ei overskrift er ei avslutta nemning og endar med punktum,
// medan eit §-avsnitt held fram i teksten. Seksjonsnummera er dessutan
// låge, medan §-nummera same staden i boka er godt over hundre.
func erTalOverskrift(s string) bool {
	m := reTalseks.FindStringSubmatch(s)
	if m == nil || !strings.HasSuffix(s, ".") {
		return false
	}
	nr, err := strconv.Atoi(m[1])
	return err == nil && nr <= 12
}
