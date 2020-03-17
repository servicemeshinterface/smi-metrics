package prometheus

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/masterminds/sprig"
	promAPI "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

type Query struct {
	Template string
	Name     string
	Values   map[string]interface{}
}

type Lookup interface {
	Get(edge model.Metric) *metrics.TrafficMetrics
	Queries() []*Query
}

type Client struct {
	ctx      context.Context
	client   promAPI.API
	interval *metrics.Interval
}

func NewClient(
	ctx context.Context,
	client promAPI.API,
	interval *metrics.Interval) *Client {
	return &Client{
		ctx:      ctx,
		client:   client,
		interval: interval,
	}
}

func (c *Client) render(
	query string, opts map[string]interface{}) (string, error) {
	buf := &bytes.Buffer{}

	seconds := int(c.interval.Window.Duration.Seconds())
	opts["window"] = fmt.Sprintf("%ds", seconds)

	tmpl, err := template.New("query").Funcs(sprig.TxtFuncMap()).Parse(query)
	if err != nil {
		return "", err
	}

	if err := tmpl.Execute(buf, opts); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *Client) Execute(
	queryTemplate string, opts map[string]interface{}) (model.Vector, error) {

	query, err := c.render(queryTemplate, opts)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"query": query,
	}).Debug("querying prometheus")

	result, err := c.client.Query(c.ctx, query, c.interval.Timestamp.Time)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"query":  query,
		"result": result.String(),
	}).Debug("query results")

	return result.(model.Vector), nil
}

func (c *Client) Update(lst Lookup) error {

	for _, query := range lst.Queries() {
		result, err := c.Execute(query.Template, query.Values)
		if err != nil {
			return err
		}

		for _, sample := range result {
			resource := lst.Get(sample.Metric)
			metric := resource.Get(query.Name)
			if metric.Name == "" {
				continue
			}

			metric.Value = apiresource.NewMilliQuantity(
				int64(sample.Value*1000), apiresource.DecimalSI)
		}
	}

	return nil
}
