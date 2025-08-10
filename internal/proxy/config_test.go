package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigUnMarshalling(t *testing.T) {
	t.Run("simple static handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/hello"}, "handler": {"static": {"message": "Hello there!"}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)

		route := config.Routes[0]
		assert.Equal(t, "/hello", route.Matcher.Path)
		assert.Equal(t, &StaticHandlerConfig{"Hello there!"}, route.Handler.Static)
	})

	t.Run("debug handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/debug"}, "handler": {"debug": {}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)
		route := config.Routes[0]
		assert.Equal(t, "/debug", route.Matcher.Path)
		assert.Equal(t, &DebugHandlerConfig{}, route.Handler.Debug)
	})

	t.Run("echo handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/echo"}, "handler": {"echo": {}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)
		route := config.Routes[0]
		assert.Equal(t, "/echo", route.Matcher.Path)
		assert.Equal(t, &EchoHandlerConfig{}, route.Handler.Echo)
	})

	t.Run("chaos handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/chaos"}, "handler": {"chaos": {"failure_chance": 0.5, "handler": {"static": {"message": "Hello there!"}}}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)
		route := config.Routes[0]
		assert.Equal(t, "/chaos", route.Matcher.Path)
		assert.Equal(t, &ChaosHandlerConfig{HandlerConfig{Static: &StaticHandlerConfig{"Hello there!"}}, 0.5}, route.Handler.Chaos)
	})

	t.Run("not found handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/notfound"}, "handler": {"not_found": {}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)
		route := config.Routes[0]
		assert.Equal(t, "/notfound", route.Matcher.Path)
		assert.Equal(t, &NotFoundHandlerConfig{}, route.Handler.NotFound)
	})

	t.Run("forward handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/forward"}, "handler": {"forward": {"url": "https://example.com"}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)
		route := config.Routes[0]
		assert.Equal(t, "/forward", route.Matcher.Path)
		assert.Equal(t, &ForwardHandlerConfig{
			URL: "https://example.com",
		}, route.Handler.Forward)
	})

	t.Run("retrier handler config", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/retrier"}, "handler": {"retrier": {"handler": {"static": {"message": "Hello there!"}}, "retries": 3, "retry_policy": "non_2xx_retry"}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		assert.Len(t, config.Routes, 1)
		route := config.Routes[0]
		assert.Equal(t, "/retrier", route.Matcher.Path)
		assert.Equal(t, &RetrierHandlerConfig{
			Handler:     HandlerConfig{Static: &StaticHandlerConfig{Message: "Hello there!"}},
			RetryPolicy: "non_2xx_retry",
			Retries:     3,
		}, route.Handler.Retrier)
	})
}

func TestConfig_CreateRouter(t *testing.T) {
	t.Run("create router with static handler from json", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/test"}, "handler": {"static": {"message": "test message"}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		router, err := config.CreateRouter()
		assert.NoError(t, err)

		assert.NotNil(t, router)
		assert.Len(t, router.Routes, 1)

		assert.Equal(t, router.Routes["/test"], &StaticHandler{
			message: "test message",
		})
	})

	t.Run("create router with debug handler from json", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/debug"}, "handler": {"debug": {}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		router, err := config.CreateRouter()
		assert.NoError(t, err)

		assert.NotNil(t, router)
		assert.Len(t, router.Routes, 1)

		assert.Equal(t, router.Routes["/debug"], &DebugHandler{})
	})

	t.Run("create router with echo handler from json", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/echo"}, "handler": {"echo": {}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		router, err := config.CreateRouter()
		assert.NoError(t, err)

		assert.NotNil(t, router)
		assert.Len(t, router.Routes, 1)

		assert.Equal(t, router.Routes["/echo"], &EchoHandler{})
	})

	t.Run("create router with chaos handler from json", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/chaos"}, "handler": {"chaos": {"failure_chance": 0.5, "handler": {"static": {"message": "Hello there!"}}}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		router, err := config.CreateRouter()
		assert.NoError(t, err)

		assert.NotNil(t, router)
		assert.Len(t, router.Routes, 1)

		assert.Equal(t, router.Routes["/chaos"], NewChaosHandler(
			&StaticHandler{message: "Hello there!"},
			0.5,
		))
	})

	t.Run("create router with not found handler from json", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/notfound"}, "handler": {"not_found": {}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		router, err := config.CreateRouter()
		assert.NoError(t, err)

		assert.NotNil(t, router)
		assert.Len(t, router.Routes, 1)

		assert.Equal(t, router.Routes["/notfound"], &NotFoundHandler{})
	})

	t.Run("create router with forward handler from json", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/forward"}, "handler": {"forward": {"url": "https://example.com"}}}]}`

		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		router, err := config.CreateRouter()
		assert.NoError(t, err)

		assert.NotNil(t, router)
		assert.Len(t, router.Routes, 1)

		handler, err := NewForwardHandler("https://example.com")
		assert.NoError(t, err)
		assert.Equal(t, router.Routes["/forward"], handler)
	})

	t.Run("multiple handlers in handler config should fail", func(t *testing.T) {
		// language=JSON
		configJson := `{"routes": [{"matcher": {"path": "/forward"}, "handler": {"forward": {"url": "https://example.com"}, "static": {"message": "Hello there!"}}}]}`
		config, err := ReadConfigFromString(configJson)
		assert.NoError(t, err)

		_, err = config.CreateRouter()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one handler must be set")
	})
}
