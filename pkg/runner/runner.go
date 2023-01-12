package runner

import (
	"bufio"
	"context"
	"io"
	"os"
	"path"
	"strings"

	"github.com/ZhuriLab/Starmap/pkg/passive"
	"github.com/ZhuriLab/Starmap/pkg/resolve"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
)

// Runner is an instance of the subdomain enumeration
// client used to orchestrate the whole process.
type Runner struct {
	options        *Options
	passiveAgent   *passive.Agent
	resolverClient *resolve.Resolver
	Resolvers      []string
}

// NewRunner creates a new runner struct instance by parsing
// the configuration options, configuring sources, reading lists
// and setting up loggers, etc.
func NewRunner(options *Options) (*Runner, error) {
	runner := &Runner{options: options}

	// Initialize the passive subdomain enumeration engine
	runner.initializePassiveEngine()

	// Initialize the active subdomain enumeration engine
	err := runner.initializeActiveEngine()
	if err != nil {
		return nil, err
	}

	return runner, nil
}

// RunEnumeration runs the subdomain enumeration flow on the targets specified
func (r *Runner) RunEnumeration(ctx context.Context) error {
	outputs := []io.Writer{r.options.Output}

	if len(r.options.Domain) > 0 {
		domainsReader := strings.NewReader(strings.Join(r.options.Domain, "\n"))
		return r.EnumerateMultipleDomains(ctx, domainsReader, outputs)
	}

	// If we have multiple domains as input,
	if r.options.DomainsFile != "" {
		f, err := os.Open(r.options.DomainsFile)
		if err != nil {
			return err
		}
		err = r.EnumerateMultipleDomains(ctx, f, outputs)
		f.Close()
		return err
	}

	// If we have STDIN input, treat it as multiple domains
	if r.options.Stdin {
		return r.EnumerateMultipleDomains(ctx, os.Stdin, outputs)
	}
	return nil
}

//	enumerates subdomains for multiple domains
//
// We keep enumerating subdomains for a given domain until we reach an error
func (r *Runner) EnumerateMultipleDomains(ctx context.Context, reader io.Reader, outputs []io.Writer) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		domain, err := sanitize(scanner.Text())
		if errors.Is(err, ErrEmptyInput) {
			continue
		}

		var file *os.File
		// If the user has specified an output file, use that output file instead
		// of creating a new output file for each domain. Else create a new file
		// for each domain in the directory.
		if r.options.OutputFile != "" {
			outputter := NewOutputter(r.options.JSON)
			file, err = outputter.createFile(r.options.OutputFile, true)
			if err != nil {
				gologger.Error().Msgf("Could not create file %s for %s: %s\n", r.options.OutputFile, r.options.Domain, err)
				return err
			}

			err, _, _ = r.EnumerateSingleDomain(ctx, domain, append(outputs, file))

			file.Close()
		} else if r.options.OutputDirectory != "" {
			outputFile := path.Join(r.options.OutputDirectory, domain)
			if r.options.JSON {
				outputFile += ".json"
			} else {
				outputFile += ".txt"
			}

			outputter := NewOutputter(r.options.JSON)
			file, err = outputter.createFile(outputFile, false)
			if err != nil {
				gologger.Error().Msgf("Could not create file %s for %s: %s\n", r.options.OutputFile, r.options.Domain, err)
				return err
			}

			err, _, _ = r.EnumerateSingleDomain(ctx, domain, append(outputs, file))

			file.Close()
		} else {
			err, _, _ = r.EnumerateSingleDomain(ctx, domain, outputs)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
