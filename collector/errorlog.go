package collector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"os/exec"
	"strconv"
	"strings"
)

type errLogMessageCollector struct {
	errLog *prometheus.Desc

	logger log.Logger
}

func init() {
	registerCollector("errlog", defaultEnabled, NewErrLogMessageCollector)
}

func NewErrLogMessageCollector(logger log.Logger) (Collector, error) {
	e := &errLogMessageCollector{
		errLog: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "errlog", "err_log"),
			"collector from /var/log/message failed and error counter",
			[]string{"app"},
			nil,
		),
		logger: logger,
	}
	return e, nil
}

func (e *errLogMessageCollector) Update(ch chan<- prometheus.Metric) error {
	cmd := exec.Command(
		"bash",
		"-c",
		"egrep -i 'failed|error' /var/log/messages | awk '{a[$5]++} END {for(i in a) print i,a[i]}'",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		level.Error(e.logger).Log("err", err)
		return err
	}

	e.parseErrLogOutPut(string(output), ch)
	return nil
}

func (e *errLogMessageCollector) parseErrLogOutPut(output string, ch chan<- prometheus.Metric) prometheus.Metric {
	for _, out := range strings.Split(output, "\n") {
		outs := strings.Fields(out)
		if len(outs) != 2 {
			continue
		}

		app, cnt := outs[0], outs[1]

		tapp := strings.ReplaceAll(strings.ReplaceAll(app, ":", ""), "-", "_")
		i, err := strconv.ParseFloat(cnt, 64)
		if err != nil {
			i = 0
		}

		ch <- prometheus.MustNewConstMetric(
			e.errLog,
			prometheus.CounterValue,
			i,
			tapp,
		)
	}
	return nil
}
