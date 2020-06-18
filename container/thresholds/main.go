package main


import (
    "net/http"
    "container/list"
	"time"
    "log"
    "fmt"
    "os"
    "strconv"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/leobrada/golang_utility"

	"github.com/Shopify/sarama"
	kafka "github.com/bwNetFlow/kafkaconnector"
)

var (
    starting_point_unix int64
    training_time int
    threshold_multiplier int
    ini_path = "./settings.ini"
    kafka_conn = kafka.Connector{}
    bytes, bandwidth uint64
    max_bandwidths [24]uint64
    data_points *list.List = list.New()
    loc *time.Location
    gauge_max_bandwidth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "thresholds",
            Help: "The max bandwidth values measured in the specified network based on peered flows by the BelWü",
        },
        []string{"hour"},
    )

    gauge_max_bandwidth_0 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("0")
    gauge_max_bandwidth_1 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("1")
    gauge_max_bandwidth_2 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("2")
    gauge_max_bandwidth_3 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("3")
    gauge_max_bandwidth_4 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("4")
    gauge_max_bandwidth_5 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("5")
    gauge_max_bandwidth_6 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("6")
    gauge_max_bandwidth_7 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("7")
    gauge_max_bandwidth_8 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("8")
    gauge_max_bandwidth_9 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("9")
    gauge_max_bandwidth_10 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("10")
    gauge_max_bandwidth_11 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("11")
    gauge_max_bandwidth_12 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("12")
    gauge_max_bandwidth_13 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("13")
    gauge_max_bandwidth_14 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("14")
    gauge_max_bandwidth_15 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("15")
    gauge_max_bandwidth_16 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("16")
    gauge_max_bandwidth_17 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("17")
    gauge_max_bandwidth_18 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("18")
    gauge_max_bandwidth_19 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("19")
    gauge_max_bandwidth_20 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("20")
    gauge_max_bandwidth_21 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("21")
    gauge_max_bandwidth_22 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("22")
    gauge_max_bandwidth_23 prometheus.Gauge = gauge_max_bandwidth.WithLabelValues("23")
)

func timeout(seconds time.Duration, ch chan<- bool) {
	timer := time.NewTimer(seconds * time.Second)
	<-timer.C
	ch <- true
}

func setMaxBandwidths() {
    gauge_max_bandwidth_0.Set(float64(max_bandwidths[0]))
    gauge_max_bandwidth_1.Set(float64(max_bandwidths[1]))
    gauge_max_bandwidth_2.Set(float64(max_bandwidths[2]))
    gauge_max_bandwidth_3.Set(float64(max_bandwidths[3]))
    gauge_max_bandwidth_4.Set(float64(max_bandwidths[4]))
    gauge_max_bandwidth_5.Set(float64(max_bandwidths[5]))
    gauge_max_bandwidth_6.Set(float64(max_bandwidths[6]))
    gauge_max_bandwidth_7.Set(float64(max_bandwidths[7]))
    gauge_max_bandwidth_8.Set(float64(max_bandwidths[8]))
    gauge_max_bandwidth_9.Set(float64(max_bandwidths[9]))
    gauge_max_bandwidth_10.Set(float64(max_bandwidths[10]))
    gauge_max_bandwidth_11.Set(float64(max_bandwidths[11]))
    gauge_max_bandwidth_12.Set(float64(max_bandwidths[12]))
    gauge_max_bandwidth_13.Set(float64(max_bandwidths[13]))
    gauge_max_bandwidth_14.Set(float64(max_bandwidths[14]))
    gauge_max_bandwidth_15.Set(float64(max_bandwidths[15]))
    gauge_max_bandwidth_16.Set(float64(max_bandwidths[16]))
    gauge_max_bandwidth_17.Set(float64(max_bandwidths[17]))
    gauge_max_bandwidth_18.Set(float64(max_bandwidths[18]))
    gauge_max_bandwidth_19.Set(float64(max_bandwidths[19]))
    gauge_max_bandwidth_20.Set(float64(max_bandwidths[20]))
    gauge_max_bandwidth_21.Set(float64(max_bandwidths[21]))
    gauge_max_bandwidth_22.Set(float64(max_bandwidths[22]))
    gauge_max_bandwidth_23.Set(float64(max_bandwidths[23]))
}

func calcAvgBandwidthOverDatapoints() {
    var data_points_sum uint64
    for e := data_points.Front(); e != nil; e = e.Next() {
        data_points_sum += e.Value.(uint64)
    }
    bandwidth_over_datapoints := data_points_sum / uint64(data_points.Len())

    timestamp := time.Now().In(loc).Hour()
    new_threshold := bandwidth_over_datapoints * uint64(threshold_multiplier)
    if(new_threshold  > max_bandwidths[timestamp]) {
        max_bandwidths[timestamp] = new_threshold
    }

    setMaxBandwidths()
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

    calcAvgBandwidthOverDatapoints()
}

func gatherMoreBytes() {
	flow := <-kafka_conn.ConsumerChannel()
	bytes += flow.GetBytes()
}

func consumeOldFlows() {
    flow := <-kafka_conn.ConsumerChannel()
    _ = flow
}

func writeThresholdsToFile(path_to_file string) {
    f, err := os.OpenFile(path_to_file, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        log.Fatal(err)
    }
    for i := 0; i < len(max_bandwidths); i++ {
        line := fmt.Sprintf("%d=%d\n", i, max_bandwidths[i])
        _, err = f.WriteString(line)
        if err != nil {
            log.Fatal(err)
        }
    }
}

func idleIfThresholdsExist(path_to_file string) {
    _, err := os.Open(path_to_file)

    // Returns if the file not yet exists
    if err != nil {
        return
    }

    thresholds, _ := reader.ReadIni(path_to_file)
    for key, value := range thresholds {
        int_key, _ := strconv.Atoi(key)
        int_value, _ := strconv.Atoi(value)
        max_bandwidths[int_key] = uint64(int_value)
    }

    _, err = os.Open(path_to_file)
    for err == nil {
        setMaxBandwidths()
        time.Sleep(60)
        _, err = os.Open(path_to_file)
    }
}

func maxCalculation() {
    // Fix path in the docker container
    idleIfThresholdsExist("/data/thresholds.txt")

    // TODO: Move to init() func
	credentials, _ := reader.ReadIni(ini_path)
    loc, _ = time.LoadLocation(credentials["timezone"])
    var err error
    if training_time, err = strconv.Atoi(credentials["training_time"]); err != nil {
        training_time = 86400
    }
    if threshold_multiplier, err = strconv.Atoi(credentials["threshold_multiplier"]); err != nil {
        threshold_multiplier = 2
    }

	kafka_conn.SetAuth(credentials["user"], credentials["pwd"])
	topic := []string{credentials["topic"]}
	kafka_conn.StartConsumer(credentials["brokers"], topic, credentials["threshold_grp_id"], sarama.OffsetNewest)
	defer kafka_conn.Close()

    // set timer for calculating avg bandwidth over timewindow of 1s
    ch := make(chan bool)

    // Consume old flows
    for i := 0; i < 100000; i++ {
        consumeOldFlows()
    }

    starting_point_unix = time.Now().Unix()

    // Start calculating bandwidth
	go timeout(time.Duration(1), ch)

    thresholds_written := false
	for {
        if(time.Now().Unix() > starting_point_unix + int64(training_time)) {
            if !thresholds_written {
                // Fix path in the docker container
                writeThresholdsToFile("/data/thresholds.txt")
                thresholds_written = true
                kafka_conn.Close()
            }
            time.Sleep(60 * time.Second)
            continue
        }
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
    go maxCalculation()
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":2112", nil)
}
