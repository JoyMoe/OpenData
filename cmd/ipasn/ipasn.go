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
	"github.com/maxmind/mmdbwriter/inserter"
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

		start, end, asn, name := parseRow(record)
		if start == nil || end == nil {
			log.Printf("Invalid IP range: %s - %s %s", record[0], record[1], record[2])
			continue
		}

		if asn == 0 {
			continue
		}

		if err := writer.InsertRangeFunc(start, end, inserter.TopLevelMergeWith(mmdbtype.Map{
			"asn":  mmdbtype.Uint32(asn),
			"name": mmdbtype.String(name),
		})); err != nil {
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

func parseRow(record []string) (net.IP, net.IP, uint32, string) {
	start := net.ParseIP(record[0])
	end := net.ParseIP(record[1])
	asn, err := strconv.ParseUint(record[2], 10, 32)

	if start == nil || end == nil || err != nil {
		return nil, nil, 0, ""
	}

	if asn == 0 {
		return start, end, 0, ""
	}

	// map 6to4 address back to IPv4
	// if start[0] == 0x20 && start[1] == 0x02 {
	// 	start = net.IPv4(start[2], start[3], start[4], start[5]).To4()
	// 	end = net.IPv4(end[2], end[3], end[4], end[5]).To4()
	// }

	name := record[4]

	return start, end, uint32(asn), name
}
