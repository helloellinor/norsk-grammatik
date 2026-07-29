package tekst

import (
	"fmt"
	"strconv"
	"strings"
)

// Peikar er ei overskrift ein kan hoppe til inne i ein del.
type Peikar struct {
	Ankar  string
	Tittel string
	Nivå   int // 2 = seksjon, 3 = underseksjon, 4 = mellomtittel
}

// Del er ein overbolk av verket - ei Afdeling, eit Tillæg eller ein av
// bolkane framanfor - med alt som høyrer under. Kvar del er eitt ark.
// Seksjonane inne i han er ikkje eigne ark, men hoppmål i registeret.
type Del struct {
	Id      int
	Nivå    int // 1 = Afdeling, 2 = seksjon
	Tittel  string
	Titlar  []Peikar // overskrifter inne i delen
	Paragr  []string // §-nummera som står i delen
	Blokker []Blokk
}

// DelOpp grupperer blokkene i lesbare delar. Ein ny del byrjar ved kvar
// Afdeling og kvar romartal-seksjon; alt anna fell inn under den delen
// som står føre. Kvar overskrift inne i delen får eit ankerpunkt, so ho
// kan naaast frå registeret.
func DelOpp(blokker []Blokk) []Del {
	var delar []Del

	ny := func(nivå int, tittel string) {
		delar = append(delar, Del{
			Id:     len(delar),
			Nivå:   nivå,
			Tittel: tittel,
		})
	}

	leggTil := func(b Blokk) {
		if len(delar) == 0 {
			ny(1, "Tittelblad")
		}
		d := &delar[len(delar)-1]

		if b.Slag == Paragraf {
			d.Paragr = append(d.Paragr, b.Nummer)
		}
		d.Blokker = append(d.Blokker, b)
	}

	for i := 0; i < len(blokker); i++ {
		b := blokker[i]
		switch b.Slag {
		case Afdeling:
			// Afdeling-overskrifta står på éi linje og namnet på neste
			// ("Førſte Afdeling." / "Lydlære."), so vi slår dei saman
			// til éin tittel og syner namnet som undertittel i staden
			// for som ei laus mellomoverskrift.
			// Berre Afdeling og Tillæg har namnet sitt på linja under
			// ("Førſte Afdeling." / "Lydlære."). Fortale, Indledning og
			// dei andre framanfor-bolkane står med namnet sitt åleine.
			if harNamnelinje(b.Tittel) && i+1 < len(blokker) && blokker[i+1].Slag == Mellomtittel {
				b.Undertittel = blokker[i+1].Tittel
				i++
			}
			tittel := b.Tittel
			if b.Undertittel != "" {
				tittel = fmt.Sprintf("%s · %s", b.Tittel, b.Undertittel)
			}
			ny(1, tittel)
			leggTil(b)
		default:
			leggTil(b)
		}
	}

	for i := range delar {
		delar[i].Titlar = titlarMedInnhald(&delar[i])
	}
	return utanTomme(delar)
}

// titlarMedInnhald plukkar ut dei overskriftene ein kan hoppe til: dei
// som faktisk har tekst under seg. Linjene paa tittelbladet - forfattar,
// verk, stad - er sette som overskrifter i Word-fila, men det staar
// ingenting under dei, so dei høyrer ikkje heime i navigasjonen.
func titlarMedInnhald(d *Del) []Peikar {
	var ut []Peikar
	for i, b := range d.Blokker {
		nivå := 0
		switch b.Slag {
		case Seksjon:
			nivå = 2
		case Underseksjon:
			nivå = 3
		case Mellomtittel:
			nivå = 4
		default:
			continue
		}
		// Kvar overskrift faar eit ankar, uansett om ho hamnar i
		// registeret - elles kjem ho ut med id="" og kan ikkje naaast
		// med ei lenke i det heile.
		ankar := fmt.Sprintf("t%d-%d", d.Id, i)
		d.Blokker[i].Ankar = ankar

		// I registeret høyrer berre dei nummererte heime - dei
		// unummererte er overskrifter i teksten, ikkje punkt ein blar
		// etter - og berre dei som faktisk har tekst under seg.
		if b.Nummer == "" || (nivå > 2 && !harTekstUnder(d.Blokker[i+1:])) {
			continue
		}
		tittel := b.Tittel
		if b.Nummer != "" {
			skil := ") "
			if nivå == 2 {
				skil = ". "
			}
			tittel = b.Nummer + skil + b.Tittel
		}
		ut = append(ut, Peikar{Ankar: ankar, Tittel: tittel, Nivå: nivå})
	}
	return ut
}

// harTekstUnder ser om det kjem tekst før neste overskrift.
func harTekstUnder(etter []Blokk) bool {
	for _, b := range etter {
		switch b.Slag {
		case Underseksjon, Mellomtittel, Seksjon, Afdeling:
			return false
		case Brødtekst, Paragraf, Merknad, Avsnitt, Oppsett, Oppslag:
			return true
		}
	}
	return false
}

// utanTomme fjernar bolkar som korkje har tekst eller underbolkar. Boka
// si eiga innhaldsliste er ein slik: sjølve listeoppføringane er
// Word-felt som ikkje ber tekst, og navigasjonen vaar gjer same nytta.
// Ei Afdeling har derimot ofte berre tittelen sin, av di teksten ligg i
// seksjonane under henne - henne skal vi halde paa.
func utanTomme(delar []Del) []Del {
	var ut []Del
	for i, d := range delar {
		harBorn := i+1 < len(delar) && delar[i+1].Nivå > d.Nivå
		// Tittelbladet staar sjølv om det ikkje har brødtekst: sida er
		// sett for seg. Elles fell bolkar utan tekst og utan underbolkar
		// bort - som den fyrste «Føreord (1997)», som berre er ei
		// oppføring i 1997-utgåva si eiga vesle innhaldsliste framme.
		if i > 0 && !d.HarTekst() && !harBorn {
			continue
		}
		d.Id = len(ut)
		for i := range d.Titlar {
			d.Titlar[i].Ankar = fmt.Sprintf("t%d-%d", d.Id, i)
		}
		for i := range d.Blokker {
			if d.Blokker[i].Ankar != "" {
				d.Blokker[i].Ankar = fmt.Sprintf("t%d-%d", d.Id, ankarNr(d.Blokker[i].Ankar))
			}
		}
		ut = append(ut, d)
	}
	return ut
}

// ankarNr hentar løpenummeret ut av eit ankar på forma "t<del>-<nr>".
func ankarNr(ankar string) int {
	if i := strings.LastIndex(ankar, "-"); i >= 0 {
		n, err := strconv.Atoi(ankar[i+1:])
		if err == nil {
			return n
		}
	}
	return 0
}

// HarTekst seier om bolken har anna enn overskrifter. Ei Afdeling har
// ofte berre namnet sitt, av di teksten ligg i seksjonane under henne.
func (d Del) HarTekst() bool {
	for _, b := range d.Blokker {
		switch b.Slag {
		case Afdeling, Seksjon, Underseksjon, Mellomtittel:
		default:
			return true
		}
	}
	return false
}

// ParagrafIndeks seier kva del kvart §-nummer står i, so ein kan hoppe
// rett til ein paragraf utan å vite kva bolk han høyrer til.
func ParagrafIndeks(delar []Del) map[string]int {
	ut := map[string]int{}
	for _, d := range delar {
		for _, nr := range d.Paragr {
			if _, finst := ut[nr]; !finst {
				ut[nr] = d.Id
			}
		}
	}
	return ut
}

func harNamnelinje(tittel string) bool {
	return reAfdeling.MatchString(tittel) || reAfdeling.MatchString(tittel+".")
}
