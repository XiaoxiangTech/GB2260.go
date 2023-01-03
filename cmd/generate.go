package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func getNameSuffix(currentSource string) string {
	ext := path.Ext(currentSource)
	pathData := currentSource[0 : len(currentSource)-len(ext)]
	splits := strings.SplitN(pathData, "-", 2)
	if len(splits) == 2 {
		return splits[1]
	}

	return ""
}

func generateFileData(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("error while close file, filename=%s, err=%+v\n", fileName, err)
		}
	}()

	divisions := "map[string]string{\n"
	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = 4
	csvReader.Comma = '\t'

	const (
		codeIndex = 2
		nameIndex = 3
	)

	for lineNum := 1; ; lineNum++ {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if lineNum == 1 {
			// skip tsv header
			continue
		}
		code := line[codeIndex]
		name := line[nameIndex]

		divisions += fmt.Sprintf(`"%s": "%s",`, code, name) + "\n"
	}
	divisions += "}"

	return divisions, nil
}

// before use it, you should use go build and then execute the generate (or generate.exe)
func main() {
	dir := "../data/mca/"
	names := []int{201904, 202112}

	gocode := `package gb2260
var (
	Divisions map[string]map[string]string
)

func init(){
	Divisions = map[string]map[string]string{
		%s
	}
}
`

	code := ""

	for _, n := range names {
		fileName := dir + fmt.Sprintf("%d.tsv", n)

		data, err := generateFileData(fileName)
		if err != nil {
			log.Fatal(err)
			return
		}

		code += fmt.Sprintf("\"%d\": %s,\n", n, data)
	}

	gocode = fmt.Sprintf(gocode, code)
	err := ioutil.WriteFile("../data.go", []byte(gocode), os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}

	cmd := exec.Command("go", "fmt", "../data.go")
	cmd.Start()

	return
}
