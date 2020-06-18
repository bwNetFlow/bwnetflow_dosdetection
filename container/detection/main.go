package main

import (
    "net/http"
    "container/list"
	"time"
    "errors"
    "strconv"
    "log"
    "os"

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
    bytes, bandwidth uint64
    max_bandwidths [24]uint64
    data_points *list.List = list.New()
    loc *time.Location
    curr_bandwidth = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "detection",
    })
)

func timeout(seconds time.Duration, ch chan<- bool) {
	timer := time.NewTimer(seconds * time.Second)
	<-timer.C
	ch <- true
}

func setAlert(alert float64) {
    curr_bandwidth.Set(alert)
}

func detect_dos() {
    var data_points_sum uint64
    for e := data_points.Front(); e != nil; e = e.Next() {
        data_points_sum += e.Value.(uint64)
    }
    bandwidth_over_datapoints := data_points_sum / uint64(data_points.Len())

    timestamp := time.Now().In(loc).Hour()
    if(bandwidth_over_datapoints > (max_bandwidths[timestamp])) {
        setAlert(float64(bandwidth_over_datapoints))
    } else {
        setAlert(0)
    }
}

func calcAvgBandwidth() {
	mbit := (bytes * 8) / 1024 / 1024
	bandwidth = mbit
    bytes = 0

    if(data_points.Len() >= 60) {
        data_points.Remove(data_points.Front())
    }

    data_points.PushBack(bandwidth)

    if(data_points.Len() < 60) {
        return
    }

    detect_dos()
}

func gatherMoreBytes() {
	flow := <-kafka_conn.ConsumerChannel()
	bytes += flow.GetBytes()
}

func consumeOldFlows() {
    flow := <-kafka_conn.ConsumerChannel()
    _ = flow
}

func maxCalculation() {
	credentials, _ := reader.ReadIni(ini_path)
    loc, _ = time.LoadLocation(credentials["timezone"])

	kafka_conn.SetAuth(credentials["user"], credentials["pwd"])
	topic := []string{credentials["topic"]}
	kafka_conn.StartConsumer(credentials["brokers"], topic, credentials["detection_grp_id"], sarama.OffsetNewest)
	defer kafka_conn.Close()

    for i := 0; i < 100000; i++ {
        consumeOldFlows()
    }

    // set timer for calculating avg bandwidth over timewindow of 1s
    ch := make(chan bool)

    // Start calculating bandwidth
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

func waitForThresholds(path_to_file string) {
    var err error = errors.New("not found")
    for err != nil {
        _, err = os.Open(path_to_file)
    }
    time.Sleep(1)
    thresholds, _ := reader.ReadIni(path_to_file)
    for key, value := range thresholds {
        int_key, _ := strconv.Atoi(key)
        int_value, _ := strconv.Atoi(value)
        max_bandwidths[int_key] = uint64(int_value)
    }
    for i := 0; i < len(max_bandwidths); i++ {
        log.Printf("%d=%d\n", i, max_bandwidths[i])
    }
}

func init() {
    // Fix path in the docker container
    waitForThresholds("/data/thresholds.txt")
}

func main() {
    go maxCalculation()
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":2112", nil)
}
