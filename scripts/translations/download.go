package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/ioutil"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/syncutil"
)

const (
	hostlistRegistryProjectID = "hostlists-registry"
	jsonMessageKey            = "message"
	jsonIndentPrefix          = ""
	jsonIndentString          = "    "
)

// download and save all translations.
func (c *twoskyClient) download(ctx context.Context, l *slog.Logger) (err error) {
	var numWorker int

	flagSet := flag.NewFlagSet("download", flag.ExitOnError)
	flagSet.Usage = func() {
		usage("download command error")
	}
	flagSet.IntVar(&numWorker, "n", 1, "number of concurrent downloads")

	err = flagSet.Parse(os.Args[2:])
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	}

	if numWorker < 1 {
		usage("count must be positive")
	}

	// download locales from AdGuard Home crowdin project
	if err = c.downloadTo(ctx, l, localesDir, defaultBaseFile, numWorker); err != nil {
		return err
	}

	// download services from AdGuard Hostlist Registry crowdin project
	c.projectID = hostlistRegistryProjectID
	if err = c.downloadTo(ctx, l, servicesLocalesDir, servicesBaseFile, numWorker); err != nil {
		return err
	}

	return nil
}

// downloads and saves all translations to the specified outputDir
// using the specified baseFile name.
func (c *twoskyClient) downloadTo(
	ctx context.Context,
	l *slog.Logger,
	outputDir string,
	baseFile string,
	numWorker int,
) (err error) {
	downloadURI := c.uri.JoinPath("download")

	wg := &sync.WaitGroup{}
	uriCh := make(chan *url.URL, len(c.langs))

	dw := &downloadWorker{
		ctx:    ctx,
		l:      l,
		failed: syncutil.NewMap[string, struct{}](),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		uriCh:     uriCh,
		outputDir: outputDir,
		baseFile:  baseFile,
	}

	for range numWorker {
		wg.Go(dw.run)
	}

	for _, lang := range c.langs {
		uri := translationURL(downloadURI, baseFile, c.projectID, lang)

		uriCh <- uri
	}

	close(uriCh)
	wg.Wait()

	printFailedLocales(ctx, l, dw.failed)

	return nil
}

// extracts the "message" values into a flat map and returns the result
// formatted with consistent JSON indentation
func normalizeServicesJSON(data []byte) ([]byte, error) {
	unwrapped, err := extractMessagesJSON(data)
	if err != nil {
		return nil, fmt.Errorf("normalize services json: extraction failed: %w", err)
	}

	indented, err := indentJSON(unwrapped)
	if err != nil {
		return nil, fmt.Errorf("normalize services json: indentation failed: %w", err)
	}

	return indented, nil
}

// converts a wrapped services JSON of shape
// {"key": {"message": "..."}} into a flat {"key": "..."}
func extractMessagesJSON(input []byte) ([]byte, error) {
	var wrapped map[string]map[string]any
	if err := json.Unmarshal(input, &wrapped); err != nil {
		return nil, fmt.Errorf("extract json: unmarshal wrapped payload: %w", err)
	}

	flattened := make(map[string]string, len(wrapped))
	for key, inner := range wrapped {
		rawValue, ok := inner[jsonMessageKey]
		if !ok {
			return nil, fmt.Errorf("extract json: missing %q field for key %q", jsonMessageKey, key)
		}

		message, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("extract json: %q field for key %q is not a string", jsonMessageKey, key)
		}

		flattened[key] = message
	}

	result, err := json.Marshal(flattened)
	if err != nil {
		return nil, fmt.Errorf("extract json: marshal flattened map: %w", err)
	}

	return result, nil
}

// formats JSON using the configured prefix and indent
func indentJSON(data []byte) (b []byte, err error) {
	var buffer bytes.Buffer

	err = json.Indent(&buffer, data, jsonIndentPrefix, jsonIndentString)
	if err != nil {
		return nil, fmt.Errorf("indent json: formatting failed: %w", err)
	}

	return buffer.Bytes(), nil
}

func printFailedLocales(
	ctx context.Context,
	l *slog.Logger,
	failed *syncutil.Map[string, struct{}],
) {
	var keys []string
	for k := range failed.Range {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return
	}

	slices.Sort(keys)

	l.InfoContext(ctx, "failed", "locales", keys)
}

// downloadWorker is a worker for downloading translations.  It uses URLs
// received from the channel to download translations and save them to files.
// Failures are stored in the failed map.  All fields must not be nil.
type downloadWorker struct {
	ctx       context.Context
	l         *slog.Logger
	failed    *syncutil.Map[string, struct{}]
	client    *http.Client
	uriCh     <-chan *url.URL
	outputDir string
	baseFile  string
}

// run handles the channel of URLs, one by one.  It returns when the channel is
// closed.  It's used to be run in a separate goroutine.
func (w *downloadWorker) run() {
	for uri := range w.uriCh {
		q := uri.Query()
		code := q.Get("language")

		err := saveToFile(w.ctx, w.l, w.client, uri, code, w.outputDir, w.baseFile)
		if err != nil {
			w.l.ErrorContext(w.ctx, "download worker", slogutil.KeyError, err)
			w.failed.Store(code, struct{}{})
		}
	}
}

// saveToFile downloads translation by url and saves it to a file, or returns
// error.
func saveToFile(
	ctx context.Context,
	l *slog.Logger,
	client *http.Client,
	uri *url.URL,
	code string,
	outputDir string,
	baseFile string,
) (err error) {
	data, err := getTranslation(ctx, l, client, uri.String())
	if err != nil {
		return fmt.Errorf("getting translation %q: %s", code, err)
	}

	if baseFile == servicesBaseFile {
		data, err = normalizeServicesJSON(data)
		if err != nil {
			return fmt.Errorf("normalize services JSON for %q: %w", code, err)
		}
	}

	name := filepath.Join(outputDir, code+".json")
	err = os.WriteFile(name, data, 0o664)
	if err != nil {
		return fmt.Errorf("writing file: %s", err)
	}

	fmt.Println(name)

	return nil
}

// getTranslation returns received translation data and error.  If err is not
// nil, data may contain a response from server for inspection.
func getTranslation(
	ctx context.Context,
	l *slog.Logger,
	client *http.Client,
	url string,
) (data []byte, err error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("requesting: %w", err)
	}

	defer slogutil.CloseAndLog(ctx, l, resp.Body, slog.LevelError)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("url: %q; status code: %s", url, http.StatusText(resp.StatusCode))

		// Go on and download the body for inspection.
	}

	limitReader := ioutil.LimitReader(resp.Body, readLimit)

	data, readErr := io.ReadAll(limitReader)

	return data, errors.WithDeferred(err, readErr)
}

// translationURL returns a new url.URL with provided query parameters.
func translationURL(oldURL *url.URL, baseFile, projectID string, lang langCode) (uri *url.URL) {
	uri = &url.URL{}
	*uri = *oldURL

	q := uri.Query()
	q.Set("format", "json")
	q.Set("filename", baseFile)
	q.Set("project", projectID)
	q.Set("language", string(lang))

	uri.RawQuery = q.Encode()

	return uri
}
