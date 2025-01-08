package reports

import (
	"encoding/csv"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCreateCsv(t *testing.T) {
	want, _ := os.Create("test.csv")
	defer want.Close()

	writer := csv.NewWriter(want)
	_ = writer.Write([]string{"test", "hehe"})
	_ = writer.Write([]string{"123 Real Street", "Bingopolis"})
	writer.Flush()

	items := [][]string{{"test", "hehe"}, {"123 Real Street", "Bingopolis"}}
	_, err := createCsv("test2.csv", items)

	wantBytes, _ := os.ReadFile("test.csv")
	gotBytes, _ := os.ReadFile("test2.csv")

	assert.Nil(t, err)
	assert.Equal(t, string(wantBytes), string(gotBytes))
}

func TestCreateCsvNoItems(t *testing.T) {
	items := [][]string{}
	_, err := createCsv("test.csv", items)
	gotBytes, _ := os.ReadFile("test.csv")

	assert.Nil(t, err)
	assert.Equal(t, "", string(gotBytes))
}
