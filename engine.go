package molecular

type Config struct {
	// MinSpeed means the minimum positive speed
	MinSpeed float64
	// MaxSpeed means the maximum positive speed
	MaxSpeed float64
}

type Engine struct {
	cfg Config
}

func NewEngine(cfg Config) *Engine {
	return &Engine{
		cfg: cfg,
	}
}

func (e *Engine) Config() Config {
	return e.cfg
}

func (e *Engine) Tick(dt float64) {
	// TODO
}
