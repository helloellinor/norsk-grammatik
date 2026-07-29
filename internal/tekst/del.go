package tekst

import "fmt"

// Del er ein lesbar bolk av verket - ei Afdeling eller ein romartal-
// seksjon med alt som høyrer under, fram til neste av same slag. Det er
// denne eininga lesaren blar i, og som innhaldslista peikar på.
type Del struct {
	Id     int
	Nivå   int // 1 = Afdeling, 2 = seksjon
	Tittel string
	Ankar  string // fyrste §-nummeret i delen, om det finst
	Bolkar []Bolk
}

// DelOpp grupperer blokkene i lesbare delar. Ein ny del byrjar ved kvar
// Afdeling og kvar romartal-seksjon; alt anna fell inn under den delen
// som står føre.
func DelOpp(bolkar []Bolk) []Del {
	var delar []Del

	ny := func(nivå int, tittel string) {
		delar = append(delar, Del{
			Id:     len(delar),
			Nivå:   nivå,
			Tittel: tittel,
		})
	}

	leggTil := func(b Bolk) {
		if len(delar) == 0 {
			ny(1, "Framan")
		}
		d := &delar[len(delar)-1]
		d.Bolkar = append(d.Bolkar, b)
		if d.Ankar == "" && b.Slag == Paragraf {
			d.Ankar = b.Nummer
		}
	}

	for i := 0; i < len(bolkar); i++ {
		b := bolkar[i]
		switch b.Slag {
		case Afdeling:
			// Afdeling-overskrifta står på éi linje og namnet på neste
			// ("Førſte Afdeling." / "Lydlære."), so vi slår dei saman
			// til éin tittel og syner namnet som undertittel i staden
			// for som ei laus mellomoverskrift.
			b.Tittel = b.Tittel
			if i+1 < len(bolkar) && bolkar[i+1].Slag == Mellomtittel {
				b.Undertittel = bolkar[i+1].Tittel
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
