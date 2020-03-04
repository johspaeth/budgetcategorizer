package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	budget "github.com/jbleduigou/budgetcategorizer"
)

// Parser provides an interface for parsing raw csv transactions
type Parser interface {
	ParseTransactions(r io.Reader) ([]budget.Transaction, error)
}

// NewParser will provide an instance of a Parser, implementation is not exposed
func NewParser() Parser {
	return &csvParser{}
}

type csvreader interface {
	ReadAll() (records [][]string, err error)
}

type csvParser struct {
}

func (c *csvParser) ParseTransactions(r io.Reader) ([]budget.Transaction, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.Comma = ';'
	reader.FieldsPerRecord = 4
	reader.FieldsPerRecord = -1
	return c.parse(reader)
}

func (c *csvParser) parse(reader csvreader) (transactions []budget.Transaction, err error) {
	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	for _, each := range rawCSVdata {
		if len(each) == 5 {
			date := each[0]
			libelle := c.sanitizeDescription(each[1])
			if c.isDebitTransaction(each) {
				debit, err := c.parseAmount(each[2])
				if err == nil {
					t := budget.NewTransaction(date, libelle, "", "Courses Alimentation", debit)
					transactions = append(transactions, t)
				} else {
					fmt.Printf("%v\n", err)
				}
			} else {
				credit, err := c.parseAmount(each[3])
				if err == nil {
					t := budget.NewTransaction(date, libelle, "", "", -credit)
					transactions = append(transactions, t)
				} else {
					fmt.Printf("%v\n", err)
				}
			}
		}
	}
	fmt.Printf("Found %v transactions\n", len(transactions))
	return transactions, nil
}

func (c *csvParser) isDebitTransaction(d []string) bool {
	if d[2] == "" && d[3] != "" {
		return false
	}
	return true
}

func (c *csvParser) sanitizeDescription(d string) string {
	capitalized := strings.Title(strings.ToLower(d))
	libelle := []byte(capitalized)
	{
		re := regexp.MustCompile(`\n`)
		libelle = re.ReplaceAll(libelle, []byte(" "))
	}
	{
		re := regexp.MustCompile(`[\s]+`)
		libelle = re.ReplaceAll(libelle, []byte(" "))
	}
	{
		re := regexp.MustCompile(`[^\x20-\x7F]`)
		libelle = re.ReplaceAll(libelle, []byte(""))
	}
	if strings.Contains(capitalized, "Cheque Emis") {
		return c.sanitizeCheque(libelle)
	}
	return string(libelle)
}

func (c *csvParser) sanitizeCheque(libelle []byte) string {
	{
		re := regexp.MustCompile(`[\/][0]+`)
		libelle = re.ReplaceAll(libelle, []byte(""))
	}
	{
		re := regexp.MustCompile(`([\d]{7,7})( )(Cheque Emis)( )`)
		libelle = re.ReplaceAll(libelle, []byte("$3 $1"))
	}
	return string(libelle)
}

func (c *csvParser) parseAmount(a string) (float64, error) {
	creditStr := []byte(a)
	{
		re := regexp.MustCompile(`,`)
		creditStr = re.ReplaceAll(creditStr, []byte("."))
	}
	credit, err := strconv.ParseFloat(string(creditStr), 64)
	return credit, err
}
