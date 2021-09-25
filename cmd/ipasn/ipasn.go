// Copyright (c) 2021 JoyMoe Interactive Entertainment Limited
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/csv"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

func main() {
	writer, err := mmdbwriter.New(mmdbwriter.Options{
		RecordSize: 24,
	})
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("data/ip2asn.tsv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			continue
		}

		start := net.ParseIP(record[0])
		end := net.ParseIP(record[1])
		asn, err := strconv.ParseUint(record[2], 10, 32)

		if start == nil || end == nil || err != nil {
			log.Printf("Invalid IP range: %s - %s %s", record[0], record[1], record[2])
			continue
		}

		if asn == 0 {
			continue
		}

		name := record[4]

		if err := writer.InsertRange(start, end, mmdbtype.Map{
			"asn":  mmdbtype.Uint32(asn),
			"name": mmdbtype.String(name),
		}); err != nil {
			log.Printf("%s - %s %s", record[0], record[1], record[2])
			log.Println(err)
			continue
		}
	}

	fh, err := os.Create("data/ipasn.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.WriteTo(fh)
	if err != nil {
		log.Fatal(err)
	}
}
