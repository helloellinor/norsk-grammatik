package tekst

import (
	"strconv"
	"strings"
	"unicode"
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
			// Eit sideskift deler ògso tavler. Fylgjer ei tavle rett etter
			// ei anna med like mange kolonnar, er det same tavla: dei seks
			// klasserekkjene i § 207 kom ut som «1) 2)» i den eine og
			// «3) 4) 5) 6)» i den andre, med overskriftsrada berre paa den
			// fyrste. Fem slike par finst i verket, alle ekte.
			if n := len(ut); n > 0 && ut[n-1].Slag == Oppsett &&
				ut[n-1].Tabell != nil && b.Tabell != nil &&
				breiddPå(ut[n-1].Tabell) == breiddPå(b.Tabell) {
				ut[n-1].Tabell.Rader = append(ut[n-1].Tabell.Rader, b.Tabell.Rader...)
				continue
			}
			b.IMerknad = høyrerTilMerknad(ut)
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

		// Held blokka føre paa ei uavslutta setning, kan denne blokka
		// ikkje vera ei overskrift, kor mykje ho enn ser ut som ei. Ein
		// overskriftsregel som slaar til likevel endrar sjølve teksten:
		// «o. ſ. v.» - halen av setninga føre - vart teken for ei
		// underoverskrift, og teksten kom ut som «o) ſ. v», med
		// punktumet bytt mot ein parentes og siste punktum borte.
		framhald := heldFramFrå(ut, t) && !opnarNyttPunkt(ut, t)

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

		case !framhald && erKort(t) && reSeksjon.MatchString(t):
			m := reSeksjon.FindStringSubmatch(t)
			sisteSeksjon = romartal(m[1])
			b.Slag, b.Nummer, b.Tittel = Seksjon, m[1], m[2]

		case !framhald && erKort(t) && erTalOverskrift(t):
			m := reTalseks.FindStringSubmatch(t)
			nr, _ := strconv.Atoi(m[1])
			if nr == sisteSeksjon+1 {
				sisteSeksjon = nr
				b.Slag, b.Nummer, b.Tittel = Seksjon, m[1], m[2]
			} else {
				b.Slag, b.Nummer, b.Tittel = Underseksjon, m[1], m[2]
			}

		// Inne i forkortingsbolken er eit oppslag eit oppslag. Utan
		// denne føre underseksjonsregelen vart «v. — Verbum (S. 56)»
		// til underoverskrifta «v) — verbum (S. 56)».
		case iForkortingar && erKort(t) && reOppslag.MatchString(t):
			m := reOppslag.FindStringSubmatch(t)
			b.Slag, b.Nummer = Oppslag, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, tal(t)-tal(m[2]))

		// Utheving maa krevjast her. «F. Ex. njota, Præſ. nyt; ...» er
		// kort nok og passar mønsteret, men er brødtekst; overskriftene
		// staar utheva, og det er det same signalet boka merkjer alle
		// dei andre overskriftene sine med.
		case !framhald && erKort(t) && heiltUtheva(b) && reStorbokst.MatchString(t):
			m := reStorbokst.FindStringSubmatch(t)
			b.Slag, b.Nummer, b.Tittel = Storbokstav, m[1], m[2]

		// Utheving, ikkje lengd. Boka merkjer overskriftene sine med
		// kursiv eller halvfeit: 24 av dei passar mønsteret og staar
		// utheva, 5 passar utan aa vera utheva - og dei fem er
		// listepunkt. Det var lengda som gjorde «b) Hardangerſk.
		// Fleertal af Maſk. har `ar'; Fl. af Fe-» til ei overskrift og
		// reiv ordet Feminin i to. Ingen ekte overskrift er lengre enn
		// grensa, so ho tapte vi ingenting paa aa sleppe.
		case !framhald && heiltUtheva(b) && reUnderseks.MatchString(t):
			m := reUnderseks.FindStringSubmatch(t)
			b.Slag, b.Nummer, b.Tittel = Underseksjon, m[1], m[2]

		case !framhald && reListepunkt.MatchString(t):
			m := reListepunkt.FindStringSubmatch(t)
			b.Slag, b.Nummer = Listepunkt, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, tal(t)-tal(m[2]))

		case !framhald && reUnderpunkt.MatchString(t):
			m := reUnderpunkt.FindStringSubmatch(t)
			b.Slag, b.Nummer = Underpunkt, m[1]
			b.Stumpar = skjerFramme(b.Stumpar, tal(t)-tal(m[2]))

		// Oppstilte døme. Kvart døme staar som si eiga blokk i kjelda og
		// er ei line for seg i trykket; utan eige slag rann dei saman til
		// vanlege, utlikna avsnitt med innrykk.
		case !framhald && reDøme.MatchString(t):
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

		// Her blir framhaldet limt inn i blokka føre.
		case framhald:
			b.Slag = Brødtekst
			limInn(ut, b.Stumpar)
			continue

		// Boka uthevar overskriftene: Afdeling-namna er halvfeite,
		// mellomoverskriftene kursive. Det er signalet - ikkje lengda.
		// Utan denne prøven vart kvar kort line ei overskrift, og lister
		// som «ei til e: Leir, heil, rein» hamna i registeret.
		case !framhald && erKort(t) && !erInnleiing(t) && heiltUtheva(b):
			b.Slag, b.Tittel = Mellomtittel, strings.TrimSuffix(t, ".")

		default:
			// Held blokka fram setninga frå blokka over, er det eit
			// sideskift i papiret og ikkje eit nytt avsnitt.
			b.Slag = Brødtekst
		}

		// Boka set baade tavler og lister inn under merknadene sine. Held
		// ei slik blokk fram ein Anm., skal ho ha same forminga som han -
		// innrykk, strek og dempa farge - elles ser ho ut som om ho høyrer
		// til brødteksten, og lesaren mistar kva ho er ei utgreiing om.
		// Fram til no galdt dette berre oppsett; punkta og døma under ein
		// merknad stod uforma midt i satsen.
		switch b.Slag {
		case Listepunkt, Underpunkt, Døme, Oppslag:
			b.IMerknad = høyrerTilMerknad(ut)
		}

		ut = append(ut, b)
	}

	return ut
}

// høyrerTilMerknad seier om blokka held fram merknaden over seg - anten
// rett etter han, eller etter ei anna blokk som alt høyrer til han, so ei
// rekkje punkt under same Anm. heng saman heile vegen.
func høyrerTilMerknad(ut []Blokk) bool {
	n := len(ut) - 1
	if n < 0 {
		return false
	}
	return ut[n].Slag == Merknad || ut[n].IMerknad
}

// limInn skøyter eit framhald paa blokka føre. Endar ho paa bindestrek,
// er ordet delt av eit sideskift - «om-» og «trent» - og daa skal streken
// bort og ingen mellomrom setjast inn. Utan dette stod det «om- trent» i
// teksten, 71 stader i verket. Alle 71 er ekte orddelingar; den einaste
// som saag ut som ein hengjande bindestrek, «der-» + «til», er ordet
// dertil.
func limInn(ut []Blokk, nye []Stump) {
	n := len(ut) - 1
	delt := false
	for i := len(ut[n].Stumpar) - 1; i >= 0; i-- {
		r := []rune(ut[n].Stumpar[i].Tekst)
		if len(r) == 0 {
			continue
		}
		if r[len(r)-1] == '-' {
			ut[n].Stumpar[i].Tekst = string(r[:len(r)-1])
			delt = true
		}
		break
	}
	if !delt {
		ut[n].Stumpar = append(ut[n].Stumpar, Stump{Tekst: " "})
	}
	ut[n].Stumpar = append(ut[n].Stumpar, nye...)
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
func heldFramFrå(ut []Blokk, t string) bool {
	n := len(ut) - 1
	if n < 0 || !kanHaldeFram(ut[n].Slag) {
		return false
	}
	return heldFram(ut[n].Tekst()) || opnarSomFramhald(t)
}

// opnarSomFramhald ser paa linja som kjem, ikkje paa den som gjekk føre.
// Ei ny setning i denne boka byrjar alltid med stor bokstav; byrjar linja
// med liten, er ho halen av setninga over - same kva den slutta paa.
//
// Det er naudsynt av di punktumet i ei forkorting ser ut som eit
// setningsslutt: «... f. Ex.» braut linja midt i, og «(Homil.» + «114).»
// vart to blokker, so sidetalet stod att som eit avsnitt for seg. 55
// blokker i verket sluttar paa ei slik forkorting, og aa telje dei opp i
// ei liste ville alltid vera ufullstendig.
//
// Eit merke opnar aldri eit framhald, sjølv om det byrjar med liten
// bokstav - «a) Infinitiv ofteſt ...» er eit nytt underpunkt.
func opnarSomFramhald(t string) bool {
	s := strings.TrimSpace(t)
	if s == "" {
		return false
	}
	if reListepunkt.MatchString(s) || reUnderpunkt.MatchString(s) ||
		reDøme.MatchString(s) || reParagraf.MatchString(s) {
		return false
	}
	r := []rune(s)
	// «114).» - talet og parentesen som lukkar ei tilvising frå lina over.
	if reLukkarParentes.MatchString(s) {
		return true
	}
	return unicode.IsLower(r[0])
}

// opnarNyttPunkt seier om teksten opnar det NESTE punktet i den lista
// blokka føre høyrer til. Berre daa skal eit merke slaa framhaldet av.
//
// Utan dette svelgde eit listepunkt som slutta utan punktum det neste.
// Blokk 1853: «1) Nordenfjelds: ... Varg, Vækja» tok «2) Veſtenfjelds:
// Ange (Aangje), ...» inn i seg som framhald, so lista kom ut nummerert
// 1) og 3), med 2)-merket staaande midt i teksten til det fyrste punktet.
//
// Det er ikkje nok at merket finst. Eit framhald kan sjølv byrje med noko
// som ser ut som eit merke: «a) og fleire.» er halen av setninga føre, og
// gjorde vi kvart merke til eit nytt punkt, reiv vi den setninga i to. So
// merket maa vera det neste i rekkja, etter eit punkt av same slaget.
func opnarNyttPunkt(ut []Blokk, t string) bool {
	n := len(ut) - 1
	if n < 0 {
		return false
	}
	føre := ut[n]
	if m := reListepunkt.FindStringSubmatch(t); m != nil && føre.Slag == Listepunkt {
		nr, feil := strconv.Atoi(m[1])
		før, feil2 := strconv.Atoi(føre.Nummer)
		return feil == nil && feil2 == nil && nr == før+1
	}
	if m := reUnderpunkt.FindStringSubmatch(t); m != nil && føre.Slag == Underpunkt {
		nr, før := []rune(m[1]), []rune(føre.Nummer)
		return len(nr) == 1 && len(før) == 1 && nr[0] == før[0]+1
	}
	return false
}

// breiddPå gjev talet paa kolonnar i den breiaste rada.
func breiddPå(t *Tabell) int {
	b := 0
	for _, r := range t.Rader {
		if n := len(r.Celler); n > b {
			b = n
		}
	}
	return b
}

// Eit sideskift kan dele kva som helst av lause tekstblokker - ògso eit
// listepunkt. «b) Hardangerſk ... Fl. af Fe-» i § 363 er nettopp det.
func kanHaldeFram(s Bolkslag) bool {
	switch s {
	case Brødtekst, Paragraf, Merknad, Listepunkt, Underpunkt:
		return true
	}
	return false
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
