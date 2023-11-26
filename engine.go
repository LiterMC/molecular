package molecular

type Config struct {
	MinSpeed float64
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
