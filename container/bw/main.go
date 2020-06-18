package main

import (
    "net/http"
    "time"
    "container/list"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/leobrada/golang_utility"

	"github.com/Shopify/sarama"
	kafka "github.com/bwNetFlow/kafkaconnector"
)

var (
    // Fix path in the docker container
    ini_path = "./settings.ini"
    kafka_conn = kafka.Connector{}
    bytes_down, bytes_up, bandwidth_down, bandwidth_up uint64
    data_points_down *list.List = list.New()
    data_points_up *list.List = list.New()
    current_bandwidth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "bw",
            Help: "The Current Bandwidth in the Specified Network Based On Peered Flows By The BelWü",
        },
        []string{"flow_direction"},
    )
    current_bandwidth_down prometheus.Gauge = current_bandwidth.WithLabelValues("down")
    current_bandwidth_up prometheus.Gauge = current_bandwidth.WithLabelValues("up")
)

func timeout(seconds time.Duration, ch chan<- bool) {
	timer := time.NewTimer(seconds * time.Second)
	<-timer.C
	ch <- true
}

func setCurrBandwidthsDown(bandwidth_over_datapoints float64) {
    current_bandwidth_down.Set(bandwidth_over_datapoints)
}

func setCurrBandwidthsUp(bandwidth_over_datapoints float64) {
    current_bandwidth_up.Set(bandwidth_over_datapoints)
}

func calcAvgBandwidthOverDatapointsDown() {
    var data_points_sum uint64
    for e := data_points_down.Front(); e != nil; e = e.Next() {
        data_points_sum += e.Value.(uint64)
    }
    bandwidth_over_datapoints := data_points_sum / uint64(data_points_down.Len())

    setCurrBandwidthsDown(float64(bandwidth_over_datapoints))
}

func calcAvgBandwidthOverDatapointsUp() {
    var data_points_sum uint64
    for e := data_points_up.Front(); e != nil; e = e.Next() {
        data_points_sum += e.Value.(uint64)
    }
    bandwidth_over_datapoints := data_points_sum / uint64(data_points_up.Len())

    setCurrBandwidthsUp(float64(bandwidth_over_datapoints))
}

func calcAvgBandwidthDown() {
	bandwidth_down := (bytes_down * 8) / 1024 / 1024
    bytes_down = 0

    if(data_points_down.Len() >= 60) {
        data_points_down.Remove(data_points_down.Front())
    }

    data_points_down.PushBack(bandwidth_down)

    if(data_points_down.Len() < 60) {
        return
    }

    calcAvgBandwidthOverDatapointsDown()
}

func calcAvgBandwidthUp() {
	bandwidth_up := (bytes_up * 8) / 1024 / 1024
    bytes_up = 0

    if(data_points_up.Len() >= 60) {
        data_points_up.Remove(data_points_up.Front())
    }

    data_points_up.PushBack(bandwidth_up)

    if(data_points_up.Len() < 60) {
        return
    }

    calcAvgBandwidthOverDatapointsUp()
}

func calcAvgBandwidth() {
    calcAvgBandwidthDown()
    calcAvgBandwidthUp()
}

func gatherMoreBytes() {
	flow := <-kafka_conn.ConsumerChannel()
    if(flow.GetFlowDirection() == 0) {
	    bytes_down += flow.GetBytes()
    } else {
        bytes_up += flow.GetBytes()
    }
}

func consumeOldFlows() {
    flow := <-kafka_conn.ConsumerChannel()
    _ = flow
}

func bandwidthCalculation() {
	credentials, _ := reader.ReadIni(ini_path)

	kafka_conn.SetAuth(credentials["user"], credentials["pwd"])
	topic := []string{credentials["topic"]}
	kafka_conn.StartConsumer(credentials["brokers"], topic, credentials["bw_grp_id"], sarama.OffsetNewest)
	defer kafka_conn.Close()

    for i := 0; i < 100000; i++ {
        consumeOldFlows()
    }

    // set timer for calculating avg bandwidth over timewindow of 1s
    ch := make(chan bool)
	go timeout(time.Duration(1), ch)

	for {
		select {
		case _ = <-ch:
			calcAvgBandwidth()
            go timeout(time.Duration(1), ch)
		default:
			gatherMoreBytes()
		}
	}
}

func main() {
    go bandwidthCalculation()
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":2112", nil)
}
