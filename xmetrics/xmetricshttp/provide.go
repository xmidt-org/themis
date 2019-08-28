package xmetricshttp

import (
	"github.com/xmidt-org/themis/xmetrics"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

func ProvideHandlerCounter(o prometheus.CounterOpts, l *ServerLabellers) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (HandlerCounter, error) {
			c, err := f.NewCounterVec(o, l.LabelNames())
			if err != nil {
				return HandlerCounter{}, err
			}

			return HandlerCounter{
				Metric:   xmetrics.LabelledCounterVec{CounterVec: c},
				Labeller: l,
			}, nil
		},
	}
}

func ProvideHandlerDurationHistogram(o prometheus.HistogramOpts, l *ServerLabellers) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (HandlerDuration, error) {
			h, err := f.NewHistogramVec(o, l.LabelNames())
			if err != nil {
				return HandlerDuration{}, err
			}

			return HandlerDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: h},
				Labeller: l,
			}, nil
		},
	}
}

func ProvideHandlerDurationSummary(o prometheus.SummaryOpts, l *ServerLabellers) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (HandlerDuration, error) {
			s, err := f.NewSummaryVec(o, l.LabelNames())
			if err != nil {
				return HandlerDuration{}, err
			}

			return HandlerDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: s},
				Labeller: l,
			}, nil
		},
	}
}

func ProvideHandlerInFlight(o prometheus.GaugeOpts) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (HandlerInFlight, error) {
			g, err := f.NewGaugeVec(o, nil)
			if err != nil {
				return HandlerInFlight{}, err
			}

			return HandlerInFlight{
				Metric: xmetrics.LabelledGaugeVec{GaugeVec: g},
			}, nil
		},
	}
}

func ProvideRoundTripperCounter(o prometheus.CounterOpts, l *ClientLabellers) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (RoundTripperCounter, error) {
			c, err := f.NewCounterVec(o, l.LabelNames())
			if err != nil {
				return RoundTripperCounter{}, err
			}

			return RoundTripperCounter{
				Metric:   xmetrics.LabelledCounterVec{CounterVec: c},
				Labeller: l,
			}, nil
		},
	}
}

func ProvideRoundTripperDurationHistogram(o prometheus.HistogramOpts, l *ClientLabellers) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (RoundTripperDuration, error) {
			h, err := f.NewHistogramVec(o, l.LabelNames())
			if err != nil {
				return RoundTripperDuration{}, err
			}

			return RoundTripperDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: h},
				Labeller: l,
			}, nil
		},
	}
}

func ProvideRoundTripperDurationSummary(o prometheus.SummaryOpts, l *ClientLabellers) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (RoundTripperDuration, error) {
			s, err := f.NewSummaryVec(o, l.LabelNames())
			if err != nil {
				return RoundTripperDuration{}, err
			}

			return RoundTripperDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: s},
				Labeller: l,
			}, nil
		},
	}
}

func ProvideRoundTripperInFlight(o prometheus.GaugeOpts) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f xmetrics.Factory) (RoundTripperInFlight, error) {
			g, err := f.NewGaugeVec(o, nil)
			if err != nil {
				return RoundTripperInFlight{}, err
			}

			return RoundTripperInFlight{
				Metric: xmetrics.LabelledGaugeVec{GaugeVec: g},
			}, nil
		},
	}
}
