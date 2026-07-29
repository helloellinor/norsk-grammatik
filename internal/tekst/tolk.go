package tekst

import "regexp"

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
	Oppsett      Bolkslag = "oppsett"      // lydtavle, bøyingsmønster
	Brødtekst    Bolkslag = "brødtekst"
)

var (
	reAfdeling = regexp.MustCompile(`^(Førſte|Første|Anden|Andet|Tredie|Fjerde|Femte) (Afdeling|Tillæg)\.?$`)
	reTalseks  = regexp.MustCompile(`^(\d+)\.\s+(.+?)\.?$`)
	// Framanfor Afdelingane står desse bolkane, og boka listar dei i si
	// eiga innhaldsliste jamsides «Førſte Afdeling» - altso toppnivå.
	// Punktumet skil overskrifta ("Fortale.") frå den same nemninga i
	// 1997-utgåva si eiga vesle innhaldsliste framme ("Fortale").
	reFramanfor = regexp.MustCompile(`^(Fortale|Indledning|Indholdsliſte|Indholdsliste|Forklaring af nogle Forkortninger)\.$`)
	reSeksjon   = regexp.MustCompile(`^([IVX]+)\.\s+(.+?)\.?$`)
	reUnderseks = regexp.MustCompile(`^([a-z])\)\s+(.+?)\.?$`)
	reParagraf  = regexp.MustCompile(`^(\d+)\.?\s+(.+)$`)
	reMerknad   = regexp.MustCompile(`^\s*Anm(?:\.|ærkning\.)\s*`)
	reOppslag   = regexp.MustCompile(`^(\S[^—]{0,24}?)\s+—\s+(.+)$`)
	reSøppel    = regexp.MustCompile(`GOTOBUTTON|PAGEREF|_Toc\d+|^\s*TOC\s`)
)
