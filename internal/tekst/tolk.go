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
	Storbokstav  Bolkslag = "storbokstav"  // A. B. C. - eit hakk over a) b) c)
	Underseksjon Bolkslag = "underseksjon" // a) b) c)
	Mellomtittel Bolkslag = "mellomtittel" // Enkelte Vokaler.
	Paragraf     Bolkslag = "paragraf"     // 8. Det norſke Sprog ...
	Merknad      Bolkslag = "merknad"      // Anm. ...
	Avsnitt      Bolkslag = "avsnitt"      // innrykt framhald
	Oppslag      Bolkslag = "oppslag"      // Ang. — Angelſachſiſk.
	Oppsett      Bolkslag = "oppsett"      // lydtavle, bøyingsmønster
	Døme         Bolkslag = "døme"         // 1, ſleppa ..., af ſleppa, ſlapp.
	Listepunkt   Bolkslag = "listepunkt"   // 1) a til e (aabent), f. Ex. ...
	Underpunkt   Bolkslag = "underpunkt"   // a) Vosſiſk. Fleertal af Maſk. ...
	Brødtekst    Bolkslag = "brødtekst"
)

var (
	reAfdeling = regexp.MustCompile(`^(Førſte|Første|Anden|Andet|Tredie|Fjerde|Femte) (Afdeling|Tillæg)\.?$`)
	reTalseks  = regexp.MustCompile(`^(\d+)\.\s+(.+?)\.?$`)
	// Framanfor Afdelingane står desse bolkane, og boka listar dei i si
	// eiga innhaldsliste jamsides «Førſte Afdeling» - altso toppnivå.
	// Punktumet skil overskrifta ("Fortale.") frå den same nemninga i
	// 1997-utgåva si eiga vesle innhaldsliste framme ("Fortale").
	reFramanfor = regexp.MustCompile(`^(Fortale|Indledning|Indholdsliſte|Indholdsliste|Forklaring af nogle Forkortninger|Enkelte Tillæg og Rettelſer|Enkelte Tillæg og Rettelser|Regiſter|Register)\.$|^Føreord \(1997\)$`)
	reSeksjon   = regexp.MustCompile(`^([IVX]+)\.\s+(.+?)\.?$`)
	// Boka nummererer med store bokstavar eitt hakk over dei smaa:
	// «IV. Orddannelſe efter Betydningen» inneheld «A. Afledede
	// Subſtantiver», som i sin tur inneheld «a) Subſt. af Verbum».
	// Innhaldslista hennar listar A, B og C, men ikkje a, b og c.
	reStorbokst = regexp.MustCompile(`^([A-Z])\.\s+(.+?)\.?$`)
	reUnderseks = regexp.MustCompile(`^([a-z])[.)]\s+(.+?)\.?$`) // boka brukar baade «a)» og «a.»
	reParagraf  = regexp.MustCompile(`^(\d+)\.?\s+(.+)$`)
	reMerknad   = regexp.MustCompile(`^\s*Anm(?:\.|ærkning\.)\s*`)
	// Aasen nummererer oppstilte døme med komma - «1, ſleppa» - og skil
	// dei dermed sjølv frå §-nummera og frå dei nummererte overskriftene,
	// som begge har punktum. I heile boka finst mønsteret ti gonger, i to
	// samanhengande rekkjer paa fem, kvar innleidd av ei line som endar
	// paa kolon. Ingen einslege treff, altso ingen falske.
	reDøme = regexp.MustCompile(`^(\d+),\s+(.+)$`)
	// Punktlister. Boka merkjer dei med parentes - «1)» og «a)» - medan
	// punktum høyrer til §-nummera og dei nummererte overskriftene. Dei
	// nestar: «a)» og «b)» ligg under «3)» i § 363. Ei utheva line med
	// same teiknet er derimot ei overskrift, og blir teken før desse.
	reListepunkt = regexp.MustCompile(`^(\d+)\)\s+(.+)$`)
	reUnderpunkt = regexp.MustCompile(`^([a-z])\)\s+(.+)$`)
	reOppslag    = regexp.MustCompile(`^(\S[^—]{0,24}?)\s+—\s+(.+)$`)
	reSøppel     = regexp.MustCompile(`GOTOBUTTON|PAGEREF|_Toc\d+|^\s*TOC\s`)
)
