// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// contractsArchiveDir is the directory in the zip archive of all contracts.
const contractsArchiveDir = "scribe_data_contracts"

// contractExtensions are the file extensions of contracts in order of preference.
var contractExtensions = []string{".yaml", ".yml"}

// contractFile is a data contract that has been converted to JSON.
type contractFile struct {
	lang     string
	json     []byte
	modified time.Time
}

// MARK: YAML to JSON

// contractToJSON converts a YAML data contract into indented JSON.
// Keys keep their order from the YAML file, as numbered keys (1, 2, ...) set the order that values are shown in.
func contractToJSON(data []byte) ([]byte, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, err
	}

	var compact bytes.Buffer
	if err := writeJSONNode(&compact, &document); err != nil {
		return nil, err
	}

	var indented bytes.Buffer
	err := json.Indent(&indented, compact.Bytes(), "", "  ")
	return indented.Bytes(), err
}

// writeJSONNode writes a YAML node to the buffer as JSON.
func writeJSONNode(buf *bytes.Buffer, node *yaml.Node) error {
	switch node.Kind {
	case yaml.DocumentNode:
		return writeJSONNode(buf, node.Content[0])

	case yaml.MappingNode:
		buf.WriteByte('{')
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if key.Kind != yaml.ScalarNode {
				return fmt.Errorf("unsupported key that isn't a string on line %d", key.Line)
			}

			if i > 0 {
				buf.WriteByte(',')
			}
			writeJSONString(buf, key.Value)
			buf.WriteByte(':')
			if err := writeJSONNode(buf, value); err != nil {
				return err
			}
		}
		buf.WriteByte('}')

	case yaml.SequenceNode:
		buf.WriteByte('[')
		for i, item := range node.Content {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeJSONNode(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')

	case yaml.AliasNode:
		return writeJSONNode(buf, node.Alias)

	case yaml.ScalarNode:
		return writeJSONScalar(buf, node)

	default:
		// Note: An empty YAML file has no nodes.
		buf.WriteString("null")
	}

	return nil
}

// writeJSONScalar writes a YAML scalar to the buffer as a JSON null, boolean, number or string.
func writeJSONScalar(buf *bytes.Buffer, node *yaml.Node) error {
	var value any
	switch node.ShortTag() {
	case "!!null":
		value = nil

	case "!!bool", "!!int", "!!float":
		if err := node.Decode(&value); err != nil {
			return err
		}

	default:
		// Note: Strings and other scalars such as dates are kept as they're written in the YAML file.
		writeJSONString(buf, node.Value)
		return nil
	}

	return writeJSONValue(buf, value)
}

// writeJSONValue writes a value to the buffer as JSON without escaping HTML characters.
func writeJSONValue(buf *bytes.Buffer, value any) error {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return err
	}

	buf.Write(bytes.TrimSuffix(out.Bytes(), []byte("\n")))
	return nil
}

// writeJSONString writes a string to the buffer as JSON without escaping HTML characters.
func writeJSONString(buf *bytes.Buffer, value string) {
	// Note: Encoding a string as JSON can't fail.
	_ = writeJSONValue(buf, value)
}

// MARK: Contract Loading

// readContract reads a contract file and converts it to JSON.
func readContract(filePath, lang string, modified time.Time) (contractFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return contractFile{}, fmt.Errorf("could not read contract file for %s: %w", lang, err)
	}

	contractJSON, err := contractToJSON(data)
	if err != nil {
		return contractFile{}, fmt.Errorf("could not convert contract for %s: %w", lang, err)
	}

	return contractFile{lang: lang, json: contractJSON, modified: modified}, nil
}

// loadContract loads the contract for a language as JSON.
// It returns an error wrapping fs.ErrNotExist if there's no contract for the language.
func loadContract(contractsDir, lang string) (contractFile, error) {
	for _, ext := range contractExtensions {
		filePath := filepath.Join(contractsDir, lang+ext)
		if info, err := os.Stat(filePath); err == nil {
			return readContract(filePath, lang, info.ModTime())
		}
	}

	return contractFile{}, fmt.Errorf("no contract file found for %s: %w", lang, fs.ErrNotExist)
}

// loadAllContracts loads all contracts in a directory as JSON, sorted by language.
// Contracts that can't be read or converted are skipped.
// It returns an error wrapping fs.ErrNotExist if the directory is missing or has no contracts.
func loadAllContracts(contractsDir string) ([]contractFile, error) {
	files, err := os.ReadDir(contractsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read contracts directory: %w", err)
	}

	contracts := []contractFile{}
	loaded := make(map[string]bool)

	// Note: Files are sorted by name, so .yaml contracts are loaded before .yml ones for the same language.
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		lang := strings.TrimSuffix(file.Name(), ext)
		if file.IsDir() || (ext != ".yaml" && ext != ".yml") || loaded[lang] {
			continue
		}

		filePath := filepath.Join(contractsDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Warning: could not read contract file %s: %v", file.Name(), err)
			continue
		}

		contract, err := readContract(filePath, lang, info.ModTime())
		if err != nil {
			log.Printf("Warning: %v", err)
			continue
		}

		contracts = append(contracts, contract)
		loaded[lang] = true
	}

	if len(contracts) == 0 {
		return nil, fmt.Errorf("no contract files found in %s: %w", contractsDir, fs.ErrNotExist)
	}

	return contracts, nil
}

// MARK: Contracts Archive

// writeContractsZip writes a zip archive with a scribe_data_contracts directory of JSON contracts.
// It's a package-level variable to allow mocking in unit tests.
var writeContractsZip = func(w io.Writer, contracts []contractFile) error {
	archive := zip.NewWriter(w)
	for _, contract := range contracts {
		entry, err := archive.CreateHeader(&zip.FileHeader{
			Name:     contractsArchiveDir + "/" + contract.lang + ".json",
			Method:   zip.Deflate,
			Modified: contract.modified,
		})
		if err == nil {
			_, err = entry.Write(contract.json)
		}
		if err != nil {
			return err
		}
	}

	return archive.Close()
}
