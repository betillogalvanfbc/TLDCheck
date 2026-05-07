package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

const ianaTLDURL = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"
const publicSuffixURL = "https://publicsuffix.org/list/public_suffix_list.dat"

func normalizeSuffix(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, ".")

	// La PSL usa reglas como *.ck o !www.ck.
	// Para generar dominios brutos, normalmente no quieres esos prefijos.
	s = strings.TrimPrefix(s, "*.")
	s = strings.TrimPrefix(s, "!")

	return s
}

func fetchList(url string, isPublicSuffixList bool) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status: %s", url, resp.Status)
	}

	seen := map[string]bool{}
	var suffixes []string

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if isPublicSuffixList {
			if strings.HasPrefix(line, "//") {
				continue
			}
		} else {
			if strings.HasPrefix(line, "#") {
				continue
			}
		}

		suffix := normalizeSuffix(line)
		if suffix == "" || seen[suffix] {
			continue
		}

		seen[suffix] = true
		suffixes = append(suffixes, suffix)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return suffixes, nil
}

func uniqueSuffixes(groups ...[]string) []string {
	seen := map[string]bool{}
	var result []string

	for _, group := range groups {
		for _, raw := range group {
			suffix := normalizeSuffix(raw)
			if suffix == "" || seen[suffix] {
				continue
			}

			seen[suffix] = true
			result = append(result, suffix)
		}
	}

	sort.Strings(result)
	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No name provided")
		return
	}

	domain := os.Args[1]

	ianaTLDs, err := fetchList(ianaTLDURL, false)
	if err != nil {
		fmt.Println("Error fetching IANA TLDs:", err)
		return
	}

	publicSuffixes, err := fetchList(publicSuffixURL, true)
	if err != nil {
		fmt.Println("Error fetching Public Suffix List:", err)
		return
	}

	// Aquí puedes conservar tus TLDs alternativos/no oficiales.
	customExtras := []string{
		"0db",
		"0z",
		"3dom",
		"4free",
		"web3",
		"crypto",
		"wallet",
		"blockchain",
	}

	suffixes := uniqueSuffixes(ianaTLDs, publicSuffixes, customExtras)

	for _, suffix := range suffixes {
		fmt.Println(domain + "." + suffix)
	}
}
