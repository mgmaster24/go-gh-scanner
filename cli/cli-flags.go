package cli

import (
	"flag"
	"fmt"
	"os"
)

type ScannerFlags struct {
	AppConfig    string
	TokensConfig string
}

func InitFlags() *ScannerFlags {
	flags := &ScannerFlags{}
	flag.StringVar(&flags.AppConfig, "c", "config.json", "Provides the location of the application configuration.")
	flag.StringVar(&flags.TokensConfig, "t", "tokens.json", "Provides the location of the tokens configuration.")

	flag.Usage = func() {
		fmt.Fprint(
			os.Stderr,
			"Provide the application config and token config.\n",
			"If not provided the default version will be used.\n",
			"Usage of go-gh-scanner:\n",
			"\tgo-gh-scanner -c path-to-config.json -t path-to-token.json\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	return flags
}
