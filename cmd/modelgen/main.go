package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ovn-org/libovsdb/modelgen"
	"github.com/ovn-org/libovsdb/ovsdb"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of modelgen:\n")
	fmt.Fprintf(os.Stderr, "\tmodelgen [flags] OVS_SCHEMA\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

var (
	outDirP  = flag.String("o", ".", "Directory where the generated files shall be stored")
	pkgNameP = flag.String("p", "ovsmodel", "Package name")
	dryRun   = flag.Bool("d", false, "Dry run")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("modelgen: ")
	flag.Usage = usage
	flag.Parse()
	outDir := *outDirP
	pkgName := *pkgNameP

	/*Option handling*/
	outDir, err := filepath.Abs(outDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}

	schemaFile, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	defer schemaFile.Close()

	schemaBytes, err := ioutil.ReadAll(schemaFile)
	if err != nil {
		log.Fatal(err)
	}

	var dbSchema ovsdb.DatabaseSchema
	if err := json.Unmarshal(schemaBytes, &dbSchema); err != nil {
		log.Fatal(err)
	}

	genOpts := []modelgen.Option{}
	if *dryRun {
		genOpts = append(genOpts, modelgen.WithDryRun())
	}
	gen, err := modelgen.NewGenerator(genOpts...)
	if err != nil {
		log.Fatal(err)
	}
	for name, table := range dbSchema.Tables {
		tmpl := modelgen.NewTableTemplate()
		args := modelgen.GetTableTemplateData(pkgName, name, &table)
		if err := gen.Generate(filepath.Join(outDir, modelgen.FileName(name)), tmpl, args); err != nil {
			log.Fatal(err)
		}
	}
	dbTemplate := modelgen.NewDBTemplate()
	dbArgs := modelgen.GetDBTemplateData(pkgName, dbSchema)
	if err := gen.Generate(filepath.Join(outDir, "model.go"), dbTemplate, dbArgs); err != nil {
		log.Fatal(err)
	}
}
