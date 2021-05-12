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
	SepConfig  *sep.Config
	Clients    *[]sep.Client
	OutFile    string
	Simplified bool
}

// GenerateRegisterInvoiceRequest generates RegisterInvoiceRequest in a quiz mode
func GenerateRegisterInvoiceRequest(params *Params) (string, error) {

	// Type Of Invoice
	TypeOfInv := sep.NONCASH
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Izaberite Vrstu računa:")
		fmt.Println("[1] Gotovinski (CASH)")
		fmt.Println("[2] Bezgotovinski (NONCASH)")
		stringValue := Scan("Vrsta računa: ")
		uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return "", err
		}
		switch uint64Value {
		case 1:
			TypeOfInv = sep.CASH
		case 2:
			TypeOfInv = sep.NONCASH
		default:
			return "", fmt.Errorf("invalid TypeOfInv")
		}
	}

	PayMethodType := sep.ACCOUNT
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Načini plaćanja:")
		switch TypeOfInv {
		case sep.CASH:
			fmt.Println("[1] Novčanice i kovanice (BANKNOTE)")
			fmt.Println("[2] Kreditna i debitna kartica banke izdata fizičkom licu (CARD)")
			fmt.Println("[3] Račun još nije plaćen. Biće plaćen zbirnim računom (ORDER)")
			fmt.Println("[4] Ostala gotovinska plaćanja (OTHER-CASH)")
			stringValue := Scan("Način plaćanja: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid PayMethodType")
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
			stringValue := Scan("Način plaćanja: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid PayMethodType")
			}
		}
	}

	// Subsequent Delivery Type
	SubseqDelivType := sep.SubseqDelivType("")
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Naknadno dostavljanje:")
		fmt.Println("[1] Da")
		fmt.Println("[2] Ne")
		stringValue := Scan("Naknadno dostavljanje: ")
		uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return "", err
		}
		if uint64Value == 1 {
			fmt.Println("Izaberite tip naknadne dostave:")
			fmt.Println("[1] Ako ENU djeluje u području bez interneta (NOINTERNET)")
			fmt.Println("[2] ENU ne radi i ne može se kreirati poruka (BOUNDBOOK)")
			fmt.Println("[3] Problem sa fiskalnim servisom (SERVICE)")
			fmt.Println("[4] Tehnička greška (TECHNICALERROR)")
			fmt.Println("[5] Naknadno slanje uslovljeno načinom poslovanja (BUSINESSNEED)")
			stringValue = Scan("Tip naknadne dostave: ")
			uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid SubseqDelivType")
			}
		}
	}

	// Invoice Ordinal Number
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue := Scan("Redni broj računa: ")
	InvOrdNum, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return "", err
	}
	// Internal Order Number
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	InternalOrdNum := Scan("Interni broj računa (ostavite prazno ako ne postoji): ")

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
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Ime ili PIB kupca: ")
	for _, it := range *params.Clients {
		if strings.Contains(it.Name, stringValue) {
			Buyer = &sep.Buyer{
				IDType:  sep.IDTypeTIN,
				IDNum:   it.TIN,
				Name:    it.Name,
				Address: it.Address,
				Town:    it.Town,
				Country: it.Country,
			}
			break
		} else if strings.Contains(it.TIN, stringValue) {
			Buyer = &sep.Buyer{
				IDType:  sep.IDTypeTIN,
				IDNum:   it.TIN,
				Name:    it.Name,
				Address: it.Address,
				Town:    it.Town,
				Country: it.Country,
			}
			break
		}
	}
	if Buyer == nil {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Kupac ne postoji, molim upišite sledeči podatke:")
		Buyer = &sep.Buyer{
			IDType:  sep.IDTypeTIN,
			IDNum:   Scan(" - Identifikacioni broj kupca (PIB): "),
			Name:    Scan(" - Ime kupca: "),
			Address: Scan(" - Adresa kupca: "),
			Town:    Scan(" - Grad kupca: "),
			Country: Scan(" - Država kupca (MNE, USA, itd.): "),
		}
	}

	// Currency
	Currency := &sep.Currency{
		Code:   sep.EUR,
		ExRate: 1.0,
	}
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		CurrencyCode := Scan("Valuta (EUR, USD, RUB, GBP, itd.): ")
		if strings.Compare(CurrencyCode, string(sep.EUR)) != 0 {
			stringValue = Scan(fmt.Sprintf("Kurs razmjene %s od %s: ", string(CurrencyCode), string(sep.EUR)))
			float64Value, err := strconv.ParseFloat(stringValue, 64)
			if err != nil {
				return "", err
			}
			Currency.Code = sep.CurrencyCodeType(CurrencyCode)
			Currency.ExRate = float64Value
		}
	}

	NumOfItems := 1
	// if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		stringValue = Scan("Količina stavke: ")
		NumOfItems, err = strconv.Atoi(stringValue)
		if err != nil {
			return "", err
		}

		if NumOfItems <= 0 {
			return "", fmt.Errorf("number of items should be greater than 0")
		}
	// }

	// Calculating the following values while fillign in Invoice.Items
	TotPriceWoVAT := 0.0
	TotVATAmt := 0.0
	TotPrice := 0.0
	SameTaxesMap := map[float64][]*sep.Item{}

	// Fill in Invoice.Items
	Items := []*sep.Item{}
	for i := 0; i < NumOfItems; i++ {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Printf("Stavka #%d:\n", i+1)
		N := ""
		if params.Simplified {
			fmt.Println("Naziv stavke (roba ili usluge):")
			fmt.Println("[1] Knjigovodstvene usluge za period")
			fmt.Println("[2] Pravne usluge")
			fmt.Println("[3] Izreda zavrsnog racuna za period")
			fmt.Println("[4] Ostalo")
			stringValue := Scan("Naziv stavke (roba ili usluge): ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
			}
			switch uint64Value {
			case 1:
				tmp := Scan("Unesite period: ")
				N = strings.Join([]string{"Knjigovodstvene usluge za", tmp}, " ")
			case 2:
				N = "Pravne usluge"
			case 3:
				tmp := Scan("Unesite period: ")
				N = strings.Join([]string{"Izreda zavrsnog racuna za", tmp}, " ")
			case 4:
				N = Scan("Unesite naziv stavke: ")
			default:
				return "", fmt.Errorf("invalid input")
			}

		} else {
			N = Scan("Naziv stavke (roba ili usluge): ")
		}

		U := "kom"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			U = Scan("Jedinica mjere (komad, jedinica za mjerenje težine, jedinica za mjerenje dužine, itd.): ")
		}

		Q := "1"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			Q = Scan(fmt.Sprintf("Broj %s: ", U))
		}

		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		UPB := Scan("Jedinična cijena prije dodavanja PDV-a: ")

		VR := "21"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			VR = Scan("Stopa PDV-a: ")
		}

		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		R := Scan("Procenat rabata: ")

		EX := sep.ExemptFromVATType("")
		if !params.Simplified {
			EX = sep.ExemptFromVATType("")
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			fmt.Println("Izuzeće od plaćanja PDV-a:")
			fmt.Println("[1] Da")
			fmt.Println("[2] Ne")
			stringValue = Scan("Izuzeće od plaćanja PDV-a: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
			}
			if uint64Value == 1 {
				fmt.Println("---------------------------------------------------------------")
				fmt.Println("Izaberite član za izuzeće od plaćanja PDV-a:")
				fmt.Println("[1] Mjesto prometa usluga (Član 17)")
				fmt.Println("[2] Poreska osnovica i ispravka poreske osnovice (Član 20)")
				fmt.Println("[3] Oslobođenja od javnog interesa (Član 26)")
				fmt.Println("[4] Ostala oslobođenja (Član 27)")
				fmt.Println("[5] Oslobođenja kod uvoza proizvoda (Član 28)")
				fmt.Println("[6] Oslobođenja kod privremenog uvoza proizvoda (Član 29)")
				fmt.Println("[7] Posebna oslobođenja (Član 30)")
				stringValue = Scan("Izuzeće od plaćanja PDV-a: ")
				uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
				if err != nil {
					return "", err
				}
				switch uint64Value {
				case 1:
					EX = sep.CL17
				case 2:
					EX = sep.CL20
				case 3:
					EX = sep.CL26
				case 4:
					EX = sep.CL27
				case 5:
					EX = sep.CL28
				case 6:
					EX = sep.CL29
				case 7:
					EX = sep.CL30
				default:
					return "", fmt.Errorf("invalid EX")
				}
			}
		}

		q, err := strconv.ParseFloat(Q, 64)
		if err != nil {
			return "", err
		}
		upb, err := strconv.ParseFloat(UPB, 64)
		if err != nil {
			return "", err
		}
		vr, err := strconv.ParseFloat(VR, 64)
		if err != nil {
			return "", err
		}
		r, err := strconv.ParseFloat(R, 64)
		if err != nil {
			return "", err
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
			params.SepConfig.TCR.BusinUnitCode,
			InvOrdNum,
			IssueDateTime.Year(),
			params.SepConfig.TCR.TCRCode,
		),
	)

	Invoice := &sep.Invoice{
		TypeOfInv:       sep.TypeOfInv(TypeOfInv),
		IsSimplifiedInv: false,
		IssueDateTime:   sep.DateTime(IssueDateTime),
		InvNum:          InvNum,
		InvOrdNum:       sep.InvOrdNum(InvOrdNum),
		TCRCode:         sep.TCRCode(params.SepConfig.TCR.TCRCode),
		IsIssuerInVAT:   IsIssuerInVAT,
		TotPriceWoVAT:   sep.Amount(TotPriceWoVAT),
		TotVATAmt:       sep.Amount(TotVATAmt),
		TotPrice:        sep.Amount(TotPrice),
		OperatorCode:    sep.OperatorCode(params.SepConfig.OperatorCode),
		BusinUnitCode:   sep.BusinUnitCode(params.SepConfig.TCR.BusinUnitCode),
		SoftCode:        sep.SoftCode(params.SepConfig.TCR.SoftCode),
		IsReverseCharge: false,
		PayMethods:      PayMethods,
		Currency:        Currency,
		Seller:          Seller,
		Buyer:           Buyer,
		Items:           &Items,
		SameTaxes:       &SameTaxes,
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
		return "", err
	}

	doc := etree.NewDocument()
	err = doc.ReadFromBytes(buf)
	if err != nil {
		return "", err
	}

	doc, err = Envelope(doc)
	if err != nil {
		return "", err
	}

	return InternalOrdNum, doc.WriteToFile(params.OutFile)
}

// Envelope wraps up given RegisterInvoiceRequest into standard SOAP Envelope
func Envelope(req *etree.Document) (*etree.Document, error) {
	doc := etree.NewDocument()

	root := doc.CreateElement("s:Envelope")
	root.CreateAttr("xmlns:s", "http://schemas.xmlsoap.org/soap/envelope/")
	body := root.CreateElement("s:Body")
	body.CreateAttr("xmlns:xsd", "http://www.w3.org/2001/XMLSchema")
	body.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")

	if len(req.Child) != 1 {
		return nil, fmt.Errorf("Invalid XML document")
	}
	body.AddChild(req.Child[0])

	doc.IndentTabs()
	// removes extra \n at the ned of the docuemnt
	doc.Root().SetTail("")

	return doc, nil
}

// Scan helper for reading user input
func Scan(message string) string {
	fmt.Print(message)
	var value string

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	value = scanner.Text()
	return value
}

// GenerateRegisterTCRRequest asks user to fill in TCR details
func GenerateRegisterTCRRequest(params *Params) error {
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	fmt.Println("REGISTRACIJA ENU")
	fmt.Println("Tip ENU:")
	fmt.Println("[1] Standardni ENU")
	fmt.Println("[2] Samonaplatni uređaj (automat)")
	stringValue := Scan("Tip ENU: ")
	uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return err
	}
	TCRType := sep.REGULAR
	switch uint64Value {
	case 1:
		TCRType = sep.REGULAR
	case 2:
		TCRType = sep.VENDING
	default:
		return fmt.Errorf("invalid TCRType")
	}

	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	TCRIntID := Scan("Interna identifikacija ENU: ")

	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Datum od kojeg će se koristiti ENU (u formati yyyy-MM-dd): ")
	ValidFrom, err := time.Parse("2006-01-02", stringValue)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Datum do kojeg će se koristiti ENU. (u formati yyyy-MM-dd): ")
	ValidTo, err := time.Parse("2006-01-02", stringValue)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	SoftCode := Scan("Kôd softvera: ")

	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	MaintainerCode := Scan("Kôd održavaoca: ")

	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	BusinUnitCode := Scan("Kôd poslovnog prostora: ")

	TCR := sep.TCR{
		Type:           TCRType,
		ValidFrom:      sep.Date(ValidFrom),
		ValidTo:        sep.Date(ValidTo),
		TCRIntID:       sep.TCRIntID(TCRIntID),
		IssuerTIN:      sep.TIN(params.SepConfig.TIN),
		SoftCode:       sep.SoftCode(SoftCode),
		MaintainerCode: sep.MaintainerCode(MaintainerCode),
		BusinUnitCode:  sep.BusinUnitCode(BusinUnitCode),
	}

	RegisterTCRRequest := sep.RegisterTCRRequest{
		ID:      "Request",
		Version: "1",
		Header: sep.Header{
			UUID:         uuid.New().String(),
			SendDateTime: sep.DateTime(time.Now()),
		},
		TCR: TCR,
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

	buf, err := xml.Marshal(RegisterTCRRequest)
	if err != nil {
		return nil
	}

	doc := etree.NewDocument()
	err = doc.ReadFromBytes(buf)
	if err != nil {
		return nil
	}

	doc, err = Envelope(doc)
	if err != nil {
		return nil
	}

	return doc.WriteToFile(params.OutFile)
}

// GeneratePlainIIC asks user to enter parameters required for generating IIC
func GeneratePlainIIC() [7]string {
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	TIN := Scan("Identifikacioni broj prodavca(PIB): ")
	IssueDateTime := Scan("Datum i vrijeme kada je račun kreiran i izdat od strane ENU.: ")
	InvOrdNum := Scan("Redni broj računa: ")
	BusinUnitCode := Scan("Kôd poslovne jedinice (prostora): ")
	TCRCode := Scan("Kôd elektronskog naplatnog uređaja: ")
	SoftCode := Scan("Kôd softvera: ")
	TotPrice := Scan("Ukupna cijena svih stavki uključujući poreze i popuste: ")

	// Orders of parameters: TIN, IssueDateTime, InvOrdNum, BusinUnitCode, TCRCode, SoftCode, TotPrice
	return [7]string{TIN, IssueDateTime, InvOrdNum, BusinUnitCode, TCRCode, SoftCode, TotPrice}
}

// PrintInvoiceDetails prints out invoice details
func PrintInvoiceDetails(inFile string, SepConfig *sep.Config, Clients *[]sep.Client, InternalOrdNum string) error {

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(inFile); err != nil {
		return err
	}
	elem := doc.FindElement("//RegisterInvoiceRequest")
	if elem == nil {
		return fmt.Errorf("invalid xml, no RegisterInvoiceRequest")
	}
	reqDoc := etree.NewDocument()
	reqDoc.SetRoot(elem.Copy())
	buf, err := reqDoc.WriteToBytes()
	if err != nil {
		return err
	}
	req := sep.RegisterInvoiceRequest{}
	if err := xml.Unmarshal(buf, &req); err != nil {
		return err
	}

	fmt.Println("---------------------------------------------------------------")
	fmt.Println("PRODAVAC")
	fmt.Printf("\t%s\n", req.Invoice.Seller.Name)
	fmt.Printf("\t%s\n", req.Invoice.Seller.Address)
	fmt.Printf("\tTel: %s\tFax: %s\n", SepConfig.Phone, SepConfig.Fax)
	fmt.Printf("\tPIB: %s\t\tPDV: %s\n", SepConfig.TIN, SepConfig.VAT)
	fmt.Printf("\tZ.R.: %s\n", SepConfig.BankAccount)
	fmt.Println("---------------------------------------------------------------")

	client := &sep.Client{}
	for _, it := range *Clients {
		if it.TIN == req.Invoice.Buyer.IDNum {
			client = &it
			break
		}
	}
	fmt.Println("KUPAC")
	fmt.Printf("\t%s\n", req.Invoice.Buyer.Name)
	fmt.Printf("\t%s\n", req.Invoice.Buyer.Address)
	fmt.Printf("\tPIB: %s\t\tPDV: %s\n", req.Invoice.Buyer.IDNum, client.VAT)
	fmt.Println("---------------------------------------------------------------")

	fmt.Println("RACUN")
	fmt.Printf("\tDATUM: %s\n", time.Time(req.Invoice.IssueDateTime).Format("2006-01-02"))
	fmt.Printf("\tRACUN br.: %s\n", req.Invoice.InvNum)
	fmt.Printf("\tINTERNI br.: %s\n", InternalOrdNum)
	fmt.Println("---------------------------------------------------------------")

	fmt.Println("STAVKE")
	fmt.Println("Rb\tNAZIV PROIZVODA/USLUGE\tJM\tKolicina\tCijena bez PDV\tVrijednost bez PDV\tRabat\tPDV Stopa\tPDV Iznos\tCijena sa PDV\tVrijednost sa PDV")
	for in, it := range *req.Invoice.Items {

		q := float64(it.Q)
		upb := float64(it.UPB)
		vr := float64(it.VR)
		r := float64(it.R)

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

		Name := it.N
		Unit := it.U
		Quantity := strconv.FormatFloat(float64(it.Q), 'f', 2, 64)
		UnitPriceBefVAT := strconv.FormatFloat(upb, 'f', 2, 64)
		PriceBefVAT := strconv.FormatFloat(upb*q, 'f', 2, 64)
		Rebate := fmt.Sprintf("%d%%", int64(r))
		VATRate := fmt.Sprintf("%d%%", int64(vr))
		VATAmount := strconv.FormatFloat(va, 'f', 2, 64)
		UnitPriceAfterVAT := strconv.FormatFloat(upa, 'f', 2, 64)
		PriceAfterVAT := strconv.FormatFloat(pa, 'f', 2, 64)

		fmt.Printf(
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			strconv.Itoa(in+1),
			fmt.Sprintf("%.22s", fmt.Sprintf("%-22s", Name)),
			fmt.Sprintf("%.4s", fmt.Sprintf("%-4s", Unit)),
			fmt.Sprintf("%.10s", fmt.Sprintf("%-10s", Quantity)),
			fmt.Sprintf("%.14s", fmt.Sprintf("%-15s", UnitPriceBefVAT)),
			fmt.Sprintf("%.18s", fmt.Sprintf("%-18s", PriceBefVAT)),
			fmt.Sprintf("%.6s", fmt.Sprintf("%-6s", Rebate)),
			fmt.Sprintf("%.9s", fmt.Sprintf("%-9s", VATRate)),
			fmt.Sprintf("%.9s", fmt.Sprintf("%-9s", VATAmount)),
			fmt.Sprintf("%.13s", fmt.Sprintf("%-13s", UnitPriceAfterVAT)),
			fmt.Sprintf("%.17s", fmt.Sprintf("%-17s", PriceAfterVAT)),
		)
	}
	fmt.Println("---------------------------------------------------------------")

	PriceBeforeVAT := float64(0)
	Rebate := float64(0)
	VATAmt := float64(0)
	for _, it := range *req.Invoice.Items {
		PriceBeforeVAT += float64(it.UPB * it.Q)
		Rebate += PriceBeforeVAT * (float64(it.R) / 100)
		VATAmt += float64(it.VA)
	}
	Base21 := PriceBeforeVAT - Rebate
	TotPrice := strconv.FormatFloat(float64(req.Invoice.TotPrice), 'f', 2, 64)
	if Rebate != 0 {
		Rebate *= -1
	}

	title := fmt.Sprintf("%.22s", fmt.Sprintf("%-22s", "Vrijednost bez PDV:"))
	fmt.Printf("%s\t\t%s\n", title, strconv.FormatFloat(PriceBeforeVAT, 'f', 2, 64))
	title = fmt.Sprintf("%.22s", fmt.Sprintf("%-22s", "Iznos rabata:"))
	fmt.Printf("%s\t\t%s\n", title, strconv.FormatFloat(Rebate, 'f', 2, 64))
	fmt.Println("---------------------------------------------------------------")
	title = fmt.Sprintf("%.22s", fmt.Sprintf("%-22s", "Osnovica za stopu 21%:"))
	fmt.Printf("%s\t\t%s\n", title, strconv.FormatFloat(Base21, 'f', 2, 64))
	title = fmt.Sprintf("%.22s", fmt.Sprintf("%-22s", "PDV po stopi 21%:"))
	fmt.Printf("%s\t\t%s\n", title, strconv.FormatFloat(VATAmt, 'f', 2, 64))
	fmt.Println("---------------------------------------------------------------")
	title = fmt.Sprintf("%.22s", fmt.Sprintf("%-22s", "IZNOS ZA UPLATU:"))
	fmt.Printf("%s\t\t%s\n", title, TotPrice)
	fmt.Println()
	return nil
}

// GenerateRegisterInvoiceRequest generates RegisterInvoiceRequest in a quiz mode
func GenerateCorrectiveRegisterInvoiceRequest(params *Params) (string, error) {

	fmt.Println("Korektivni račun")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	fmt.Println()

	IICRef := Scan("IKOF referenca na originalni račun: ")
	
	fmt.Println("---------------------------------------------------------------")
	fmt.Println()
	stringValue := Scan("Datum i vrijeme kada je originalni račun kreiran i izdat od strane ENU: ")
	CorrectiveInvIssueDateTime, err := time.Parse(time.RFC3339, stringValue)
	if err != nil {
		return "", err
	}

	// Type Of Invoice
	TypeOfInv := sep.NONCASH
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Izaberite Vrstu računa:")
		fmt.Println("[1] Gotovinski (CASH)")
		fmt.Println("[2] Bezgotovinski (NONCASH)")
		stringValue := Scan("Vrsta računa: ")
		uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return "", err
		}
		switch uint64Value {
		case 1:
			TypeOfInv = sep.CASH
		case 2:
			TypeOfInv = sep.NONCASH
		default:
			return "", fmt.Errorf("invalid TypeOfInv")
		}
	}

	PayMethodType := sep.ACCOUNT
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Načini plaćanja:")
		switch TypeOfInv {
		case sep.CASH:
			fmt.Println("[1] Novčanice i kovanice (BANKNOTE)")
			fmt.Println("[2] Kreditna i debitna kartica banke izdata fizičkom licu (CARD)")
			fmt.Println("[3] Račun još nije plaćen. Biće plaćen zbirnim računom (ORDER)")
			fmt.Println("[4] Ostala gotovinska plaćanja (OTHER-CASH)")
			stringValue := Scan("Način plaćanja: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid PayMethodType")
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
			stringValue := Scan("Način plaćanja: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid PayMethodType")
			}
		}
	}

	// Subsequent Delivery Type
	SubseqDelivType := sep.SubseqDelivType("")
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Naknadno dostavljanje:")
		fmt.Println("[1] Da")
		fmt.Println("[2] Ne")
		stringValue := Scan("Naknadno dostavljanje: ")
		uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return "", err
		}
		if uint64Value == 1 {
			fmt.Println("Izaberite tip naknadne dostave:")
			fmt.Println("[1] Ako ENU djeluje u području bez interneta (NOINTERNET)")
			fmt.Println("[2] ENU ne radi i ne može se kreirati poruka (BOUNDBOOK)")
			fmt.Println("[3] Problem sa fiskalnim servisom (SERVICE)")
			fmt.Println("[4] Tehnička greška (TECHNICALERROR)")
			fmt.Println("[5] Naknadno slanje uslovljeno načinom poslovanja (BUSINESSNEED)")
			stringValue = Scan("Tip naknadne dostave: ")
			uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid SubseqDelivType")
			}
		}
	}

	// Invoice Ordinal Number
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Redni broj računa: ")
	InvOrdNum, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return "", err
	}
	// Internal Order Number
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	InternalOrdNum := Scan("Interni broj računa (ostavite prazno ako ne postoji): ")

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
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Ime ili PIB kupca: ")
	for _, it := range *params.Clients {
		if strings.Contains(it.Name, stringValue) {
			Buyer = &sep.Buyer{
				IDType:  sep.IDTypeTIN,
				IDNum:   it.TIN,
				Name:    it.Name,
				Address: it.Address,
				Town:    it.Town,
				Country: it.Country,
			}
			break
		} else if strings.Contains(it.TIN, stringValue) {
			Buyer = &sep.Buyer{
				IDType:  sep.IDTypeTIN,
				IDNum:   it.TIN,
				Name:    it.Name,
				Address: it.Address,
				Town:    it.Town,
				Country: it.Country,
			}
			break
		}
	}
	if Buyer == nil {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Kupac ne postoji, molim upišite sledeči podatke:")
		Buyer = &sep.Buyer{
			IDType:  sep.IDTypeTIN,
			IDNum:   Scan(" - Identifikacioni broj kupca (PIB): "),
			Name:    Scan(" - Ime kupca: "),
			Address: Scan(" - Adresa kupca: "),
			Town:    Scan(" - Grad kupca: "),
			Country: Scan(" - Država kupca (MNE, USA, itd.): "),
		}
	}

	// Currency
	Currency := &sep.Currency{
		Code:   sep.EUR,
		ExRate: 1.0,
	}
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		CurrencyCode := Scan("Valuta (EUR, USD, RUB, GBP, itd.): ")
		if strings.Compare(CurrencyCode, string(sep.EUR)) != 0 {
			stringValue = Scan(fmt.Sprintf("Kurs razmjene %s od %s: ", string(CurrencyCode), string(sep.EUR)))
			float64Value, err := strconv.ParseFloat(stringValue, 64)
			if err != nil {
				return "", err
			}
			Currency.Code = sep.CurrencyCodeType(CurrencyCode)
			Currency.ExRate = float64Value
		}
	}

	NumOfItems := 1
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		stringValue = Scan("Količina stavke: ")
		NumOfItems, err = strconv.Atoi(stringValue)
		if err != nil {
			return "", err
		}

		if NumOfItems <= 0 {
			return "", fmt.Errorf("number of items should be greater than 0")
		}
	}

	// Calculating the following values while fillign in Invoice.Items
	TotPriceWoVAT := 0.0
	TotVATAmt := 0.0
	TotPrice := 0.0
	SameTaxesMap := map[float64][]*sep.Item{}

	// Fill in Invoice.Items
	Items := []*sep.Item{}
	for i := 0; i < NumOfItems; i++ {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Printf("Stavka #%d:\n", i+1)
		N := ""
		if params.Simplified {
			fmt.Println("Naziv stavke (roba ili usluge):")
			fmt.Println("[1] Knjigovodstvene usluge za period")
			fmt.Println("[2] Pravne usluge")
			fmt.Println("[3] Izreda zavrsnog racuna za period")
			fmt.Println("[4] Ostalo")
			stringValue := Scan("Naziv stavke (roba ili usluge): ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
			}
			switch uint64Value {
			case 1:
				tmp := Scan("Unesite period: ")
				N = strings.Join([]string{"Knjigovodstvene usluge za", tmp}, " ")
			case 2:
				N = "Pravne usluge"
			case 3:
				tmp := Scan("Unesite period: ")
				N = strings.Join([]string{"Izreda zavrsnog racuna za", tmp}, " ")
			case 4:
				N = Scan("Unesite naziv stavke: ")
			default:
				return "", fmt.Errorf("invalid input")
			}

		} else {
			N = Scan("Naziv stavke (roba ili usluge): ")
		}

		U := "kom"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			U = Scan("Jedinica mjere (komad, jedinica za mjerenje težine, jedinica za mjerenje dužine, itd.): ")
		}

		Q := "1"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			Q = Scan(fmt.Sprintf("Broj %s: ", U))
		}

		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		UPB := Scan("Jedinična cijena prije dodavanja PDV-a: ")

		VR := "21"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			VR = Scan("Stopa PDV-a: ")
		}

		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		R := Scan("Procenat rabata: ")

		EX := sep.ExemptFromVATType("")
		if !params.Simplified {
			EX = sep.ExemptFromVATType("")
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			fmt.Println("Izuzeće od plaćanja PDV-a:")
			fmt.Println("[1] Da")
			fmt.Println("[2] Ne")
			stringValue = Scan("Izuzeće od plaćanja PDV-a: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
			}
			if uint64Value == 1 {
				fmt.Println("---------------------------------------------------------------")
				fmt.Println("Izaberite član za izuzeće od plaćanja PDV-a:")
				fmt.Println("[1] Mjesto prometa usluga (Član 17)")
				fmt.Println("[2] Poreska osnovica i ispravka poreske osnovice (Član 20)")
				fmt.Println("[3] Oslobođenja od javnog interesa (Član 26)")
				fmt.Println("[4] Ostala oslobođenja (Član 27)")
				fmt.Println("[5] Oslobođenja kod uvoza proizvoda (Član 28)")
				fmt.Println("[6] Oslobođenja kod privremenog uvoza proizvoda (Član 29)")
				fmt.Println("[7] Posebna oslobođenja (Član 30)")
				stringValue = Scan("Izuzeće od plaćanja PDV-a: ")
				uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
				if err != nil {
					return "", err
				}
				switch uint64Value {
				case 1:
					EX = sep.CL17
				case 2:
					EX = sep.CL20
				case 3:
					EX = sep.CL26
				case 4:
					EX = sep.CL27
				case 5:
					EX = sep.CL28
				case 6:
					EX = sep.CL29
				case 7:
					EX = sep.CL30
				default:
					return "", fmt.Errorf("invalid EX")
				}
			}
		}

		q, err := strconv.ParseFloat(Q, 64)
		if err != nil {
			return "", err
		}
		upb, err := strconv.ParseFloat(UPB, 64)
		if err != nil {
			return "", err
		}
		vr, err := strconv.ParseFloat(VR, 64)
		if err != nil {
			return "", err
		}
		r, err := strconv.ParseFloat(R, 64)
		if err != nil {
			return "", err
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

		if va <= 0.0 {
			va = 0.0
		}

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
			params.SepConfig.TCR.BusinUnitCode,
			InvOrdNum,
			IssueDateTime.Year(),
			params.SepConfig.TCR.TCRCode,
		),
	)

	Invoice := &sep.Invoice{
		TypeOfInv:       sep.TypeOfInv(TypeOfInv),
		IsSimplifiedInv: false,
		IssueDateTime:   sep.DateTime(IssueDateTime),
		InvNum:          InvNum,
		InvOrdNum:       sep.InvOrdNum(InvOrdNum),
		TCRCode:         sep.TCRCode(params.SepConfig.TCR.TCRCode),
		IsIssuerInVAT:   IsIssuerInVAT,
		TotPriceWoVAT:   sep.Amount(TotPriceWoVAT),
		TotVATAmt:       sep.Amount(TotVATAmt),
		TotPrice:        sep.Amount(TotPrice),
		OperatorCode:    sep.OperatorCode(params.SepConfig.OperatorCode),
		BusinUnitCode:   sep.BusinUnitCode(params.SepConfig.TCR.BusinUnitCode),
		SoftCode:        sep.SoftCode(params.SepConfig.TCR.SoftCode),
		IsReverseCharge: false,
		PayMethods:      PayMethods,
		Currency:        Currency,
		Seller:          Seller,
		Buyer:           Buyer,
		Items:           &Items,
		SameTaxes:       &SameTaxes,
		CorrectiveInv: 	 &sep.CorrectiveInv{
			IICRef: 		IICRef,
			IssueDateTime: 	CorrectiveInvIssueDateTime,
			Type: 			sep.CORRECTIVE,
		},
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
		return "", err
	}

	doc := etree.NewDocument()
	err = doc.ReadFromBytes(buf)
	if err != nil {
		return "", err
	}

	doc, err = Envelope(doc)
	if err != nil {
		return "", err
	}

	return InternalOrdNum, doc.WriteToFile(params.OutFile)
}

// GenerateRegisterInvoiceRequest generates RegisterInvoiceRequest in a quiz mode
func GenerateSummaryRegisterInvoiceRequest(params *Params) (string, error) {

	fmt.Println()
	fmt.Println("Sumarni račun")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	fmt.Println()

	SumInvIICRefs := []sep.SumInvIICRef{}
	
	IICRef := Scan("IKOF referenca na originalni račun: ")

	fmt.Println("---------------------------------------------------------------")
	fmt.Println()
	stringValue := Scan("Datum i vrijeme kada je originalni račun kreiran i izdat od strane ENU: ")
	CorrectiveInvIssueDateTime, err := time.Parse(time.RFC3339, stringValue)
	if err != nil {
		return "", err
	}

	SumInvIICRefs = append(
		SumInvIICRefs,
		sep.SumInvIICRef{
			IIC: IICRef,
			IssueDateTime: CorrectiveInvIssueDateTime,
		},
	)

	stringValue = Scan("Koliko korektivnih računa povezano sa originalnim računom: ")
	uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return "", err
	}
	if uint64Value == 0 {
		return "", fmt.Errorf("invalid num")
	}
	for i := 0; i < int(uint64Value); i++ {
		IICRef = Scan(fmt.Sprintf("IKOF referenca na originalni račun #%d: ", i))
		fmt.Println("---------------------------------------------------------------")
		fmt.Println()
		stringValue := Scan("Datum i vrijeme kada je originalni račun kreiran i izdat od strane ENU: ")
		CorrectiveInvIssueDateTime, err = time.Parse(time.RFC3339, stringValue)
		if err != nil {
			return "", err
		}

		SumInvIICRefs = append(
			SumInvIICRefs,
			sep.SumInvIICRef{
				IIC: IICRef,
				IssueDateTime: CorrectiveInvIssueDateTime,
			},
		)
	}
	
	// Type Of Invoice
	TypeOfInv := sep.NONCASH
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Izaberite Vrstu računa:")
		fmt.Println("[1] Gotovinski (CASH)")
		fmt.Println("[2] Bezgotovinski (NONCASH)")
		stringValue := Scan("Vrsta računa: ")
		uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return "", err
		}
		switch uint64Value {
		case 1:
			TypeOfInv = sep.CASH
		case 2:
			TypeOfInv = sep.NONCASH
		default:
			return "", fmt.Errorf("invalid TypeOfInv")
		}
	}

	PayMethodType := sep.ACCOUNT
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Načini plaćanja:")
		switch TypeOfInv {
		case sep.CASH:
			fmt.Println("[1] Novčanice i kovanice (BANKNOTE)")
			fmt.Println("[2] Kreditna i debitna kartica banke izdata fizičkom licu (CARD)")
			fmt.Println("[3] Račun još nije plaćen. Biće plaćen zbirnim računom (ORDER)")
			fmt.Println("[4] Ostala gotovinska plaćanja (OTHER-CASH)")
			stringValue := Scan("Način plaćanja: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid PayMethodType")
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
			stringValue := Scan("Način plaćanja: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid PayMethodType")
			}
		}
	}

	// Subsequent Delivery Type
	SubseqDelivType := sep.SubseqDelivType("")
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Naknadno dostavljanje:")
		fmt.Println("[1] Da")
		fmt.Println("[2] Ne")
		stringValue := Scan("Naknadno dostavljanje: ")
		uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return "", err
		}
		if uint64Value == 1 {
			fmt.Println("Izaberite tip naknadne dostave:")
			fmt.Println("[1] Ako ENU djeluje u području bez interneta (NOINTERNET)")
			fmt.Println("[2] ENU ne radi i ne može se kreirati poruka (BOUNDBOOK)")
			fmt.Println("[3] Problem sa fiskalnim servisom (SERVICE)")
			fmt.Println("[4] Tehnička greška (TECHNICALERROR)")
			fmt.Println("[5] Naknadno slanje uslovljeno načinom poslovanja (BUSINESSNEED)")
			stringValue = Scan("Tip naknadne dostave: ")
			uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("invalid SubseqDelivType")
			}
		}
	}

	// Invoice Ordinal Number
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Redni broj računa: ")
	InvOrdNum, err := strconv.ParseUint(stringValue, 10, 64)
	if err != nil {
		return "", err
	}
	// Internal Order Number
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	InternalOrdNum := Scan("Interni broj računa (ostavite prazno ako ne postoji): ")

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
	fmt.Println()
	fmt.Println("---------------------------------------------------------------")
	stringValue = Scan("Ime ili PIB kupca: ")
	for _, it := range *params.Clients {
		if strings.Contains(it.Name, stringValue) {
			Buyer = &sep.Buyer{
				IDType:  sep.IDTypeTIN,
				IDNum:   it.TIN,
				Name:    it.Name,
				Address: it.Address,
				Town:    it.Town,
				Country: it.Country,
			}
			break
		} else if strings.Contains(it.TIN, stringValue) {
			Buyer = &sep.Buyer{
				IDType:  sep.IDTypeTIN,
				IDNum:   it.TIN,
				Name:    it.Name,
				Address: it.Address,
				Town:    it.Town,
				Country: it.Country,
			}
			break
		}
	}
	if Buyer == nil {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Println("Kupac ne postoji, molim upišite sledeči podatke:")
		Buyer = &sep.Buyer{
			IDType:  sep.IDTypeTIN,
			IDNum:   Scan(" - Identifikacioni broj kupca (PIB): "),
			Name:    Scan(" - Ime kupca: "),
			Address: Scan(" - Adresa kupca: "),
			Town:    Scan(" - Grad kupca: "),
			Country: Scan(" - Država kupca (MNE, USA, itd.): "),
		}
	}

	// Currency
	Currency := &sep.Currency{
		Code:   sep.EUR,
		ExRate: 1.0,
	}
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		CurrencyCode := Scan("Valuta (EUR, USD, RUB, GBP, itd.): ")
		if strings.Compare(CurrencyCode, string(sep.EUR)) != 0 {
			stringValue = Scan(fmt.Sprintf("Kurs razmjene %s od %s: ", string(CurrencyCode), string(sep.EUR)))
			float64Value, err := strconv.ParseFloat(stringValue, 64)
			if err != nil {
				return "", err
			}
			Currency.Code = sep.CurrencyCodeType(CurrencyCode)
			Currency.ExRate = float64Value
		}
	}

	NumOfItems := 1
	if !params.Simplified {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		stringValue = Scan("Količina stavke: ")
		NumOfItems, err = strconv.Atoi(stringValue)
		if err != nil {
			return "", err
		}

		if NumOfItems <= 0 {
			return "", fmt.Errorf("number of items should be greater than 0")
		}
	}

	// Calculating the following values while fillign in Invoice.Items
	TotPriceWoVAT := 0.0
	TotVATAmt := 0.0
	TotPrice := 0.0
	SameTaxesMap := map[float64][]*sep.Item{}

	// Fill in Invoice.Items
	Items := []*sep.Item{}
	for i := 0; i < NumOfItems; i++ {
		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		fmt.Printf("Stavka #%d:\n", i+1)
		N := ""
		if params.Simplified {
			fmt.Println("Naziv stavke (roba ili usluge):")
			fmt.Println("[1] Knjigovodstvene usluge za period")
			fmt.Println("[2] Pravne usluge")
			fmt.Println("[3] Izreda zavrsnog racuna za period")
			fmt.Println("[4] Ostalo")
			stringValue := Scan("Naziv stavke (roba ili usluge): ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
			}
			switch uint64Value {
			case 1:
				tmp := Scan("Unesite period: ")
				N = strings.Join([]string{"Knjigovodstvene usluge za", tmp}, " ")
			case 2:
				N = "Pravne usluge"
			case 3:
				tmp := Scan("Unesite period: ")
				N = strings.Join([]string{"Izreda zavrsnog racuna za", tmp}, " ")
			case 4:
				N = Scan("Unesite naziv stavke: ")
			default:
				return "", fmt.Errorf("invalid input")
			}

		} else {
			N = Scan("Naziv stavke (roba ili usluge): ")
		}

		U := "kom"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			U = Scan("Jedinica mjere (komad, jedinica za mjerenje težine, jedinica za mjerenje dužine, itd.): ")
		}

		Q := "1"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			Q = Scan(fmt.Sprintf("Broj %s: ", U))
		}

		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		UPB := Scan("Jedinična cijena prije dodavanja PDV-a: ")

		VR := "21"
		if !params.Simplified {
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			VR = Scan("Stopa PDV-a: ")
		}

		fmt.Println()
		fmt.Println("---------------------------------------------------------------")
		R := Scan("Procenat rabata: ")

		EX := sep.ExemptFromVATType("")
		if !params.Simplified {
			EX = sep.ExemptFromVATType("")
			fmt.Println()
			fmt.Println("---------------------------------------------------------------")
			fmt.Println("Izuzeće od plaćanja PDV-a:")
			fmt.Println("[1] Da")
			fmt.Println("[2] Ne")
			stringValue = Scan("Izuzeće od plaćanja PDV-a: ")
			uint64Value, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return "", err
			}
			if uint64Value == 1 {
				fmt.Println("---------------------------------------------------------------")
				fmt.Println("Izaberite član za izuzeće od plaćanja PDV-a:")
				fmt.Println("[1] Mjesto prometa usluga (Član 17)")
				fmt.Println("[2] Poreska osnovica i ispravka poreske osnovice (Član 20)")
				fmt.Println("[3] Oslobođenja od javnog interesa (Član 26)")
				fmt.Println("[4] Ostala oslobođenja (Član 27)")
				fmt.Println("[5] Oslobođenja kod uvoza proizvoda (Član 28)")
				fmt.Println("[6] Oslobođenja kod privremenog uvoza proizvoda (Član 29)")
				fmt.Println("[7] Posebna oslobođenja (Član 30)")
				stringValue = Scan("Izuzeće od plaćanja PDV-a: ")
				uint64Value, err = strconv.ParseUint(stringValue, 10, 64)
				if err != nil {
					return "", err
				}
				switch uint64Value {
				case 1:
					EX = sep.CL17
				case 2:
					EX = sep.CL20
				case 3:
					EX = sep.CL26
				case 4:
					EX = sep.CL27
				case 5:
					EX = sep.CL28
				case 6:
					EX = sep.CL29
				case 7:
					EX = sep.CL30
				default:
					return "", fmt.Errorf("invalid EX")
				}
			}
		}

		q, err := strconv.ParseFloat(Q, 64)
		if err != nil {
			return "", err
		}
		upb, err := strconv.ParseFloat(UPB, 64)
		if err != nil {
			return "", err
		}
		vr, err := strconv.ParseFloat(VR, 64)
		if err != nil {
			return "", err
		}
		r, err := strconv.ParseFloat(R, 64)
		if err != nil {
			return "", err
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

		if va <= 0.0 {
			va = 0.0
		}

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
			params.SepConfig.TCR.BusinUnitCode,
			InvOrdNum,
			IssueDateTime.Year(),
			params.SepConfig.TCR.TCRCode,
		),
	)

	Invoice := &sep.Invoice{
		TypeOfInv:       sep.TypeOfInv(TypeOfInv),
		IsSimplifiedInv: false,
		IssueDateTime:   sep.DateTime(IssueDateTime),
		InvNum:          InvNum,
		InvOrdNum:       sep.InvOrdNum(InvOrdNum),
		TCRCode:         sep.TCRCode(params.SepConfig.TCR.TCRCode),
		IsIssuerInVAT:   IsIssuerInVAT,
		TotPriceWoVAT:   sep.Amount(TotPriceWoVAT),
		TotVATAmt:       sep.Amount(TotVATAmt),
		TotPrice:        sep.Amount(TotPrice),
		OperatorCode:    sep.OperatorCode(params.SepConfig.OperatorCode),
		BusinUnitCode:   sep.BusinUnitCode(params.SepConfig.TCR.BusinUnitCode),
		SoftCode:        sep.SoftCode(params.SepConfig.TCR.SoftCode),
		IsReverseCharge: false,
		PayMethods:      PayMethods,
		Currency:        Currency,
		Seller:          Seller,
		Buyer:           Buyer,
		Items:           &Items,
		SameTaxes:       &SameTaxes,
		// CorrectiveInv: 	 &sep.CorrectiveInv{
		// 	IICRef: 		IICRef,
		// 	IssueDateTime: 	CorrectiveInvIssueDateTime,
		// 	Type: 			sep.CORRECTIVE,
		// },
		SumInvIICRefs: 	 &SumInvIICRefs,
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
		return "", err
	}

	doc := etree.NewDocument()
	err = doc.ReadFromBytes(buf)
	if err != nil {
		return "", err
	}

	doc, err = Envelope(doc)
	if err != nil {
		return "", err
	}

	return InternalOrdNum, doc.WriteToFile(params.OutFile)
}