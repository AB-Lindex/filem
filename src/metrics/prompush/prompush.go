package prompush

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/AB-Lindex/filem/src/metrics"
	"github.com/rs/zerolog/log"
)

// Settings contain the settings and stored metric-values
type Settings struct {
	URL      string            `yaml:"url"` // http://PUSHGATEWAY_FQDN
	Job      string            `yaml:"job"`
	Instance string            `yaml:"instance"`
	Labels   map[string]string `yaml:"labels"`
	Vars     []string          `yaml:"vars"`

	values map[string]interface{}
}

var (
	errDatatypeNotSupported = fmt.Errorf("datatype not supported")
	errNoSlashes            = fmt.Errorf("slashes or spaces not allowed in metric-values")
)

// GetHandler returns the metric.Target for the pushgateway-handler
func (t *Settings) GetHandler() (metrics.Target, error) {

	if strings.ContainsAny(t.Job, "\\/ ") ||
		strings.ContainsAny(t.Instance, "\\/ ") {
		return nil, errNoSlashes
	}

	if t.Job == "" {
		t.Job = "filem"
	}

	if t.Instance == "" {
		t.Instance, _ = os.Hostname()
		t.Instance = strings.ToLower(t.Instance)
	}
	if t.Instance == "" {
		t.Instance = "default"
	}

	return t, nil
}

func safelabel(k string) string {
	k = strings.ToLower(k)
	result := make([]rune, 0, len(k))
	for _, ch := range k {
		switch true {
		case ch >= 'a' && ch <= 'z':
			result = append(result, ch)
		default:
			result = append(result, '_')
		}
	}
	return string(result)
}

func makeLabel(name string, keys map[string]interface{}) string {
	if len(keys) == 0 {
		return name
	}

	var buf = &bytes.Buffer{}
	buf.WriteString(name)
	buf.WriteRune('{')

	i := 0
	for k, v := range keys {
		if i > 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(safelabel(k))
		buf.WriteString("=\"")
		buf.WriteString(fmt.Sprint(v))
		buf.WriteString("\"")
		i++
	}

	buf.WriteRune('}')
	return buf.String()
}

// Set creates a metric+keys with a value and saves
func (t *Settings) Set(name string, value interface{}, keys map[string]interface{}) error {

	switch value.(type) {
	case int:
	case int64:
	case float64:
	default:
		return errDatatypeNotSupported
	}

	if t.values == nil {
		t.values = make(map[string]interface{})
	}
	t.values[makeLabel(name, keys)] = value
	return nil
}

// Send sends all metrics collected so far
func (t *Settings) Send() error {

	var body = &bytes.Buffer{}

	var keys = make([]string, len(t.values))
	for k := range t.values {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool { return strings.Compare(keys[i], keys[j]) < 0 })

	var lastkey string
	for _, k := range keys {
		var field = k
		p := strings.IndexRune(k, '{')
		if p >= 0 {
			field = k[:p]
		}
		if field != lastkey {
			lastkey = field
			fmt.Fprintf(body, "# TYPE %s gauge\n", field)
		}

		switch vv := t.values[k].(type) {
		case int, int64:
			fmt.Fprintf(body, "%s %d\n", k, vv)
		case float64:
			fmt.Fprintf(body, "%s %f\n", k, vv)
		}
	}

	buf := body.Bytes()
	body = bytes.NewBuffer(buf) // reset position

	url := fmt.Sprintf("%s/metrics/job/%s/instance/%s", t.URL, t.Job, t.Instance)
	for k, v := range t.Labels {
		url += fmt.Sprintf("/%s/%s", k, v)
	}
	log.Debug().Msgf("Pushgateway url: %s", url)
	fmt.Println(string(buf))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("prompush-error: statuscode (%s) != ok\n%s", resp.Status, string(buf))
	}

	return nil
}
