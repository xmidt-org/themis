package xmetricshttp

// Reporter encapsulates the notion of a metric
type Reporter interface {
	// Report uses the supplied labels/values to report a given value
	Report(lvs []string, value float64)
}

type ReporterFunc func([]string, float64)

func (rf ReporterFunc) Report(lvs []string, value float64) {
	rf(lvs, value)
}

type DiscardReporter struct{}

func (dr DiscardReporter) Report([]string, float64) {
}

// AdderReporter is a Reporter that expects positive values only
type AdderReporter interface {
	Reporter
}

// SetterReporter is a Reporter that expects values to be the current value of the underlying metric
type SetterReporter interface {
	Reporter
}

// ObserverReporter is a Reporter that expects values to be observations in a sequence
type ObserverReporter interface {
	Reporter
}
