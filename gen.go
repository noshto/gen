package gen

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/google/uuid"
	"github.com/noshto/sep"
)

// Params represents collection of parameters needed for Generate function
type Params struct {
	SepConfig *sep.Config
	OutFile   string
}

// PrintUsage prints welcome message
func PrintUsage() {
	fmt.Println("--------------------")
	fmt.Println("Welcome to GEN - simple util that saves you from frustrating process of invoices fiscalization!")
	fmt.Println("--------------------")
	fmt.Println("This app is intented to help you with generation of an invoice request that meets efi.tax.gov.me fiscalization service requirements.")
	fmt.Println("You will be guided through the minimal list of questions sufficient for invoice fiscalization.")
	fmt.Println()
	// fmt.Println("Please check the following list of requirements before proceeding:")
	// fmt.Println("----------------------------------")
	// fmt.Println("PREDUSLOVI ZA FISKALIZACIJU RAČUNA")
	// fmt.Println("PREDUSLOVI ZA FISKALIZACIJU RAČUNA")
}

// Generate generates REgisterInvoiceRequest in a quiz mode
func Generate(params *Params) error {

	// Type Of Invoice
	TypeOfInv := sep.NONCASH
	fmt.Println("Izaberite Vrstu računa:")
	fmt.Println("[1] Gotovinski (CASH)")
	fmt.Println("[2] Bezgotovinski (NONCASH)")
	stringValue := scan("Vrsta računa: ")
	uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return err
	}
	switch uint64Value {
	case 1:
		TypeOfInv = sep.CASH
	case 2:
		TypeOfInv = sep.NONCASH
	default:
		TypeOfInv = sep.NONCASH
	}

	PayMethodType := sep.ACCOUNT
	fmt.Println("Načini plaćanja:")
	switch TypeOfInv {
	case sep.CASH:
		fmt.Println("[1] Novčanice i kovanice (BANKNOTE)")
		fmt.Println("[2] Kreditna i debitna kartica banke izdata fizičkom licu (CARD)")
		fmt.Println("[3] Račun još nije plaćen. Biće plaćen zbirnim računom (ORDER)")
		fmt.Println("[4] Ostala gotovinska plaćanja (OTHER-CASH)")
		stringValue = scan("Način plaćanja: ")
		uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return err
		}
		switch uint64Value {
		case 1:
			PayMethodType = sep.BANKNOTE
		case 2:
			PayMethodType = sep.CARD
		case 3:
			PayMethodType = sep.ORDER
		case 4:
			PayMethodType = sep.OTHER_CASH
		default:
			PayMethodType = sep.OTHER
		}
	case sep.NONCASH:
		fmt.Println("[1] Kreditna i debitna kartica banke izdata poreskom obvezniku (BUSINESSCARD)")
		fmt.Println("[2] Jednokratni vaučer (SVOUCHER)")
		fmt.Println("[3] Kartice izdate od preduzeća prodavca, poklon kartice i slične prepaid kartice (COMPANY)")
		fmt.Println("[4] Račun još nije plaćen. Biće plaćen zbirnim računom (ORDER)")
		fmt.Println("[5] Plaćanje avansom (ADVANCE)")
		fmt.Println("[6] Transakcioni račun (virman) (ACCOUNT)")
		fmt.Println("[7] Faktoring (FACTORING)")
		fmt.Println("[8] Ostala bezgotovinska plaćanja (OTHER)")
		stringValue = scan("Način plaćanja: ")
		uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return err
		}
		switch uint64Value {
		case 1:
			PayMethodType = sep.BUSINESSCARD
		case 2:
			PayMethodType = sep.SVOUCHER
		case 3:
			PayMethodType = sep.COMPANY
		case 4:
			PayMethodType = sep.ORDER
		case 5:
			PayMethodType = sep.ADVANCE
		case 6:
			PayMethodType = sep.ACCOUNT
		case 7:
			PayMethodType = sep.FACTORING
		case 8:
			PayMethodType = sep.OTHER
		default:
			PayMethodType = sep.OTHER
		}
	}

	// Subsequent Delivery Type
	SubseqDelivType := sep.SubseqDelivType("")
	stringValue = scan("Naknadno dostavljanje (da ili ne): ")
	if strings.Compare(stringValue, "da") == 0 {
		fmt.Println("Izaberite tip naknadne dostave:")
		fmt.Println("[1] Ako ENU djeluje u području bez interneta (NOINTERNET)")
		fmt.Println("[2] ENU ne radi i ne može se kreirati poruka (BOUNDBOOK)")
		fmt.Println("[3] Problem sa fiskalnim servisom (SERVICE)")
		fmt.Println("[4] Tehnička greška (TECHNICALERROR)")
		fmt.Println("[5] Naknadno slanje uslovljeno načinom poslovanja (BUSINESSNEED)")
		stringValue = scan("Tip naknadne dostave: ")
		uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return err
		}
		switch uint64Value {
		case 1:
			SubseqDelivType = sep.NOINTERNET
		case 2:
			SubseqDelivType = sep.BOUNDBOOK
		case 3:
			SubseqDelivType = sep.SERVICE
		case 4:
			SubseqDelivType = sep.TECHNICALERROR
		case 5:
			SubseqDelivType = sep.BUSINESSNEED
		default:
			SubseqDelivType = sep.BOUNDBOOK
		}
	}

	// Invoice Ordinal Number
	stringValue = scan("Redni broj računa: ")
	InvOrdNum, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return err
	}
	// Internal Order Number
	InternalOrdNum := scan("Interni broj računa (ostavite prazno ako ne postoji): ")

	// Seller
	Seller := &sep.Seller{
		IDType:  sep.IDTypeTIN,
		IDNum:   params.SepConfig.TIN,
		Name:    params.SepConfig.Name,
		Address: params.SepConfig.Address,
		Town:    params.SepConfig.Town,
		Country: params.SepConfig.Country,
	}

	// Fill in Buyer fields. Check through sep.Config.Buyers by name or PIB
	var Buyer *sep.Buyer
	stringValue = scan("Ime ili PIB kupca: ")
	for _, it := range params.SepConfig.Buyers {
		if strings.Contains(it.Name, stringValue) {
			Buyer = &it
			break
		} else if strings.Contains(it.IDNum, stringValue) {
			Buyer = &it
			break
		}
	}
	if Buyer == nil {
		fmt.Println("Kupac ne postoji, molim upišite sledeči podatke:")
		Buyer = &sep.Buyer{
			IDType:  sep.IDTypeTIN,
			IDNum:   scan(" - Identifikacioni broj kupca (PIB): "),
			Name:    scan(" - Ime kupca: "),
			Address: scan(" - Adresa kupca: "),
			Town:    scan(" - Grad kupca: "),
			Country: scan(" - Država kupca (MNE, USA, itd.): "),
		}
	}

	// Currency
	Currency := &sep.Currency{
		Code:   sep.EUR,
		ExRate: 1.0,
	}
	CurrencyCode := scan("Valuta (EUR, USD, RUB, GBP, itd.): ")
	if strings.Compare(CurrencyCode, string(sep.EUR)) != 0 {
		stringValue = scan(fmt.Sprintf("[5] Kurs razmjene %v od %v: ", Currency, string(sep.EUR)))
		float64Value, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return err
		}
		Currency.Code = sep.CurrencyCodeType(CurrencyCode)
		Currency.ExRate = float64Value
	}

	stringValue = scan("Količina stavke: ")
	NumOfItems, err := strconv.Atoi(stringValue)
	if err != nil {
		return err
	}

	if NumOfItems <= 0 {
		return fmt.Errorf("number of items should be greater than 0")
	}

	// Calculating the following values while fillign in Invoice.Items
	TotPriceWoVAT := 0.0
	TotVATAmt := 0.0
	TotPrice := 0.0
	SameTaxesMap := map[float64][]*sep.Item{}

	// Fill in Invoice.Items
	Items := []*sep.Item{}
	for i := 0; i < NumOfItems; i++ {
		fmt.Printf("Stavka #%d:\n", i+1)
		N := scan("Naziv stavke (roba ili usluge): ")
		U := scan("Jedinica mjere (komad, jedinica za mjerenje težine, jedinica za mjerenje dužine, itd.): ")
		Q := scan("Količina ili broj stavki: ")
		UPB := scan("Jedinična cijena prije dodavanja PDV-a: ")
		VR := scan("Stopa PDV-a: ")
		R := scan("Procenat rabata: ")
		EX := ""
		stringValue = scan("Izuzeće od plaćanja PDV-a (da ili ne): ")
		if strings.Compare(stringValue, "da") == 0 {
			fmt.Println("Izaberite član za izuzeće od plaćanja PDV-a:")
			fmt.Println("[1] Mjesto prometa usluga (Član 17)")
			fmt.Println("[2] Poreska osnovica i ispravka poreske osnovice (Član 20)")
			fmt.Println("[3] Oslobođenja od javnog interesa (Član 26)")
			fmt.Println("[4] Ostala oslobođenja (Član 27)")
			fmt.Println("[5] Oslobođenja kod uvoza proizvoda (Član 28)")
			fmt.Println("[6] Oslobođenja kod privremenog uvoza proizvoda (Član 29)")
			fmt.Println("[7] Posebna oslobođenja (Član 30)")
			EX = scan("Izuzeće od plaćanja PDV-a: ")
		}

		q, err := strconv.ParseFloat(Q, 64)
		if err != nil {
			return err
		}
		upb, err := strconv.ParseFloat(UPB, 64)
		if err != nil {
			return err
		}
		vr, err := strconv.ParseFloat(VR, 64)
		if err != nil {
			return err
		}
		r, err := strconv.ParseFloat(R, 64)
		if err != nil {
			return err
		}

		// Calculations
		// upbr is for "Unit Price Before VAT, Rabat applied"
		upbr := (upb - upb*(r/100))
		// pb is for "Price Before VAT"
		pb := upbr * q

		// uva is for "Unit VAT Amount, Rabat applied"
		uva := upbr * (vr / 100)
		// va is for "VAT Amount"
		va := uva * q

		// upa is for "Unit Price After VAT, Rabat applied"
		upa := upbr + uva
		// pa is for "Price After VAT, Rabat applied"
		pa := pb + va

		Item := &sep.Item{
			N:   N,
			U:   U,
			Q:   sep.Amount(q),
			UPB: sep.Amount(upb),
			UPA: sep.Amount(upa),
			R:   sep.Amount(r),
			RR:  true,
			EX:  sep.ExemptFromVATType(EX),
			PB:  sep.Amount(pb),
			VR:  sep.Amount(vr),
			VA:  sep.Amount(va),
			PA:  sep.Amount(pa),
		}
		Items = append(Items, Item)

		TotPriceWoVAT += pb
		TotVATAmt += va
		TotPrice += pa

		if val, ok := SameTaxesMap[vr]; ok {
			SameTaxesMap[vr] = append(val, Item)
		} else {
			SameTaxesMap[vr] = []*sep.Item{Item}
		}
	}

	// Fill in SameTaxes
	SameTaxes := []*sep.SameTax{}
	for key, value := range SameTaxesMap {
		SameTax := &sep.SameTax{
			NumOfItems: int64(len(value)),
			VATRate:    sep.Amount(key),
		}
		for _, it := range value {
			SameTax.PriceBefVAT += it.PB
			SameTax.VATAmt += it.VA
		}

		SameTaxes = append(SameTaxes, SameTax)
	}

	IsIssuerInVAT := strings.Compare(params.SepConfig.VAT, "") != 0

	PayMethods := []*sep.PayMethod{}
	PayMethods = append(PayMethods, &sep.PayMethod{
		Type: PayMethodType,
		Amt:  sep.Amount(TotPrice),
	})

	IssueDateTime := time.Now()

	InvNum := sep.InvNum(
		fmt.Sprintf(
			"%s/%d/%d/%s",
			params.SepConfig.BusinUnitCode,
			InvOrdNum,
			IssueDateTime.Year(),
			params.SepConfig.TCRCode,
		),
	)

	Invoice := &sep.Invoice{
		TypeOfInv:       sep.TypeOfInv(TypeOfInv),
		IsSimplifiedInv: false,
		IssueDateTime:   sep.DateTime(IssueDateTime),
		InvNum:          InvNum,
		InvOrdNum:       sep.InvOrdNum(InvOrdNum),
		TCRCode:         sep.TCRCode(params.SepConfig.TCRCode),
		IsIssuerInVAT:   IsIssuerInVAT,
		TotPriceWoVAT:   sep.Amount(TotPriceWoVAT),
		TotVATAmt:       sep.Amount(TotVATAmt),
		TotPrice:        sep.Amount(TotPrice),
		OperatorCode:    sep.OperatorCode(params.SepConfig.OperatorCode),
		BusinUnitCode:   sep.BusinUnitCode(params.SepConfig.BusinUnitCode),
		SoftCode:        sep.SoftCode(params.SepConfig.SoftCode),
		IsReverseCharge: false,
		PayMethods:      PayMethods,
		Currency:        Currency,
		Seller:          Seller,
		Buyer:           Buyer,
		Items:           &Items,
		SameTaxes:       &SameTaxes,
		InternalOrdNum:  InternalOrdNum,
	}

	// Generate RegisterInvoiceRequest
	RegisterInvoiceRequest := &sep.RegisterInvoiceRequest{
		ID:      "Request",
		Version: "1",
		Header: sep.Header{
			UUID:            uuid.New().String(),
			SendDateTime:    sep.DateTime(IssueDateTime),
			SubseqDelivType: SubseqDelivType,
		},
		Invoice: *Invoice,
		Signature: sep.Signature{
			SignedInfo: sep.SignedInfo{
				CanonicalizationMethod: sep.CanonicalizationMethod{
					Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
				},
				SignatureMethod: sep.SignatureMethod{
					Algorithm: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
				},
				Reference: sep.Reference{
					URI: "#Request",
					Transforms: []sep.Transform{
						{Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature"},
						{Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#"},
					},
					DigestMethod: sep.DigestMethod{
						Algorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
					},
				},
			},
		},
	}

	buf, err := xml.Marshal(RegisterInvoiceRequest)
	if err != nil {
		return nil
	}

	doc := etree.NewDocument()
	err = doc.ReadFromBytes(buf)
	if err != nil {
		return nil
	}

	doc.IndentTabs()
	// removes extra \n at the ned of the docuemnt
	doc.Root().SetTail("")

	return doc.WriteToFile(params.OutFile)
}

func scan(message string) string {
	fmt.Print(message)
	var value string

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	value = scanner.Text()
	return value
}
