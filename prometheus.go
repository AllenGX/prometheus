package Prometheus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var requestCount *kitprometheus.Counter
var requestLatency *kitprometheus.Summary

func init() {
	fieldKeys := []string{"method", "error"}
	requestCount = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

}

func WrapEndpoint(in interface{}) interface{} {
	// range members with reflect
	prometheusWrap := func(ep endpoint.Endpoint, epName string) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (rep interface{}, err error) {
			defer func(begin time.Time) {
				lvs := []string{"method", epName, "error", fmt.Sprint(err != nil)}
				requestCount.With(lvs...).Add(1)
				requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
			}(time.Now())
			resp, err := ep(ctx, req)
			return resp, err
		}
	}
	vIn := reflect.ValueOf(in)
	newOne := reflect.New(reflect.TypeOf(in))
	vOut := reflect.Indirect(newOne)
	for i := 0; i < vIn.NumField(); i++ {
		endpoint := vIn.Field(i).Interface().(endpoint.Endpoint)
		endpoint = prometheusWrap(endpoint, vIn.Type().Field(i).Name)
		//logs.Info(vIn.Type().Field(i).Name + " is ok")
		vOut.Field(i).Set(reflect.ValueOf(endpoint))
	}
	return newOne.Elem().Interface()
}

func Run(addr string) {
	log.Println("transport", "prometheus", "addr", addr)
	http.ListenAndServe(addr, promhttp.Handler())
}
