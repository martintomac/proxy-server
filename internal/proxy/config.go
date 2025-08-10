package proxy

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Config struct {
	Routes []RouteConfig `json:"routes"`
}

func ReadConfigFromString(jsonString string) (*Config, error) {
	return ReadConfigFromBytes([]byte(jsonString))
}

func ReadConfigFromBytes(jsonBytes []byte) (*Config, error) {
	config := Config{}
	err := json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type RouteConfig struct {
	Matcher MatcherConfig `json:"matcher"`
	Handler HandlerConfig `json:"handler"`
}

type MatcherConfig struct {
	Path string `json:"path"`
}

type HandlerConfig struct {
	Static   *StaticHandlerConfig   `json:"static"`
	Forward  *ForwardHandlerConfig  `json:"forward"`
	Debug    *DebugHandlerConfig    `json:"debug"`
	Echo     *EchoHandlerConfig     `json:"echo"`
	NotFound *NotFoundHandlerConfig `json:"not_found"`
	Chaos    *ChaosHandlerConfig    `json:"chaos"`
	Fanout   *FanOutHandlerConfig   `json:"fanout"`
	Retrier  *RetrierHandlerConfig  `json:"retrier"`
}

func (h *HandlerConfig) createHandler() (Handler, error) {
	val := reflect.ValueOf(*h)

	typeToConfig := make(map[string]any, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.IsNil() {
			typeToConfig[val.Type().Field(i).Name] = field.Interface()
		}
	}

	if len(typeToConfig) == 0 {
		return nil, fmt.Errorf("no handler set")
	}
	if len(typeToConfig) > 1 {
		names := make([]string, len(typeToConfig))
		for name := range typeToConfig {
			names = append(names, name)
		}
		return nil, fmt.Errorf("exactly one handler must be set, got %v", names)
	}

	for _, config := range typeToConfig {
		return config.(interface{ createHandler() (Handler, error) }).createHandler()
	}
	return nil, fmt.Errorf("unreachable state error")
}

type StaticHandlerConfig struct {
	Message string `json:"message"`
}

func (c *StaticHandlerConfig) createHandler() (Handler, error) {
	return &StaticHandler{message: c.Message}, nil
}

type ForwardHandlerConfig struct {
	URL     string `json:"url"`
	Timeout string `json:"timeout,omitempty"` // e.g., "30s"
}

func (c *ForwardHandlerConfig) createHandler() (Handler, error) {
	return NewForwardHandler(c.URL)
}

type DebugHandlerConfig struct {
}

func (c *DebugHandlerConfig) createHandler() (Handler, error) {
	return &DebugHandler{}, nil
}

type EchoHandlerConfig struct {
}

func (c *EchoHandlerConfig) createHandler() (Handler, error) {
	return &EchoHandler{}, nil
}

type NotFoundHandlerConfig struct {
}

func (c *NotFoundHandlerConfig) createHandler() (Handler, error) {
	return &NotFoundHandler{}, nil
}

type ChaosHandlerConfig struct {
	Handler       HandlerConfig `json:"handler"`
	FailureChance float64       `json:"failure_chance"`
}

func (c *ChaosHandlerConfig) createHandler() (Handler, error) {
	wrappedHandler, err := c.Handler.createHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to create wrapped handler for chaos: %w", err)
	}
	return NewChaosHandler(wrappedHandler, c.FailureChance), nil
}

// FanOut handler config
type FanOutHandlerConfig struct {
	Handlers         []HandlerConfig `json:"handlers"`
	ResponseStrategy string          `json:"response_strategy"` // e.g., "first_successful"
}

func (c *FanOutHandlerConfig) createHandler() (Handler, error) {
	handlers := make([]Handler, len(c.Handlers))
	for i, handlerConfig := range c.Handlers {
		handler, err := handlerConfig.createHandler()
		if err != nil {
			return nil, fmt.Errorf("failed to create handler %d for fanout: %w", i, err)
		}
		handlers[i] = handler
	}

	var strategy ResponseStrategy
	switch c.ResponseStrategy {
	case "first_successful":
		strategy = &FirstSuccessfulResponseStrategy{}
	default:
		return nil, fmt.Errorf("unknown response strategy: %s", c.ResponseStrategy)
	}

	return &FanOutHandler{
		Handlers:         handlers,
		ResponseStrategy: strategy,
	}, nil
}

type RetrierHandlerConfig struct {
	Handler     HandlerConfig `json:"handler"`
	RetryPolicy string        `json:"retry_policy"`
	Retries     int           `json:"retries"`
}

func (c *RetrierHandlerConfig) createHandler() (Handler, error) {
	var retryPolicy RetryPolicy

	policy := c.RetryPolicy

	switch policy {
	case "", "non_2xx_retry":
		retryPolicy = &RetryOnNon2xxRetryPolicy{}
	default:
		return nil, fmt.Errorf("unknown retry policy: %s", policy)
	}
	handler, err := c.Handler.createHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to create handler for retrier: %w", err)
	}
	return &RetrierHandler{
		Handler:     handler,
		RetryPolicy: retryPolicy,
		Retries:     c.Retries,
	}, nil
}

// CreateRouter creates a PathRouter from the configuration
func (c *Config) CreateRouter() (*PathRouter, error) {
	router := NewPathRouter()

	for _, route := range c.Routes {
		handler, err := route.Handler.createHandler()
		if err != nil {
			return nil, fmt.Errorf("failed to create handler for route %s: %w", route.Matcher.Path, err)
		}
		router.AddRoute(route.Matcher.Path, handler)
	}

	return router, nil
}
