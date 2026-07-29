package tekst

import "fmt"

// Peikar er ei overskrift ein kan hoppe til inne i ein del.
type Peikar struct {
	Ankar  string
	Tittel string
	Nivå   int // 3 = underseksjon, 4 = mellomtittel
}

// Del er ein lesbar bolk av verket - ei Afdeling eller ein romartal-
// seksjon med alt som høyrer under, fram til neste av same slag. Det er
// denne eininga lesaren blar i, og som innhaldslista peikar på.
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
			ny(1, "Framanfor")
		}
		d := &delar[len(delar)-1]

		switch b.Slag {
		case Underseksjon, Mellomtittel:
			b.Ankar = fmt.Sprintf("t%d-%d", d.Id, len(d.Titlar))
			nivå := 4
			if b.Slag == Underseksjon {
				nivå = 3
			}
			tittel := b.Tittel
			if b.Nummer != "" {
				tittel = b.Nummer + ") " + b.Tittel
			}
			d.Titlar = append(d.Titlar, Peikar{Ankar: b.Ankar, Tittel: tittel, Nivå: nivå})
		case Paragraf:
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
			if i+1 < len(blokker) && blokker[i+1].Slag == Mellomtittel {
				b.Undertittel = blokker[i+1].Tittel
				i++
			}
			tittel := b.Tittel
			if b.Undertittel != "" {
				tittel = fmt.Sprintf("%s · %s", b.Tittel, b.Undertittel)
			}
			ny(1, tittel)
			leggTil(b)
		case Seksjon:
			ny(2, fmt.Sprintf("%s. %s", b.Nummer, b.Tittel))
			leggTil(b)
		default:
			leggTil(b)
		}
	}

	return delar
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
