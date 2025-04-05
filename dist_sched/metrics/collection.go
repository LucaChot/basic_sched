package metrics

import (

	log "github.com/sirupsen/logrus"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

type MetricCollector struct {
    cpu     linuxproc.Stat
    mem     linuxproc.MemInfo
    disk    []linuxproc.DiskStat
    net     []linuxproc.NetworkStat
}


func New() MetricCollector {
    mc := MetricCollector{}

    go mc.Collect()

    return mc
}

func (mc *MetricCollector) Collect() {
    for {
        mc.collectCPU()
        mc.collectRAM()
        mc.collectMemory()
        mc.collectNetwork()
    }
}


func (mc *MetricCollector) collectCPU() {
    stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Fatal("FAILED TO READ /proc/stat")
	}
    mc.cpu = *stat
}

func (mc *MetricCollector) collectRAM() {
	stat, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Fatal("FAILED TO READ /proc/net/dev")
	}
    mc.mem = *stat
}

/*
TODO: Look into whether I should collect ReadDisk() vs ReadDiskStat()
*/
func (mc *MetricCollector) collectMemory() {
	stat, err := linuxproc.ReadDiskStats("/proc/diskstats")
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Fatal("FAILED TO READ /proc/net/dev")
	}
    mc.disk = stat
}

func (mc *MetricCollector) collectNetwork() {
	stat, err := linuxproc.ReadNetworkStat("/proc/net/dev")
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Fatal("FAILED TO READ /proc/net/dev")
	}
    mc.net = stat
}
