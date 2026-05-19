package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/andreagrandi/logbasset/internal/errors"
)

func outputJSON(data any, pretty bool) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		errors.HandleErrorAndExit(errors.NewParseError("failed to marshal JSON", err))
	}

	fmt.Println(string(output))
}

func outputNumericCSV(values []float64) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	record := make([]string, len(values))
	for i, val := range values {
		record[i] = strconv.FormatFloat(val, 'f', -1, 64)
	}
	writer.Write(record)
}
