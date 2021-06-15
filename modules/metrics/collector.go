// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package metrics

import (
	"code.gitea.io/gitea/models"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "gitea_"

// Collector implements the prometheus.Collector interface and
// exposes gitea metrics for prometheus
type Collector struct {
	Follows       *prometheus.Desc
	LoginSources  *prometheus.Desc
	Oauths        *prometheus.Desc
	Organizations *prometheus.Desc
	Teams         *prometheus.Desc
	UpdateTasks   *prometheus.Desc
	Users         *prometheus.Desc
}

// NewCollector returns a new Collector with all prometheus.Desc initialized
func NewCollector() Collector {
	return Collector{
		Follows: prometheus.NewDesc(
			namespace+"follows",
			"Number of Follows",
			nil, nil,
		),
		LoginSources: prometheus.NewDesc(
			namespace+"loginsources",
			"Number of LoginSources",
			nil, nil,
		),
		Oauths: prometheus.NewDesc(
			namespace+"oauths",
			"Number of Oauths",
			nil, nil,
		),
		Organizations: prometheus.NewDesc(
			namespace+"organizations",
			"Number of Organizations",
			nil, nil,
		),
		Teams: prometheus.NewDesc(
			namespace+"teams",
			"Number of Teams",
			nil, nil,
		),
		UpdateTasks: prometheus.NewDesc(
			namespace+"updatetasks",
			"Number of UpdateTasks",
			nil, nil,
		),
		Users: prometheus.NewDesc(
			namespace+"users",
			"Number of Users",
			nil, nil,
		),
	}

}

// Describe returns all possible prometheus.Desc
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Follows
	ch <- c.LoginSources
	ch <- c.Oauths
	ch <- c.Organizations
	ch <- c.Teams
	ch <- c.UpdateTasks
	ch <- c.Users
}

// Collect returns the metrics with values
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	stats := models.GetStatistic()

	ch <- prometheus.MustNewConstMetric(
		c.Follows,
		prometheus.GaugeValue,
		float64(stats.Counter.Follow),
	)
	ch <- prometheus.MustNewConstMetric(
		c.LoginSources,
		prometheus.GaugeValue,
		float64(stats.Counter.LoginSource),
	)
	ch <- prometheus.MustNewConstMetric(
		c.Oauths,
		prometheus.GaugeValue,
		float64(stats.Counter.Oauth),
	)
	ch <- prometheus.MustNewConstMetric(
		c.Organizations,
		prometheus.GaugeValue,
		float64(stats.Counter.Org),
	)
	ch <- prometheus.MustNewConstMetric(
		c.Teams,
		prometheus.GaugeValue,
		float64(stats.Counter.Team),
	)
	ch <- prometheus.MustNewConstMetric(
		c.UpdateTasks,
		prometheus.GaugeValue,
		float64(stats.Counter.UpdateTask),
	)
	ch <- prometheus.MustNewConstMetric(
		c.Users,
		prometheus.GaugeValue,
		float64(stats.Counter.User),
	)
}
