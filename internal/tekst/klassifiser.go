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

	// Boka blandar talsystem i seksjonsnummereringa, men rekkjefylgja
	// held fram tvers gjennom. Fyrste Tillæg har «I. Den nordenfjeldſke
	// Række», so «2. Den veſtenfjeldſke» og «3. Den søndenfjeldſke» - og
	// alle tre staar jamsides i hennar eiga innhaldsliste. Difor er ei
	// kort, arabisk nummerert overskrift ein seksjon nettopp naar talet
	// held fram rekkja, og elles ei underoverskrift.
	sisteSeksjon := 0
	// Oppslagsmønsteret («Ang. — Angelſachſiſk.») passar òg paa korte
	// linjer med tankestrek i brødteksten - t.d. avlydsrekkjene. Det skal
	// difor berre gjelde inne i forkortingsbolken.
	iForkortingar := false

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
			sisteSeksjon = 0
			b.Slag, b.Tittel = Afdeling, strings.TrimSuffix(t, ".")

		// Òg desse er halvfeite i boka. Uthevinga skil den ekte
		// overskrifta «Føreord (1997)» frå den same nemninga i
		// 1997-utgåva si eiga vesle innhaldsliste framme, som staar
		// heilt umarkert.
		case reFramanfor.MatchString(t) && heiltUtheva(b):
			sisteSeksjon = 0
			iForkortingar = strings.HasPrefix(t, "Forklaring af nogle Forkortninger")
			b.Slag, b.Tittel = Afdeling, strings.TrimSuffix(t, ".")
			// Same bolknamnet står stundom to gonger etter kvarandre,
			// av di boka har det både som overskrift og som kolumnetittel.
			if n := len(ut); n > 0 && ut[n-1].Slag == Afdeling && ut[n-1].Tittel == b.Tittel {
				continue
			}

		case erKort(t) && reSeksjon.MatchString(t):
			m := reSeksjon.FindStringSubmatch(t)
			sisteSeksjon = romartal(m[1])
			b.Slag, b.Nummer, b.Tittel = Seksjon, m[1], m[2]

		case erKort(t) && erTalOverskrift(t):
			m := reTalseks.FindStringSubmatch(t)
			nr, _ := strconv.Atoi(m[1])
			if nr == sisteSeksjon+1 {
				sisteSeksjon = nr
				b.Slag, b.Nummer, b.Tittel = Seksjon, m[1], m[2]
			} else {
				b.Slag, b.Nummer, b.Tittel = Underseksjon, m[1], m[2]
			}

		case erKort(t) && reUnderseks.MatchString(t):
			m := reUnderseks.FindStringSubmatch(t)
			b.Slag, b.Nummer, b.Tittel = Underseksjon, m[1], m[2]

		// Oppstilte døme. Kvart døme staar som si eiga blokk i kjelda og
		// er ei line for seg i trykket; utan eige slag rann dei saman til
		// vanlege, utlikna avsnitt med innrykk.
		case reDøme.MatchString(t):
			m := reDøme.FindStringSubmatch(t)
			b.Slag, b.Nummer = Døme, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, tal(t)-tal(m[2]))

		// Alt anna som byrjar med eit tal er eit §-avsnitt, same kor
		// kort det er: sideskifta i papiret deler somme paragrafar i to,
		// so fyrste luten kan vera berre ei halv linje (t.d. § 153).
		case reParagraf.MatchString(t):
			m := reParagraf.FindStringSubmatch(t)
			b.Slag, b.Nummer = Paragraf, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, tal(t)-tal(m[2]))

		case iForkortingar && erKort(t) && reOppslag.MatchString(t):
			m := reOppslag.FindStringSubmatch(t)
			b.Slag, b.Nummer = Oppslag, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, tal(t)-tal(m[2]))

		// Held blokka føre paa ei uavslutta setning, er dette framhaldet
		// hennar - same kor kort det er. Utan denne prøven vart smaabitar
		// som «o. ſ. v.» tekne for overskrifter, av di dei er korte, og
		// fall ut av merknaden dei høyrer til.
		case heldFramFrå(ut):
			b.Slag = Brødtekst
			n := len(ut) - 1
			ut[n].Stumpar = append(ut[n].Stumpar, Stump{Tekst: " "})
			ut[n].Stumpar = append(ut[n].Stumpar, b.Stumpar...)
			continue

		// Boka uthevar overskriftene: Afdeling-namna er halvfeite,
		// mellomoverskriftene kursive. Det er signalet - ikkje lengda.
		// Utan denne prøven vart kvar kort line ei overskrift, og lister
		// som «ei til e: Leir, heil, rein» hamna i registeret.
		case erKort(t) && !erInnleiing(t) && heiltUtheva(b):
			b.Slag, b.Tittel = Mellomtittel, strings.TrimSuffix(t, ".")

		default:
			// Held blokka fram setninga frå blokka over, er det eit
			// sideskift i papiret og ikkje eit nytt avsnitt.
			b.Slag = Brødtekst
		}

		ut = append(ut, b)
	}

	return ut
}

// heiltUtheva seier om heile blokka staar utheva - kursiv eller
// halvfeit. Overskriftene i boka gjer det; brødtekst og lister ikkje.
func heiltUtheva(b Blokk) bool {
	kursiv, feit, sett := true, true, false
	for _, s := range b.Stumpar {
		if strings.TrimSpace(s.Tekst) == "" {
			continue
		}
		sett = true
		if !s.Kursiv {
			kursiv = false
		}
		if !s.Feit {
			feit = false
		}
	}
	return sett && (kursiv || feit)
}

// heldFramFrå seier om siste blokka er avbroten midt i ei setning, so
// det som kjem no er framhaldet hennar. Sideskifta i papiret deler baade
// avsnitt og merknader i to.
func heldFramFrå(ut []Blokk) bool {
	n := len(ut) - 1
	return n >= 0 && kanHaldeFram(ut[n].Slag) && heldFram(ut[n].Tekst())
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

func erKort(s string) bool { return tal(s) < 60 }

// romartal gjer eit seksjonsnummer om til eit tal, so rekkja kan fylgjast
// tvers gjennom talsystema.
func romartal(s string) int {
	verd := map[byte]int{'I': 1, 'V': 5, 'X': 10}
	sum := 0
	for i := 0; i < len(s); i++ {
		v := verd[s[i]]
		if i+1 < len(s) && v < verd[s[i+1]] {
			sum -= v
		} else {
			sum += v
		}
	}
	return sum
}

// erInnleiing skil ei kort innleiingssetning frå ei overskrift. Framfor
// oppsetta staar det ofte ei line som «De Former ſom forefindes i
// Bygdemaalene, ere:» - ho er kort nok til aa sjaa ut som ei overskrift,
// men ei overskrift er ei nemning og endar ikkje med kolon eller komma.
func erInnleiing(s string) bool {
	t := strings.TrimSpace(s)
	return strings.HasSuffix(t, ":") || strings.HasSuffix(t, ",") || strings.HasSuffix(t, ";")
}

// tal er talet paa teikn, ikkje byte. skjerFramme tel teikn, so gjev ein
// han eit bytetal i staden, skjer han for mykje: tankestreken er eitt
// teikn men tre byte, og daa forsvann dei to fyrste bokstavane i kvart
// oppslag i forkortingslista.
func tal(s string) int { return len([]rune(s)) }

// heldFram seier om ein tekst er avbroten midt i ei setning.
func heldFram(tekst string) bool {
	t := strings.TrimSpace(tekst)
	if t == "" {
		return false
	}
	// Siste RUNE, ikkje siste byte: » og — er flerbyte, og med byte-snitt
	// kunne dei aldri matche - same feilen som tal() ovanfor rettar.
	r := []rune(t)
	return !strings.ContainsRune(".:;!?»)—", r[len(r)-1])
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
