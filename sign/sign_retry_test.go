package sign

import (
	"bytes"
	"crypto"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/digitorus/pdf"
)

// A signature bigger than the estimated placeholder restarts the signing pass.
// The restarted pass writes the output itself, so the pass that started it must
// not write again: doing so left the signed document in the file twice, with the
// signature covering only the first half — the shape every verifier reads as
// "altered after signing".
func TestSignatureLongerThanTheEstimateWritesTheDocumentOnce(t *testing.T) {
	cert, key := loadCertificateAndKey(t)
	input, err := os.Open("../testfiles/testfile20.pdf")
	if err != nil {
		t.Skipf("no test file: %v", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		t.Fatal(err)
	}
	reader, err := pdf.NewReader(input, info.Size())
	if err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(t.TempDir(), "signed.pdf")
	file, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	context := SignContext{
		PDFReader:  reader,
		InputFile:  input,
		OutputFile: file,
		SignData: SignData{
			Signature:   SignDataSignature{CertType: ApprovalSignature},
			Signer:      key,
			Certificate: cert,
			objectId:    uint32(reader.XrefInformation.ItemCount) + 2,
		},
		// Far too small for any real signature, which is the point: this is the
		// path a qualified certificate takes and a throwaway test key does not.
		SignatureMaxLengthBase: 32,
	}
	context.SignData.DigestAlgorithm = crypto.SHA256
	if err := context.SignPDF(); err != nil {
		t.Fatalf("sign: %v", err)
	}

	signed, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if half := len(signed) / 2; len(signed)%2 == 0 && bytes.Equal(signed[:half], signed[half:]) {
		t.Fatalf("the document was written twice: %d bytes, both halves identical", len(signed))
	}

	ranges := regexp.MustCompile(`/ByteRange\s*\[\s*(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s*\]`).FindAllSubmatch(signed, -1)
	if len(ranges) == 0 {
		t.Fatal("signed file carries no /ByteRange")
	}
	last := ranges[len(ranges)-1]
	start, _ := strconv.Atoi(string(last[1]))
	gapEnd, _ := strconv.Atoi(string(last[3]))
	tail, _ := strconv.Atoi(string(last[4]))
	if start != 0 || gapEnd+tail != len(signed) {
		t.Errorf("signature covers %d..%d of a %d byte file", start, gapEnd+tail, len(signed))
	}
}
