// skann hentar sider frå ei PDF-skanning i høg oppløysing, binariserer dei,
// og køyrer OCR med tesseracts frk-modell (LSTM-fraktur).
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	pdf := flag.String("pdf", "", "sti til PDF-fila")
	utmappe := flag.String("ut", "sider", "mappe for utdata")
	fraSide := flag.Int("fra", 1, "fyrste side")
	tilSide := flag.Int("til", 1, "siste side")
	sprak := flag.String("sprak", "frk", "tesseract-spraakmodell")
	flag.Parse()

	if *pdf == "" {
		fmt.Fprintln(os.Stderr, "treng -pdf")
		os.Exit(1)
	}

	if err := os.MkdirAll(*utmappe, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for side := *fraSide; side <= *tilSide; side++ {
		if err := handsamSide(*pdf, *utmappe, side, *sprak); err != nil {
			fmt.Fprintf(os.Stderr, "side %d: %v\n", side, err)
			continue
		}
		fmt.Printf("side %d ferdig\n", side)
	}
}

func handsamSide(pdf, utmappe string, side int, sprak string) error {
	namn := fmt.Sprintf("side-%03d", side)
	prefiks := filepath.Join(utmappe, namn)

	if err := køyr("pdftoppm", "-png", "-r", "400", "-f", fmt.Sprint(side), "-l", fmt.Sprint(side), pdf, prefiks); err != nil {
		return fmt.Errorf("pdftoppm: %w", err)
	}

	png, err := finnRaaBilete(prefiks)
	if err != nil {
		return err
	}

	bw := prefiks + "-bw.png"
	if err := køyr("magick", png, "-colorspace", "Gray", "-normalize", "-threshold", "55%", bw); err != nil {
		return fmt.Errorf("magick: %w", err)
	}

	if err := køyr("tesseract", bw, prefiks, "-l", sprak, "--oem", "1", "--psm", "3"); err != nil {
		return fmt.Errorf("tesseract: %w", err)
	}

	return nil
}

// pdftoppm legg sidetalet til prefikset sjølv, med talet paddar avhengig
// av kor mange sider PDF-en har totalt - difor let vi han fortelje oss
// filnamnet i staden for å gjette padding sjølve.
func finnRaaBilete(prefiks string) (string, error) {
	treff, err := filepath.Glob(prefiks + "-*.png")
	if err != nil {
		return "", err
	}
	if len(treff) == 0 {
		return "", fmt.Errorf("fann ikkje utdata frå pdftoppm for %s", prefiks)
	}
	return treff[0], nil
}

func køyr(namn string, arg ...string) error {
	cmd := exec.Command(namn, arg...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
