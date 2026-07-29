package tekst

import (
	"strings"
	"unicode"
)

// tynnrom er U+2009 THIN SPACE, ein femtedels geviert.
const tynnrom = ' '

// Setningsrom legg inn den breie avstanden boka har etter setningar.
//
// Maalt paa faksimilen: vanleg ordmellomrom er 28 px, mellomrommet etter
// eit setningspunktum 49 px - altso 1,75 gonger. Etter eit
// forkortingspunktum er det derimot 31 px, som er vanleg avstand. Boka
// skil altso mellom dei to, og det maa vi òg, for teksten er full av
// «f. Ex.», «d. v. ſ.» og «G. N.».
//
// Tynnromet staar føre det vanlege mellomrommet, ikkje etter, so linja
// framleis kan brytast paa vanleg vis og tynnromet blir hengande i
// slutten av linja der det ikkje synest.
func Setningsrom(s string) string {
	runar := []rune(s)
	var ut strings.Builder
	ut.Grow(len(s) + 8)

	for i := 0; i < len(runar); i++ {
		r := runar[i]
		ut.WriteRune(r)
		if !erSetningsslutt(r) {
			continue
		}
		if i+1 >= len(runar) || runar[i+1] != ' ' {
			continue
		}
		if !opnarNySetning(runar, i+2) {
			continue
		}
		if r == '.' && erForkorting(runar, i) {
			continue
		}
		ut.WriteRune(tynnrom)
	}
	return ut.String()
}

func erSetningsslutt(r rune) bool {
	return r == '.' || r == '!' || r == '?'
}

// opnarNySetning ser om det som følgjer ser ut som byrjinga paa ei ny
// setning. Boka har store forbokstavar i alle substantiv, so dette er
// ikkje eit sikkert teikn aaleine - men saman med lengda paa ordet føre
// er det godt nok.
func opnarNySetning(runar []rune, i int) bool {
	for i < len(runar) && runar[i] == ' ' {
		i++
	}
	if i >= len(runar) {
		return false
	}
	return unicode.IsUpper(runar[i])
}

// erForkorting ser paa ordet føre punktumet. Ved maalinga skilde vi
// forkortingane ut paa at kjerna er eitt eller to teikn, eller eit tal -
// «f.», «S.», «Chr.», «132.» - og dei viste seg aa ha vanleg avstand
// etter seg.
func erForkorting(runar []rune, punktum int) bool {
	slutt := punktum
	start := slutt
	for start > 0 && (unicode.IsLetter(runar[start-1]) || unicode.IsDigit(runar[start-1])) {
		start--
	}
	kjerne := runar[start:slutt]
	if len(kjerne) == 0 || len(kjerne) <= 2 {
		return true
	}
	for _, r := range kjerne {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
